package api

import (
	"encoding/json"
	"os"
)

// Config represents the application configuration
type Config struct {
	Database DatabaseConfig `json:"database"`
	Gemini   GeminiConfig   `json:"gemini"`
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Driver          string `json:"driver"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	DBName          string `json:"dbname"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"`
}

// GeminiConfig holds Gemini API settings
type GeminiConfig struct {
	APIKey    string `json:"api_key"`
	ModelName string `json:"model_name"`
}

// LoadConfig loads the application configuration from a JSON file
func LoadConfig() (Config, error) {
	var config Config
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(configFile, &config)
	return config, err
}