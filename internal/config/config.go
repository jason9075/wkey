package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type VisualConfig struct {
	BarCount       int     `json:"bar_count"`
	BarColorStart  string  `json:"bar_color_start"`
	BarColorEnd    string  `json:"bar_color_end"`
	AnimationSpeed float64 `json:"animation_speed"`
}

type Config struct {
	OpenAIAPIKey string       `json:"openai_api_key"`
	Language     string       `json:"language"`
	Visual       VisualConfig `json:"visual"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Visual: VisualConfig{
			BarCount:       32,
			BarColorStart:  "#00FFFF", // Cyan
			BarColorEnd:    "#8A2BE2", // BlueViolet
			AnimationSpeed: 1.0,
		},
	}

	// Check Config File
	configPath, err := getConfigPath()
	if err != nil {
		return cfg, nil
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return cfg, nil
	}

	file, err := os.Open(configPath)
	if err != nil {
		return cfg, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Ensure defaults for visual if not provided in JSON
	if cfg.OpenAIAPIKey == "" {
		cfg.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	}
	if cfg.Language == "" {
		cfg.Language = "zh"
	}
	if cfg.Visual.BarCount == 0 {
		cfg.Visual.BarCount = 32
	}
	if cfg.Visual.BarColorStart == "" {
		cfg.Visual.BarColorStart = "#00FFFF"
	}
	if cfg.Visual.BarColorEnd == "" {
		cfg.Visual.BarColorEnd = "#8A2BE2"
	}
	if cfg.Visual.AnimationSpeed == 0 {
		cfg.Visual.AnimationSpeed = 1.0
	}

	return cfg, nil
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "wkey", "config.json"), nil
}
