use std::collections::HashSet;
use std::sync::mpsc::{Receiver, RecvTimeoutError, Sender, TryRecvError, channel};
use std::thread;
use std::time::Duration;

use anyhow::{Result, anyhow};
use cpal::traits::{DeviceTrait, HostTrait, StreamTrait};
use cpal::{Device, SampleFormat, StreamConfig};

pub const SAMPLE_RATE: u32 = 16_000;

/// earshot wants exactly 256 samples (16 ms) at 16 kHz.
const VAD_FRAME: usize = 256;
const VAD_THRESHOLD: f32 = 0.5;

/// Speech is reported for this long after the detector stops seeing it, so a
/// trailing word is not clipped mid-syllable.
const HANGOVER_FRAMES: usize = 28; // ~450 ms
/// Frames kept before onset, recovering the attack the detector needed to fire.
const PREFILL_FRAMES: usize = 28;
/// Consecutive speech frames before onset is believed, rejecting clicks.
const ONSET_FRAMES: usize = 4;

const RESAMPLER_CHUNK: usize = 1024;

/// How often the recorder drains the callback queue while recording.
const DRAIN_INTERVAL: Duration = Duration::from_millis(20);

pub enum Command {
    /// None means the system default. Takes effect on the next recording, so a
    /// change mid-take cannot truncate what is being said.
    SetDevice(Option<String>),
    /// Open the stream without recording so a later Start costs nothing.
    Warm,
    Start,
    Stop(Sender<Vec<f32>>),
    Cancel,
    Shutdown,
}

/// Owns the capture stream on its own thread. cpal delivers audio on a realtime
/// callback that must not block, so it only forwards buffers; every conversion
/// happens here.
pub struct Recorder {
    commands: Sender<Command>,
}

impl Recorder {
    pub fn spawn(levels: Sender<f32>) -> Self {
        let (tx, rx) = channel();
        thread::spawn(move || run(rx, levels));
        Self { commands: tx }
    }

    pub fn warm(&self) {
        let _ = self.commands.send(Command::Warm);
    }

    pub fn set_device(&self, name: Option<String>) {
        let _ = self.commands.send(Command::SetDevice(name));
    }

    pub fn start(&self) {
        let _ = self.commands.send(Command::Start);
    }

    pub fn cancel(&self) {
        let _ = self.commands.send(Command::Cancel);
    }

    pub fn shutdown(&self) {
        let _ = self.commands.send(Command::Shutdown);
    }

    /// Stops capture and returns the speech that was gathered, at 16 kHz mono.
    pub fn stop(&self) -> Result<Vec<f32>> {
        let (tx, rx) = channel();
        self.commands
            .send(Command::Stop(tx))
            .map_err(|_| anyhow!("recorder thread is gone"))?;
        rx.recv().map_err(|_| anyhow!("recorder dropped the reply"))
    }
}

fn run(commands: Receiver<Command>, levels: Sender<f32>) {
    let mut stream: Option<StreamGuard> = None;
    let mut recording = false;
    let mut pipeline = Pipeline::new();
    let mut preferred: Option<String> = None;

    loop {
        // While recording, wake often enough to keep converting as audio
        // arrives. Blocking until the next command would defer every resample
        // and VAD frame to Stop, putting that work on the critical path.
        let command = if recording {
            match commands.recv_timeout(DRAIN_INTERVAL) {
                Ok(command) => Some(command),
                Err(RecvTimeoutError::Timeout) => None,
                Err(RecvTimeoutError::Disconnected) => return,
            }
        } else {
            match commands.recv() {
                Ok(command) => Some(command),
                Err(_) => return,
            }
        };

        if recording {
            if let Some(guard) = stream.as_ref() {
                pipeline.feed(&guard.take());
            }
        }

        let Some(command) = command else { continue };

        match command {
            Command::SetDevice(name) => {
                if preferred != name {
                    preferred = name;
                    // Drop the open stream so the next warm or start reopens on
                    // the newly chosen device.
                    stream = None;
                }
            }
            Command::Warm => {
                if stream.is_none() {
                    stream = StreamGuard::open(levels.clone(), preferred.as_deref()).ok();
                }
            }
            Command::Start => {
                if stream.is_none() {
                    stream = StreamGuard::open(levels.clone(), preferred.as_deref()).ok();
                }
                pipeline.reset(stream.as_ref().map(|s| s.rate).unwrap_or(SAMPLE_RATE));
                recording = true;
            }
            Command::Stop(reply) => {
                recording = false;
                let samples = drain(&mut stream, &mut pipeline);
                let _ = reply.send(samples);
            }
            Command::Cancel => {
                recording = false;
                drain(&mut stream, &mut pipeline);
            }
            Command::Shutdown => return,
        }
    }
}

/// Pulls whatever the callback has queued, converts it, and returns the speech.
fn drain(stream: &mut Option<StreamGuard>, pipeline: &mut Pipeline) -> Vec<f32> {
    if let Some(guard) = stream.as_ref() {
        pipeline.feed(&guard.take());
    }
    pipeline.finish()
}

