package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

type Response struct {
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
	VoiceURL string `json:"voice_url,omitempty"`
}

func getTextToSpeech(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	text := r.FormValue("text")
	text = trimWhitespace(text)

	if text == "" {
		response := Response{Status: "error", Message: "Text is required."}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check audio cache
	audioFilePath := getVoiceFromCache(text)
	if audioFilePath != "" {
		response := Response{Status: "success", VoiceURL: audioFilePath}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Set Google Application Credentials
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "./credentials/service_account_credentials.json")

	// Create a context
	ctx := context.Background()

	// Create a client
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// Perform the Text-to-Speech request
	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "th-TH",
			Name:         "th-TH-Standard-A",
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save the audio content to a file
	audioFilePath = filepath.Join("files/com_voice/", fmt.Sprintf("%x.mp3", md5Hash(text)))
	os.MkdirAll(filepath.Dir(audioFilePath), 0777)
	if err := os.WriteFile(audioFilePath, resp.AudioContent, 0644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Response
	response := Response{Status: "success", Message: "Audio content generated successfully.", VoiceURL: audioFilePath}
	json.NewEncoder(w).Encode(response)
}

func trimWhitespace(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\n", ""))
}

func md5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func getVoiceFromCache(text string) string {
	if text == "" {
		return ""
	}
	audioFilePath := filepath.Join("files/com_voice/", fmt.Sprintf("%x.mp3", md5Hash(text)))
	if _, err := os.Stat(audioFilePath); err == nil {
		return audioFilePath
	}
	return ""
}

func main() {
	// Handle API endpoint
	http.HandleFunc("/text_to_speech/api_v1", getTextToSpeech)

	// Handle route for 'static' directory
	const staticDir = "./files/com_voice"
	const routePrefix = "/text_to_speech/files/com_voice/"
	fs := http.FileServer(http.Dir(staticDir))
	http.Handle(routePrefix, http.StripPrefix(routePrefix, fs))

	// Start server
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
