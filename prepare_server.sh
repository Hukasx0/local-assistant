#!/bin/bash

cd server/
cargo build --release
cd ..
mkdir server_work
cp server/target/release/whisper server_work/
mkdir server_work/models_voice
cd server_work
chmod +x whisper
curl -L "https://huggingface.co/chatgpt-openai/ggml-model-whisper/resolve/main/ggml-tiny.en.bin" -o models_voice/ggml-tiny.en.bin
curl -L "https://github.com/Hukasx0/ai-companion/releases/download/0.8.5/AI_companion" -o AI_companion
chmod +x AI_companion
mkdir models
curl -L "https://huggingface.co/TheBloke/Llama-2-7B-Chat-GGML/resolve/main/llama-2-7b-chat.ggmlv3.q4_0.bin" -o models/llama-2-7b-chat.ggmlv3.q4_0.bin
python -m venv tts
source tts/bin/activate
pip install TTS
deactivate
echo "#!/bin/bash
./AI_companion &
./whisper &
echo \"[~]
now you need to start tts in another terminal:
source tts/bin/activate
cd tts/
tts-server --model_name tts_models/en/ljspeech/fast_pitch &
[~]\""> server_start.sh
chmod +x server_start.sh
cd ..
