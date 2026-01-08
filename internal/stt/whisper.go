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
)

type Client struct {
	apiKey       string
	mockResponse string
}

func NewClient(mockResponse string) (*Client, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	
	// If mockResponse is provided, we don't strictly need the API key
	if mockResponse == "" && os.Getenv("APP_ENV") != "dev" && apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	// Default mock for dev mode if not explicitly provided
	if mockResponse == "" && os.Getenv("APP_ENV") == "dev" {
		mockResponse = "This is a mock transcription from dev mode."
	}

	return &Client{apiKey: apiKey, mockResponse: mockResponse}, nil
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
	
	// Add language field (optional, but requested to avoid auto-detect issues)
	// Defaulting to zh-TW as per request example, but maybe configurable?
	// User said: "Default specify language (e.g. zh or zh-TW)"
	err = writer.WriteField("language", "zh") 
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

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
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
