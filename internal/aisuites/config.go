package aisuites

// AIConfig holds the provider-agnostic AI settings.
type AIConfig struct {
	Provider string // "openai", "anthropic", "ollama"
	Model    string // e.g. "gpt-4o-mini", "claude-3-haiku-20240307", "llama3"
	APIKey   string
	BaseURL  string // optional: for OpenAI-compatible endpoints (Groq, Together, etc.)
}
