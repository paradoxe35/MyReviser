use std::path::{Path, PathBuf};

use anyhow::{Result, anyhow};

/// Holds the loaded model between calls. Loading costs seconds; recording costs
/// milliseconds. Keeping the session resident is the largest win available to
/// dictation latency.
pub struct Engine {
    loaded: Option<Loaded>,
}

struct Loaded {
    path: PathBuf,
    session: transcribe_cpp::Session,
}

impl Engine {
    pub fn new() -> Self {
        Self { loaded: None }
    }

    pub fn unload(&mut self) {
        self.loaded = None;
    }

    /// Idempotent: reloading the resident model is free, so the host may call
    /// this on every dictation.
    pub fn load(&mut self, path: &Path) -> Result<()> {
        if self.loaded.as_ref().is_some_and(|l| l.path == path) {
            return Ok(());
        }

        let model = transcribe_cpp::Model::load_with(path, &transcribe_cpp::ModelOptions::default())
            .map_err(|e| anyhow!("failed to load {}: {e}", path.display()))?;
        let session = model
            .session_with(&transcribe_cpp::SessionOptions::default())
            .map_err(|e| anyhow!("failed to open a session: {e}"))?;

        self.loaded = Some(Loaded {
            path: path.to_path_buf(),
            session,
        });
        Ok(())
    }

    pub fn transcribe(&mut self, samples: &[f32], language: Option<&str>) -> Result<String> {
        let loaded = self
            .loaded
            .as_mut()
            .ok_or_else(|| anyhow!("no model loaded"))?;

        let options = transcribe_cpp::RunOptions {
            language: language.map(str::to_owned),
            ..Default::default()
        };

        loaded
            .session
            .run(samples, &options)
            .map(|out| out.text.trim().to_owned())
            .map_err(|e| anyhow!("transcription failed: {e}"))
    }
}

/// Reads a 16 kHz mono 16-bit WAV, the format the engine expects.
pub fn read_wav(path: &Path) -> Result<Vec<f32>> {
    let mut reader = hound::WavReader::open(path)?;
    let spec = reader.spec();

    if spec.channels != 1 || spec.sample_rate != 16_000 {
        return Err(anyhow!(
            "expected 16 kHz mono, got {} Hz and {} channels",
            spec.sample_rate,
            spec.channels
        ));
    }

    Ok(reader
        .samples::<i16>()
        .filter_map(Result::ok)
        .map(|s| s as f32 / i16::MAX as f32)
        .collect())
}
