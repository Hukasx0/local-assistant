#!/bin/bash

cd server_work/
./AI_companion &
./whisper &
echo "[~]
now you need to start tts in another terminal:
source tts/bin/activate
cd tts/
tts-server --model_name tts_models/en/ljspeech/fast_pitch &
[~]"