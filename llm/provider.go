package llm

import (
	"sentinelmesh/config"
)

// Provider is the interface for LLM completions
type Provider interface {
	ChatCompletion(systemPrompt, userPrompt string) (string, error)
	StreamCompletion(systemPrompt, userPrompt string, ch chan<- string) error
}

// NewProvider creates the appropriate LLM provider based on config
func NewProvider(cfg *config.LLMConfig) Provider {
	switch cfg.Provider {
	case "nim":
		if cfg.NIM.APIKey != "" {
			return NewNIMProvider(&cfg.NIM)
		}
		// fallback to ollama if no NIM key
		return NewOllamaProvider(&cfg.Ollama)
	case "ollama":
		return NewOllamaProvider(&cfg.Ollama)
	default:
		if cfg.NIM.APIKey != "" {
			return NewNIMProvider(&cfg.NIM)
		}
		return NewOllamaProvider(&cfg.Ollama)
	}
}
