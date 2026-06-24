package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Settings struct {
	ModelName      string `json:"model_name"`
	Language       string `json:"language"`
	ShowTimestamps bool   `json:"show_timestamps"`
	OutputDir      string `json:"output_dir"`
	Threads        int    `json:"threads"`
	WindowFrame    string `json:"window_frame"`
	WhisperPath    string `json:"whisper_path"`
}

func defaultWindowFrame() string {
	if runtime.GOOS == "darwin" {
		return "system"
	}
	return "custom"
}

func DefaultSettings() Settings {
	return Settings{
		ModelName:      "",
		Language:       "auto",
		ShowTimestamps: true,
		OutputDir:      "",
		Threads:        0,
		WindowFrame:    defaultWindowFrame(),
	}
}

func configPath() string {
	cfgDir := configDir()
	return filepath.Join(cfgDir, "settings.json")
}

func configDir() string {
	cd, err := os.UserConfigDir()
	if err != nil {
		cd = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(cd, "Crisper")
}

func Load() (Settings, error) {
	s := DefaultSettings()
	p := configPath()
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return s, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return s, fmt.Errorf("parse config: %w", err)
	}

	if s.Language == "en" {
		s.Language = "auto"
		Save(s)
	}

	return s, nil
}

func Save(s Settings) error {
	if err := os.MkdirAll(configDir(), 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(configPath(), data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func ModelsDir() string {
	return filepath.Join(configDir(), "models")
}

func WhisperBinDir() string {
	return filepath.Join(configDir(), "bin")
}
