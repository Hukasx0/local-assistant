#!/bin/bash

cd server/
cargo build --release
cd ..
mkdir server_work
mkdir server_work/stt
cp server/target/release/server server_work/stt/stt
mkdir server_work/stt/models
cd server_work
chmod +x stt/stt
curl -L "https://huggingface.co/chatgpt-openai/ggml-model-whisper/resolve/main/ggml-tiny.en.bin" -o stt/models/ggml-tiny.en.bin
curl -L "https://github.com/Hukasx0/ai-companion/releases/download/0.8.5/download_linux.sh" -o download_companion.sh
chmod +x download_companion.sh
./download_companion.sh
chmod +x AI_companion/AI_companion
python -m venv tts
source tts/bin/activate
pip install TTS
deactivate
echo "#!/bin/bash
./AI_companion/AI_companion &
./stt/stt &
source tts/bin/activate
cd tts
tts-server --model_name tts_models/en/ljspeech/fast_pitch &" > server_start.sh
chmod +x server_start.sh
cd ..
