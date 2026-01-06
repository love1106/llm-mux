package to_ir

import (
	"github.com/nghyane/llm-mux/internal/translator"
	"github.com/nghyane/llm-mux/internal/translator/ir"
)

type openAIParser struct{}

func (openAIParser) Parse(payload []byte) (*ir.UnifiedChatRequest, error) {
	return ParseOpenAIRequest(payload)
}

func (openAIParser) ParseResponse(payload []byte) ([]ir.Message, *ir.Usage, error) {
	return nil, nil, nil
}

func (openAIParser) ParseChunk(payload []byte) ([]ir.UnifiedEvent, error) {
	return nil, nil
}

func (openAIParser) Format() string { return "openai" }

type claudeParser struct{}

func (claudeParser) Parse(payload []byte) (*ir.UnifiedChatRequest, error) {
	return ParseClaudeRequest(payload)
}

func (claudeParser) ParseResponse(payload []byte) ([]ir.Message, *ir.Usage, error) {
	return nil, nil, nil
}

func (claudeParser) ParseChunk(payload []byte) ([]ir.UnifiedEvent, error) {
	return nil, nil
}

func (claudeParser) Format() string { return "claude" }

type geminiParser struct{}

func (geminiParser) Parse(payload []byte) (*ir.UnifiedChatRequest, error) {
	return ParseGeminiRequest(payload)
}

func (geminiParser) ParseResponse(payload []byte) ([]ir.Message, *ir.Usage, error) {
	return nil, nil, nil
}

func (geminiParser) ParseChunk(payload []byte) ([]ir.UnifiedEvent, error) {
	return nil, nil
}

func (geminiParser) Format() string { return "gemini" }

type ollamaParser struct{}

func (ollamaParser) Parse(payload []byte) (*ir.UnifiedChatRequest, error) {
	return ParseOllamaRequest(payload)
}

func (ollamaParser) ParseResponse(payload []byte) ([]ir.Message, *ir.Usage, error) {
	return nil, nil, nil
}

func (ollamaParser) ParseChunk(payload []byte) ([]ir.UnifiedEvent, error) {
	return nil, nil
}

func (ollamaParser) Format() string { return "ollama" }

func init() {
	translator.RegisterToIR("openai", openAIParser{})
	translator.RegisterToIR("cline", openAIParser{})
	translator.RegisterToIR("codex", openAIParser{})
	translator.RegisterToIR("openai-response", openAIParser{})
	translator.RegisterToIR("claude", claudeParser{})
	translator.RegisterToIR("gemini", geminiParser{})
	// Note: "gemini-cli" is not registered - it uses the same format as "gemini"
	// The difference is transport (envelope wrapping), handled by executor
	translator.RegisterToIR("ollama", ollamaParser{})
}
