package pricing

import "fmt"

// Per 1M tokens (USD). Best-effort estimates for /cost display.
var perMillion = map[string]struct{ In, Out float64 }{
	"deepseek":       {0.27, 1.10},
	"deepseek-chat":  {0.27, 1.10},
	"openai":         {2.50, 10.0},
	"gpt-4o":         {2.50, 10.0},
	"claude":         {3.0, 15.0},
	"glm":            {0.50, 0.50},
	"qwen":           {0.80, 2.0},
	"ollama":         {0, 0},
}

func EstimateCost(provider, model string, prompt, completion int) float64 {
	rates, ok := perMillion[model]
	if !ok {
		rates, ok = perMillion[provider]
	}
	if !ok {
		rates = perMillion["deepseek"]
	}
	return float64(prompt)*rates.In/1_000_000 + float64(completion)*rates.Out/1_000_000
}

func FormatCost(usd float64) string {
	if usd <= 0 {
		return "$0.00 (local/free)"
	}
	if usd < 0.01 {
		return fmt.Sprintf("$%.4f", usd)
	}
	return fmt.Sprintf("$%.2f", usd)
}
