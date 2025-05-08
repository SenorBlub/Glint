package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const groqAPIURL = "https://api.groq.com/openai/v1/chat/completions"

// ViewImage sends a base64-encoded image to Groq's vision API and returns the extracted text.
func ViewImage(origin, name, base64Image string) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY environment variable is not set")
	}

	// Construct the image data URL
	imageDataURL := fmt.Sprintf("data:image/jpeg;base64,%s", base64Image)

	// Prepare the request payload
	payload := map[string]interface{}{
		"model": "llama-3.2-90b-vision-preview",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "Describe the contents of this image.",
					},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": imageDataURL,
						},
					},
				},
			},
		},
		"temperature":           0.7,
		"max_completion_tokens": 1024,
		"top_p":                 1,
		"stream":                false,
	}

	// Marshal the payload to JSON
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", groqAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error string `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for errors in the response
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("image-to-text failed: %s", result.Error)
	}

	if len(result.Choices) == 0 || result.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("no content returned from Groq API")
	}

	return result.Choices[0].Message.Content, nil
}
