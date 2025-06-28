package groq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const apiURL = "https://api.groq.com/openai/v1/chat/completions"

var (
	geminiURL = "https://api.gemini.com/v1/"
)
var (
	currentModel    = "llama3-70b-8192"
	availableModels = map[string]string{
		"llama3-70b-8192":                       "Llama 3 70B",
		"llama3-8b-8192":                        "Llama 3 8B",
		"mixtral-8x7b-32768":                    "Mixtral 8x7B",
		"gemma-7b-it":                           "Gemma 7B",
		"llama3-groq-70b-8192-tool-use-preview": "Llama 3 70B Tools",
		"llama3-groq-8b-8192-tool-use-preview":  "Llama 3 8B Tools",
	}
)

func GetAvailableModels() map[string]string {
	return availableModels
}

func GetCurrentModel() string {
	return currentModel
}

func SetModel(model string) error {
	if _, exists := availableModels[model]; !exists {
		return fmt.Errorf("model %s not available", model)
	}
	currentModel = model
	return nil
}

func SendPrompt(prompt string) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("API variable not set")
	}

	payload := Request{
		Model: currentModel,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var parsed Response
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}

	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("no response recieved")
	}

	return parsed.Choices[0].Message.Content, nil
}
