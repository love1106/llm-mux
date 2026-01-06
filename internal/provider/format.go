package provider

// Format represents the API data format for request/response translation.
// It is distinct from the Provider ID (which handles transport/auth).
type Format string

const (
	FormatUnknown     Format = ""
	FormatOpenAI      Format = "openai"
	FormatClaude      Format = "claude"
	FormatGemini      Format = "gemini"
	FormatOllama      Format = "ollama"
	FormatCodex       Format = "codex"
	FormatAntigravity Format = "antigravity"
)

// FromString converts a string identifier to a typed Format.
// Used primarily by the Registry to map string keys to parsers.
func FromString(v string) Format {
	return Format(v)
}

func (f Format) String() string {
	return string(f)
}

// IsGeminiFormat checks if a format string represents Gemini format.
// This handles legacy "gemini-cli" which uses the same data format as "gemini".
func IsGeminiFormat(format string) bool {
	return format == "gemini" || format == "gemini-cli"
}
