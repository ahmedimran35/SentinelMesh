package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	LLM      LLMConfig
	Server   ServerConfig
	DB       DBConfig
	Monitor  MonitorConfig
	RateLimit int
}

type LLMConfig struct {
	Provider string // "nim" or "ollama"
	NIM      NIMConfig
	Ollama   OllamaConfig
}

type NIMConfig struct {
	APIKey    string
	Model     string
	MaxTokens int
	Endpoint  string
}

type OllamaConfig struct {
	URL   string
	Model string
}

type ServerConfig struct {
	Port string
	Host string
}

type DBConfig struct {
	Path string
}

type MonitorConfig struct {
	MaxConcurrent  int
	DefaultInterval time.Duration
}

func Load() *Config {
	return &Config{
		LLM: LLMConfig{
			Provider: getEnv("LLM_PROVIDER", "nim"),
			NIM: NIMConfig{
				APIKey:    getEnv("NIM_API_KEY", ""),
				Model:     getEnv("NIM_MODEL", "meta/llama-3.1-70b-instruct"),
				MaxTokens: getEnvInt("NIM_MAX_TOKENS", 4096),
				Endpoint:  getEnv("NIM_ENDPOINT", "https://integrate.api.nvidia.com/v1/chat/completions"),
			},
			Ollama: OllamaConfig{
				URL:   getEnv("OLLAMA_URL", "http://localhost:11434"),
				Model: getEnv("OLLAMA_MODEL", "llama3.1"),
			},
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "8090"),
			Host: getEnv("HOST", "0.0.0.0"),
		},
		DB: DBConfig{
			Path: getEnv("DB_PATH", "./sentinelmesh.db"),
		},
		Monitor: MonitorConfig{
			MaxConcurrent:   getEnvInt("MAX_CONCURRENT_SCANS", 5),
			DefaultInterval: getEnvDuration("DEFAULT_SCAN_INTERVAL", 24*time.Hour),
		},
		RateLimit: getEnvInt("API_RATE_LIMIT", 10),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
