package main

import (
	"fmt"
	"flag"
	"os/exec"
	"bytes"
	"io/ioutil"
	"net/http"
	"encoding/json"
	// pacman -S mplayer
	"github.com/hegedustibor/htgo-tts"
)

type PromptResponse struct {
	Id int
    Ai bool
	Text string
    Date string
}

func main() {

	// Small test of reading audio file and sending it to AI companion API
	// get ai-companion here		https://github.com/Hukasx0/ai-companion

	server := flag.String("a", "http://localhost", "address of the backend server")
	flag.Parse()

	args := []string{
		"-f", "alsa",
		"-i", "default",
		"-t", "5",
		"-ar", "16000",
		"-ac", "2",
		"-sample_fmt", "s16",
		"-y",
		"/tmp/out.wav",
		// records 5 second audio
		// 16bit 16khz wav stereo saved in /tmp folder, with overwriting existing file
		// you need to have ffmpeg to make this work
		// sudo pacman -S ffmpeg
	}

	cmd := exec.Command("ffmpeg", args...)
	output, err_f := cmd.CombinedOutput()
	if err_f != nil {
		fmt.Println("Error while running ffmpeg:", err_f)
		fmt.Println("Error log:")
		fmt.Println(string(output))
		return
	}

	filePath := "/tmp/out.wav"
	audioData, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error while reading audio file: ", err)
		return
	}
	resp, err := http.Post(*server+":2000"+"/api/voice", "audio/wav", bytes.NewReader(audioData))
	if err != nil {
		fmt.Println("Error while sending POST request: ", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error while sending audio file: ", resp.Status)
		return
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while getting server response: ", err)
		return
	}

	// here we should have audio transcript in responseBody if everything is correct

	promptData := map[string]string{
		"prompt": string(responseBody),
	}

	jsonPrompt, err := json.Marshal(promptData)
	if err != nil {
		fmt.Println("Error while JSON encoding: ", err)
		return
	}

	resp2, err := http.Post(*server+":3000"+"/api/prompt", "application/json", bytes.NewBuffer(jsonPrompt))
	if err != nil {
		fmt.Println("Error while sending POST request: ", err)
		return
	}
	defer resp2.Body.Close()

	responseBody2, err2 := ioutil.ReadAll(resp2.Body)
	if err2 != nil {
		fmt.Println("Error while getting server response: ", err2)
		return
	}

	var promptStruct PromptResponse

	err3 := json.Unmarshal(responseBody2, &promptStruct)

	if err3 != nil {
		fmt.Println("Error while JSON parsing: ", err3)
		return
	}

	speech := htgotts.Speech{Folder: "test", Language: "en"}
    speech.Speak(promptStruct.Text)
}
