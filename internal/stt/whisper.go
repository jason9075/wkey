package stt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Client struct {
	apiKey       string
	language     string
	mockResponse string
	verbose      bool
}

func NewClient(apiKey string, language string, mockResponse string, verbose bool) (*Client, error) {
	// If mockResponse is provided, we don't strictly need the API key
	if mockResponse == "" && os.Getenv("APP_ENV") != "dev" && apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is missing. Please set OPENAI_API_KEY env var or configure it in ~/.config/wkey/config.json")
	}

	return &Client{apiKey: apiKey, language: language, mockResponse: mockResponse, verbose: verbose}, nil
}

type transcriptionResponse struct {
	Text string `json:"text"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) Transcribe(filename string) (string, error) {
	if c.mockResponse != "" {
		return c.mockResponse, nil
	}

	if c.verbose {
		keyLen := len(c.apiKey)
		maskedKey := "missing"
		if keyLen > 8 {
			maskedKey = c.apiKey[:4] + "..." + c.apiKey[keyLen-4:]
		}
		fmt.Printf("Transcribing %s (Language: %s, API Key: %s)\n", filename, c.language, maskedKey)
	}

	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file field
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	// Add model field
	err = writer.WriteField("model", "whisper-1")
	if err != nil {
		return "", fmt.Errorf("failed to write model field: %w", err)
	}
	
	// Add language field
	err = writer.WriteField("language", c.language) 
	if err != nil {
		return "", fmt.Errorf("failed to write language field: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if c.verbose || resp.StatusCode != http.StatusOK {
		fmt.Printf("API Status: %d\n", resp.StatusCode)
		if len(respBody) > 0 {
			fmt.Printf("API Response: %s\n", string(respBody))
		}
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return "", fmt.Errorf("Invalid API Key")
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			return "", fmt.Errorf("Rate Limit Exceeded")
		}
		return "", fmt.Errorf("API Error: %d", resp.StatusCode)
	}

	var result transcriptionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API returned error: %s", result.Error.Message)
	}

	return result.Text, nil
}
