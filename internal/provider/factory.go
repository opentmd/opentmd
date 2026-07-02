package provider

import (
	"fmt"
	"strings"

	"github.com/opentmd/opentmd-cli/internal/cliname"
	"github.com/opentmd/opentmd-cli/internal/config"
	"github.com/opentmd/opentmd-cli/internal/deepseek"
	"github.com/opentmd/opentmd-cli/internal/llm"
)

func New(cfg *config.Config) (llm.Provider, error) {
	name, pc, err := cfg.ActiveProvider()
	if err != nil {
		return nil, err
	}
	if pc.APIKey == "" && name != "ollama" {
		return nil, fmt.Errorf("API key for provider %q is not set; run: %s login", name, cliname.Current())
	}
	if name == "ollama" && pc.APIKey == "" {
		pc.APIKey = "ollama"
	}

	baseURL := pc.BaseURL
	if name == "ollama" && !strings.Contains(baseURL, "/v1") {
		// Ollama OpenAI-compatible endpoint
		baseURL = strings.TrimRight(baseURL, "/") + "/v1"
	}

	return deepseek.New(baseURL, pc.APIKey), nil
}
