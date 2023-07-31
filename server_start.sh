#!/bin/bash

./server_work/AI_companion/AI_companion &
./server_work/stt/stt &
source server_work/tts/bin/activate
cd server_work/tts
tts-server --model_name tts_models/en/ljspeech/fast_pitch &