struct StreamGuard {
    _stream: cpal::Stream,
    rate: u32,
    channels: usize,
    incoming: Receiver<Vec<f32>>,
}

impl StreamGuard {
    fn open(levels: Sender<f32>, preferred: Option<&str>) -> Result<Self> {
        let device = open_device(preferred)?;
        let config = preferred_config(&device)?;

        let rate = config.config.sample_rate;
        let channels = config.config.channels as usize;
        let (tx, rx) = channel();

        let stream = build_stream(&device, &config, tx, levels)?;
        // cpal 0.18 no longer starts a stream on creation. Without this the
        // callback never fires and every recording comes back silent.
        stream.play()?;

        Ok(Self {
            _stream: stream,
            rate,
            channels,
            incoming: rx,
        })
    }

    fn take(&self) -> Vec<f32> {
        let mut out = Vec::new();
        loop {
            match self.incoming.try_recv() {
                Ok(chunk) => out.extend(mono(&chunk, self.channels)),
                Err(TryRecvError::Empty) | Err(TryRecvError::Disconnected) => return out,
            }
        }
    }
}

fn mono(interleaved: &[f32], channels: usize) -> Vec<f32> {
    if channels <= 1 {
        return interleaved.to_vec();
    }
    interleaved
        .chunks(channels)
        .map(|frame| frame.iter().sum::<f32>() / channels as f32)
        .collect()
}

struct SelectedConfig {
    config: StreamConfig,
    format: SampleFormat,
}

/// Falls back to the default when the chosen device is gone: a microphone that
/// was unplugged should not stop dictation from working at all.
fn open_device(preferred: Option<&str>) -> Result<Device> {
    let host = host();

    if let Some(wanted) = preferred {
        match host.input_devices() {
            Ok(devices) => {
                if let Some(device) = devices.filter(|d| d.to_string() == wanted).next() {
                    return Ok(device);
                }
                tracing::warn!("Input device '{wanted}' is unavailable, using the default");
            }
            Err(e) => tracing::warn!("Could not enumerate input devices: {e}"),
        }
    }

    host.default_input_device()
        .ok_or_else(|| anyhow!("no input device available"))
}

/// Input device names, the default marked with a leading '*'.
pub fn devices() -> (Vec<String>, Option<String>) {
    let host = host();
    let names = host
        .input_devices()
        .map(|devices| {
            let mut seen = HashSet::new();
            devices
                .map(|d| d.to_string())
                .filter(|name| !name.is_empty() && seen.insert(name.clone()))
                .collect()
        })
        .unwrap_or_default();
    let default = host.default_input_device().map(|d| d.to_string());
    (names, default)
}

fn host() -> cpal::Host {
    // ALSA rather than cpal's default on Linux: PulseAudio and PipeWire both
    // expose an ALSA interface, and going direct avoids a resampling hop.
    #[cfg(target_os = "linux")]
    {
        cpal::host_from_id(cpal::HostId::Alsa).unwrap_or_else(|_| cpal::default_host())
    }
    #[cfg(not(target_os = "linux"))]
    {
        cpal::default_host()
    }
}

/// Takes the device's own rate rather than demanding 16 kHz. Forcing a rate the
/// hardware does not want is how Bluetooth headsets end up in headset profile,
/// or ALSA refuses the stream outright.
fn preferred_config(device: &Device) -> Result<SelectedConfig> {
    let default = device.default_input_config()?;
    let rate = default.sample_rate();

    let best = device
        .supported_input_configs()?
        .filter(|range| range.min_sample_rate() <= rate && rate <= range.max_sample_rate())
        .max_by_key(|range| match range.sample_format() {
            SampleFormat::F32 => 3,
            SampleFormat::I16 => 2,
            SampleFormat::I32 => 1,
            _ => 0,
        });

    match best {
        Some(range) => Ok(SelectedConfig {
            format: range.sample_format(),
            config: range.with_sample_rate(rate).config(),
        }),
        None => Ok(SelectedConfig {
            format: default.sample_format(),
            config: default.config(),
        }),
    }
}

fn build_stream(
    device: &Device,
    selected: &SelectedConfig,
    samples: Sender<Vec<f32>>,
    levels: Sender<f32>,
) -> Result<cpal::Stream> {
    let error = |e| eprintln!("audio stream error: {e}");

    let stream = match selected.format {
        SampleFormat::F32 => device.build_input_stream(
            selected.config.clone(),
            move |data: &[f32], _: &_| forward(data.to_vec(), &samples, &levels),
            error,
            None,
        )?,
        SampleFormat::I16 => device.build_input_stream(
            selected.config.clone(),
            move |data: &[i16], _: &_| {
                let converted = data.iter().map(|s| *s as f32 / i16::MAX as f32).collect();
                forward(converted, &samples, &levels)
            },
            error,
            None,
        )?,
        SampleFormat::I32 => device.build_input_stream(
            selected.config.clone(),
            move |data: &[i32], _: &_| {
                let converted = data.iter().map(|s| *s as f32 / i32::MAX as f32).collect();
                forward(converted, &samples, &levels)
            },
            error,
            None,
        )?,
        other => return Err(anyhow!("unsupported sample format {other:?}")),
    };

    Ok(stream)
}

