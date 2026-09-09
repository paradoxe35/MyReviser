use std::path::Path;

#[cfg(feature = "gguf")]
fn gguf(path: &Path) -> transcribe_cpp::Result<String> {
    let model = transcribe_cpp::Model::load_with(path, &transcribe_cpp::ModelOptions::default())?;
    let mut session = model.session_with(&transcribe_cpp::SessionOptions::default())?;
    Ok(session.run(&[0.0f32; 16000], &transcribe_cpp::RunOptions::default())?.text)
}

#[cfg(feature = "onnx")]
fn onnx(dir: &str) -> String {
    let mut config = sherpa_onnx::OfflineRecognizerConfig::default();
    config.model_config.sense_voice.model = Some(format!("{dir}/model.int8.onnx"));
    config.model_config.tokens = Some(format!("{dir}/tokens.txt"));

    let Some(recognizer) = sherpa_onnx::OfflineRecognizer::create(&config) else {
        return "recognizer failed to load".into();
    };

    let stream = recognizer.create_stream();
    stream.accept_waveform(16000, &[0.0f32; 16000]);
    recognizer.decode(&stream);

    stream.get_result().map(|r| r.text).unwrap_or_default()
}

fn main() {
    let args: Vec<String> = std::env::args().collect();
    let _ = Path::new("");

    #[cfg(feature = "gguf")]
    {
        println!("gguf engine: transcribe-cpp {}", transcribe_cpp::version());
        if args.get(1).map(String::as_str) == Some("gguf") {
            println!("{:?}", gguf(Path::new(&args[2])));
        }
    }

    #[cfg(feature = "onnx")]
    {
        println!("onnx engine: sherpa-onnx");
        if args.get(1).map(String::as_str) == Some("onnx") {
            println!("{}", onnx(&args[2]));
        }
    }
}
