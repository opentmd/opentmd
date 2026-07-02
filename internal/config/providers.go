package config

// ProviderPreset describes a built-in LLM provider option for setup wizard.
type ProviderPreset struct {
	Name        string
	Label       string
	BaseURL     string
	DefaultModel string
	NeedAPIKey  bool
}

// ProviderPresets lists supported providers (OpenAI-compatible APIs).
var ProviderPresets = []ProviderPreset{
	{
		Name:         "deepseek",
		Label:        "DeepSeek",
		BaseURL:      "https://api.deepseek.com",
		DefaultModel: "deepseek-chat",
		NeedAPIKey:   true,
	},
	{
		Name:         "openai",
		Label:        "OpenAI",
		BaseURL:      "https://api.openai.com/v1",
		DefaultModel: "gpt-4o",
		NeedAPIKey:   true,
	},
	{
		Name:         "claude",
		Label:        "Claude (via OpenRouter)",
		BaseURL:      "https://openrouter.ai/api/v1",
		DefaultModel: "anthropic/claude-sonnet-4",
		NeedAPIKey:   true,
	},
	{
		Name:         "glm",
		Label:        "智谱 GLM",
		BaseURL:      "https://open.bigmodel.cn/api/paas/v4",
		DefaultModel: "glm-4-plus",
		NeedAPIKey:   true,
	},
	{
		Name:         "qwen",
		Label:        "通义 Qwen",
		BaseURL:      "https://dashscope.aliyuncs.com/compatible-mode/v1",
		DefaultModel: "qwen-max",
		NeedAPIKey:   true,
	},
	{
		Name:         "ollama",
		Label:        "Ollama (本地)",
		BaseURL:      "http://127.0.0.1:11434",
		DefaultModel: "llama3.2",
		NeedAPIKey:   false,
	},
}

func PresetByName(name string) (ProviderPreset, bool) {
	for _, p := range ProviderPresets {
		if p.Name == name {
			return p, true
		}
	}
	return ProviderPreset{}, false
}

func (c *Config) ApplyPreset(p ProviderPreset) {
	if c.Providers == nil {
		c.Providers = map[string]ProviderConfig{}
	}
	pc := c.Providers[p.Name]
	if pc.BaseURL == "" {
		pc.BaseURL = p.BaseURL
	}
	c.Providers[p.Name] = pc
	c.Default.Provider = p.Name
	if c.Default.Model == "" || c.Default.Model == DefaultModel {
		c.Default.Model = p.DefaultModel
	}
}

func defaultProvidersMap() map[string]ProviderConfig {
	m := map[string]ProviderConfig{}
	for _, p := range ProviderPresets {
		key := ""
		if !p.NeedAPIKey {
			key = "ollama"
		}
		m[p.Name] = ProviderConfig{BaseURL: p.BaseURL, APIKey: key}
	}
	return m
}
