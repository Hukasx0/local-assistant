use whisper_rs::{WhisperContext, FullParams, SamplingStrategy};
use futures_util::StreamExt as _;
use actix_web::{get, post, web, App, HttpResponse, HttpServer};
use std::fs;

#[post("/api/voice")]
async fn voice_transcript(mut received: actix_web::web::Payload) -> HttpResponse {

    // curl -X POST -H "Content-Type: audio/wav" -T audio.wav http://localhost:3000/api/voice

    let mut wav_data = web::BytesMut::new();
    while let Some(chunk) = received.next().await {
        let data = chunk.unwrap();
        wav_data.extend_from_slice(&data);
    }
    let mut reader = hound::WavReader::new(&wav_data[..]);
    match reader {
        Ok(_) => {}
        Err(e) => {
            return HttpResponse::BadRequest().body(
                format!("Only .wav sample rate 16KHz is supported, {}", e));
        }
    }
    let mut model_path: String = String::from("");
    let dir_path = "models/";
    if let Ok(entries) = fs::read_dir(dir_path) {
        for entry in entries {
            if let Ok(entry) = entry {
                if let Some(file_name) = entry.file_name().to_str() {
                    if file_name.ends_with(".bin") {
                        model_path = "models_voice/".to_owned()+file_name;
                        println!("loaded model models/{}", file_name);
                    }
                }
            }
        }
    }
    if model_path == "" {
        eprintln!("You need to put your AI model (with .bin format - ggml) in models/ folder");
        panic!();
    }
    let mut ctx = WhisperContext::new(&model_path).unwrap();
    let mut state = ctx.create_state().expect("failed to create key");
    let mut params = FullParams::new(SamplingStrategy::Greedy { best_of: 0 });
    #[allow(unused_variables)]
    let hound::WavSpec {
        channels,
        sample_rate,
        bits_per_sample,
        ..
    } = reader.as_ref().unwrap().spec();
    let mut audio = whisper_rs::convert_integer_to_float_audio(
        &reader
            .unwrap()
            .samples::<i16>()
            .map(|s| s.expect("invalid sample"))
            .collect::<Vec<_>>(),
    );
    if channels == 2 {
        audio = whisper_rs::convert_stereo_to_mono_audio(&audio).unwrap();
    } else if channels != 1 {
        panic!(">2 channels unsupported");
    }

    if sample_rate != 16000 {
        panic!("sample rate must be 16KHz");
    }
    state.full(params, &audio[..]).expect("failed to run model");
    let num_segments = state
    .full_n_segments()
    .expect("failed to get number of segments");
    let mut transcript = String::new();
    for i in 0..num_segments {
        let segment = state
            .full_get_segment_text(i)
            .expect("failed to get segment");
        transcript += &segment;
    }
    HttpResponse::Ok().body(format!("{}", &transcript))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {

    let port: u16 = 2000;
    let hostname: &str = "0.0.0.0";

    let payload_cfg = web::PayloadConfig::default().limit(5_000_000);

    println!("local assistant backend works at:\n -> http://{}:{}/", hostname, port);
    println!("You can access it here:\n -> http://localhost:{}/", port);
    HttpServer::new(move || {
        App::new()
            .app_data(payload_cfg.clone())
            .service(voice_transcript)
    })
    .bind((hostname, port))?
    .run()
    .await
}
