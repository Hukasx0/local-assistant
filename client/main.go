package main

import (
	"fmt"
	"flag"
	"os"
	"os/exec"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"strings"
	"regexp"
	"net/url"
	"path/filepath"
	"github.com/agnivade/levenshtein"
	// pacman -S mplayer
)

type PromptResponse struct {
	Id int
    Ai bool
	Text string
    Date string
}

func get_voice_transcript(server *string, seconds string, filePath string) string {
	args := []string{
		"-f", "alsa",
		"-i", "default",
		"-t", seconds,
		"-ar", "16000",
		"-ac", "2",
		"-sample_fmt", "s16",
		"-y",
		filePath,
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
		return ""
	}

	audioData, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error while reading audio file: ", err)
		return ""
	}
	resp, err := http.Post(*server+":2000"+"/api/voice", "audio/wav", bytes.NewReader(audioData))
	if err != nil {
		fmt.Println("Error while sending POST request: ", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error while sending audio file: ", resp.Status)
		return ""
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while getting server response: ", err)
		return ""
	}

	return string(responseBody)
}

func tts(server *string, text string) {
	text = url.QueryEscape(strings.ReplaceAll(text, ">", ""))
	resp, err := http.Get(*server+":5002/api/tts?text="+text)
	if err != nil {
		fmt.Println("Error while doing GET request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Wrong status code, error: ", resp.Status)
		return
	}
	file, err := os.Create("/tmp/play.wav")
	if err != nil {
		fmt.Println("Error while creating a file: ", err)
		return
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println("Error while writing to file", err)
		return
	}
	cmd := exec.Command("xdg-open", "/tmp/play.wav")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error while playing audio file: ", err)
		return
	}
}

func get_most_similiar(search string, data []string) string {
	if len(data) == 0 {
		return ""
	}

	minDistance := levenshtein.ComputeDistance(search, data[0])

	mostSimilarElement := data[0]

	for _, s := range data[1:] {
		distance := levenshtein.ComputeDistance(search, s)
		if distance < minDistance {
			minDistance = distance
			mostSimilarElement = s
		}
	}

	return mostSimilarElement
}

func play_music_file(music_file_name string) {
	files, err := ioutil.ReadDir("music")
	if err != nil {
		fmt.Println("Error while reading music folder: ", err)
		return
	}
	var music_files []string
	for _, file := range files {
		name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		extension := filepath.Ext(file.Name())
		music_files = append(music_files, name+extension)
	}
	music_file := get_most_similiar(music_file_name, music_files)
	cmd := exec.Command("xdg-open", "music/"+music_file)
	err_p := cmd.Run()
	if err_p != nil {
		fmt.Println("Error while playing audio file: ", err_p)
		return
	}
}

func main() {
	// Small test of reading audio file and sending it to AI companion API
	// get ai-companion here		https://github.com/Hukasx0/ai-companion

	server := flag.String("a", "http://localhost", "address of the backend server")
	flag.Parse()

	for {
				// listen 24/7 for command "hey companion!"
		call := strings.ToLower(get_voice_transcript(server, "5", "/tmp/hey.wav"))
		if strings.Contains(call, "hey companion") {
					// if user said command then listen for commands for ai-companion
			cmd := exec.Command("xdg-open", "sounds/listening.wav")
			err_p := cmd.Run()
			if err_p != nil {
				fmt.Println("Error while playing audio file: ", err_p)
				return
			}
			voice_message := get_voice_transcript(server, "8", "/tmp/out.wav")
			reg := regexp.MustCompile(`\[.*?\]`)
			voice_message = reg.ReplaceAllString(voice_message, "")
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(voice_message)), "play ") {
				song_name := strings.TrimSpace(voice_message[len("play "):])
				play_music_file(song_name)
			} else {
				promptData := map[string]string{
					"prompt": voice_message,
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
				tts(server, promptStruct.Text)
			}
		}
	}
}
