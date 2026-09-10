use std::ffi::c_char;
use std::os::raw::{c_float, c_int};
use std::path::{Path, PathBuf};
use std::sync::mpsc::channel;
use std::thread;

use parking_lot::Mutex;

use crate::ffi::ffi_types::{
    FFIErrorCode, SttHandle, c_str_to_string, set_last_error, string_to_c_str,
};
use crate::stt::audio::{self, Recorder};
use crate::stt::engine::Engine;

/// Reports microphone level while recording, so the host can draw a meter.
pub type LevelCallback = extern "C" fn(c_float);

pub struct SpeechRecogniser {
    engine: Mutex<Engine>,
    recorder: Recorder,
    recording: Mutex<bool>,
}

impl SpeechRecogniser {
    fn new(level: LevelCallback) -> Self {
        let (tx, rx) = channel();

        // Levels arrive far faster than a UI can use them; the host is called
        // on this thread, never on the audio callback.
        thread::spawn(move || {
            for rms in rx {
                level(rms);
            }
        });

        Self {
            engine: Mutex::new(Engine::new()),
            recorder: Recorder::spawn(tx),
            recording: Mutex::new(false),
        }
    }
}

fn recogniser<'a>(handle: SttHandle) -> Option<&'a SpeechRecogniser> {
    if handle.is_null() {
        set_last_error("Null speech handle provided".to_string());
        return None;
    }
    Some(unsafe { &*(handle as *mut SpeechRecogniser) })
}

#[unsafe(no_mangle)]
pub extern "C" fn encre_stt_new(level: LevelCallback) -> SttHandle {
    Box::into_raw(Box::new(SpeechRecogniser::new(level))) as SttHandle
}

#[unsafe(no_mangle)]
pub unsafe extern "C" fn encre_stt_free(handle: SttHandle) {
    if handle.is_null() {
        return;
    }
    let recogniser = unsafe { Box::from_raw(handle as *mut SpeechRecogniser) };
    recogniser.recorder.shutdown();
}

/// Loads a model and keeps it resident. Idempotent for the same path, so the
/// host may call it on every dictation.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn encre_stt_load(handle: SttHandle, path: *const c_char) -> c_int {
    let Some(recogniser) = recogniser(handle) else {
        return FFIErrorCode::NullPointer as c_int;
    };
    if path.is_null() {
        set_last_error("Null model path provided".to_string());
        return FFIErrorCode::NullPointer as c_int;
    }

    let path = match unsafe { c_str_to_string(path) } {
        Ok(path) => path,
        Err(e) => {
            set_last_error(format!("Invalid model path: {e}"));
            return FFIErrorCode::InvalidUtf8 as c_int;
        }
    };

    match recogniser.recorder.load(PathBuf::from(&path)) {
        Ok(_) => FFIErrorCode::Success as c_int,
        Err(e) => {
            set_last_error(e.to_string());
            FFIErrorCode::OperationFailed as c_int
        }
    }
}

#[unsafe(no_mangle)]
pub unsafe extern "C" fn encre_stt_unload(handle: SttHandle) -> c_int {
    let Some(recogniser) = recogniser(handle) else {
        return FFIErrorCode::NullPointer as c_int;
    };
    recogniser.recorder.unload();
    FFIErrorCode::Success as c_int
}

#[unsafe(no_mangle)]
pub unsafe extern "C" fn encre_stt_start(handle: SttHandle) -> c_int {
    let Some(recogniser) = recogniser(handle) else {
        return FFIErrorCode::NullPointer as c_int;
    };

    let mut recording = recogniser.recording.lock();
    if *recording {
        set_last_error("Already recording".to_string());
        return FFIErrorCode::OperationFailed as c_int;
    }

    recogniser.recorder.start();
    *recording = true;
    FFIErrorCode::Success as c_int
}

/// Stops recording and transcribes. Blocks for as long as inference takes, so
/// the host must call it off its UI thread.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn encre_stt_stop(handle: SttHandle) -> *mut c_char {
    let Some(recogniser) = recogniser(handle) else {
        return std::ptr::null_mut();
    };

    {
        let mut recording = recogniser.recording.lock();
        if !*recording {
            set_last_error("Not recording".to_string());
            return std::ptr::null_mut();
        }
        *recording = false;
    }

    let stopped = match recogniser.recorder.stop() {
        Ok(stopped) => stopped,
        Err(e) => {
            set_last_error(e.to_string());
            return std::ptr::null_mut();
        }
    };

    if let Some(text) = stopped.text {
        return string_to_c_str(text);
    }

    // Silence is not a failure: the user pressed and released without speaking.
    if stopped.samples.is_empty() {
        return string_to_c_str(String::new());
    }

    match recogniser.engine.lock().transcribe(&stopped.samples, None) {
        Ok(text) => string_to_c_str(text),
        Err(e) => {
            set_last_error(e.to_string());
            std::ptr::null_mut()
        }
    }
}

#[unsafe(no_mangle)]
pub unsafe extern "C" fn encre_stt_cancel(handle: SttHandle) -> c_int {
    let Some(recogniser) = recogniser(handle) else {
        return FFIErrorCode::NullPointer as c_int;
    };

    *recogniser.recording.lock() = false;
    recogniser.recorder.cancel();
    FFIErrorCode::Success as c_int
}

/// Transcribes a 16 kHz mono WAV without touching the microphone, so a model
/// can be verified from settings.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn encre_stt_transcribe_file(
    handle: SttHandle,
    path: *const c_char,
) -> *mut c_char {
    let Some(recogniser) = recogniser(handle) else {
        return std::ptr::null_mut();
    };
    if path.is_null() {
        set_last_error("Null audio path provided".to_string());
        return std::ptr::null_mut();
    }

    let path = match unsafe { c_str_to_string(path) } {
        Ok(path) => path,
        Err(e) => {
            set_last_error(format!("Invalid audio path: {e}"));
            return std::ptr::null_mut();
        }
    };

    let samples = match crate::stt::engine::read_wav(Path::new(&path)) {
        Ok(samples) => samples,
        Err(e) => {
            set_last_error(e.to_string());
            return std::ptr::null_mut();
        }
    };

    match recogniser.engine.lock().transcribe(&samples, None) {
        Ok(text) => string_to_c_str(text),
        Err(e) => {
            set_last_error(e.to_string());
            std::ptr::null_mut()
        }
    }
}

/// Selects the capture device by name. Null or empty means the system default.
/// Takes effect on the next recording.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn encre_stt_set_device(handle: SttHandle, name: *const c_char) -> c_int {
    let Some(recogniser) = recogniser(handle) else {
        return FFIErrorCode::NullPointer as c_int;
    };

    let wanted = if name.is_null() {
        None
    } else {
        match unsafe { c_str_to_string(name) } {
            Ok(name) if !name.is_empty() => Some(name),
            Ok(_) => None,
            Err(e) => {
                set_last_error(format!("Invalid device name: {e}"));
                return FFIErrorCode::InvalidUtf8 as c_int;
            }
        }
    };

    recogniser.recorder.set_device(wanted);
    FFIErrorCode::Success as c_int
}

/// Input device names, newline separated, the default marked with a leading '*'.
#[unsafe(no_mangle)]
pub extern "C" fn encre_stt_devices() -> *mut c_char {
    let (devices, default) = audio::devices();

    let listed = devices
        .into_iter()
        .map(|name| {
            if Some(&name) == default.as_ref() {
                format!("*{name}")
            } else {
                name
            }
        })
        .collect::<Vec<_>>()
        .join("\n");

    string_to_c_str(listed)
}