/// Runs on the realtime audio callback: send and return, never allocate slowly
/// or block, or the driver drops buffers.
fn forward(data: Vec<f32>, samples: &Sender<Vec<f32>>, levels: &Sender<f32>) {
    if !data.is_empty() {
        let sum: f32 = data.iter().map(|s| s * s).sum();
        let _ = levels.send((sum / data.len() as f32).sqrt());
    }
    let _ = samples.send(data);
}

/// Resamples to 16 kHz, then keeps only the frames the detector calls speech.
struct Pipeline {
    resampler: Option<rubato::FftFixedIn<f32>>,
    pending: Vec<f32>,
    frame: Vec<f32>,
    detector: earshot::Detector,
    speech: Vec<f32>,
    prefill: std::collections::VecDeque<Vec<f32>>,
    onset: usize,
    hangover: usize,
}

impl Pipeline {
    fn new() -> Self {
        Self {
            resampler: None,
            pending: Vec::new(),
            frame: Vec::with_capacity(VAD_FRAME),
            detector: earshot::Detector::default(),
            speech: Vec::new(),
            prefill: std::collections::VecDeque::with_capacity(PREFILL_FRAMES),
            onset: 0,
            hangover: 0,
        }
    }

    fn reset(&mut self, input_rate: u32) {
        // A resampler carries FFT overlap between calls; reusing one across
        // recordings leaks the tail of the previous take into the next.
        self.resampler = (input_rate != SAMPLE_RATE)
            .then(|| {
                rubato::FftFixedIn::<f32>::new(
                    input_rate as usize,
                    SAMPLE_RATE as usize,
                    RESAMPLER_CHUNK,
                    1,
                    1,
                )
                .ok()
            })
            .flatten();

        self.pending.clear();
        self.frame.clear();
        self.speech.clear();
        self.prefill.clear();
        self.detector = earshot::Detector::default();
        self.onset = 0;
        self.hangover = 0;
    }

    fn feed(&mut self, samples: &[f32]) {
        if samples.is_empty() {
            return;
        }

        let resampled = self.resample(samples);
        for sample in resampled {
            self.frame.push(sample);
            if self.frame.len() == VAD_FRAME {
                let frame = std::mem::replace(&mut self.frame, Vec::with_capacity(VAD_FRAME));
                self.classify(frame);
            }
        }
    }

    fn resample(&mut self, samples: &[f32]) -> Vec<f32> {
        let Some(resampler) = self.resampler.as_mut() else {
            return samples.to_vec();
        };

        use rubato::Resampler;
        self.pending.extend_from_slice(samples);

        let mut out = Vec::new();
        while self.pending.len() >= RESAMPLER_CHUNK {
            let chunk: Vec<f32> = self.pending.drain(..RESAMPLER_CHUNK).collect();
            if let Ok(mut done) = resampler.process(&[chunk], None) {
                out.append(&mut done[0]);
            }
        }
        out
    }

    fn classify(&mut self, frame: Vec<f32>) {
        let speaking = self.detector.predict_f32(&frame) >= VAD_THRESHOLD;

        if speaking {
            self.onset += 1;
        } else {
            self.onset = 0;
        }

        if self.onset >= ONSET_FRAMES {
            // Onset confirmed: replay the buffered attack, then stay open for
            // the hangover so the tail of the utterance is not cut.
            self.speech.extend(self.prefill.drain(..).flatten());
            self.hangover = HANGOVER_FRAMES;
        }

        if self.hangover > 0 {
            self.hangover -= 1;
            self.speech.extend_from_slice(&frame);
            return;
        }

        if self.prefill.len() == PREFILL_FRAMES {
            self.prefill.pop_front();
        }
        self.prefill.push_back(frame);
    }

    /// Flushes the resampler's delay line and returns the recording.
    fn finish(&mut self) -> Vec<f32> {
        if !self.pending.is_empty() {
            let tail: Vec<f32> = std::mem::take(&mut self.pending);
            let mut padded = tail;
            padded.resize(RESAMPLER_CHUNK, 0.0);
            let flushed = self.resample(&padded);
            for sample in flushed {
                self.frame.push(sample);
                if self.frame.len() == VAD_FRAME {
                    let frame = std::mem::replace(&mut self.frame, Vec::with_capacity(VAD_FRAME));
                    self.classify(frame);
                }
            }
        }

        if self.hangover > 0 && !self.frame.is_empty() {
            self.speech.extend_from_slice(&self.frame);
        }
        self.frame.clear();

        std::mem::take(&mut self.speech)
    }
}
