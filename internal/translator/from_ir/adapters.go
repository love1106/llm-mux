package from_ir

import (
	"github.com/nghyane/llm-mux/internal/translator"
	"github.com/nghyane/llm-mux/internal/translator/ir"
)

type geminiConverter struct{}

func (geminiConverter) ConvertRequest(req *ir.UnifiedChatRequest) ([]byte, error) {
	return (&GeminiProvider{}).ConvertRequest(req)
}

func (geminiConverter) ToResponse(messages []ir.Message, usage *ir.Usage, model string) ([]byte, error) {
	return ToGeminiResponse(messages, usage, model)
}

func (geminiConverter) ToChunk(event ir.UnifiedEvent, model string) ([]byte, error) {
	return ToGeminiChunk(event, model)
}

func (geminiConverter) Provider() string { return "gemini" }

type claudeConverter struct{}

func (claudeConverter) ConvertRequest(req *ir.UnifiedChatRequest) ([]byte, error) {
	return (&ClaudeProvider{}).ConvertRequest(req)
}

func (claudeConverter) ToResponse(messages []ir.Message, usage *ir.Usage, model string) ([]byte, error) {
	return ToClaudeResponse(messages, usage, model, "")
}

func (claudeConverter) ToChunk(event ir.UnifiedEvent, model string) ([]byte, error) {
	return ToClaudeSSE(event, nil)
}

func (claudeConverter) Provider() string { return "claude" }

type openaiConverter struct{}

func (openaiConverter) ConvertRequest(req *ir.UnifiedChatRequest) ([]byte, error) {
	return ToOpenAIRequest(req)
}

func (openaiConverter) ToResponse(messages []ir.Message, usage *ir.Usage, model string) ([]byte, error) {
	return ToOpenAIChatCompletion(messages, usage, model, "")
}

func (openaiConverter) ToChunk(event ir.UnifiedEvent, model string) ([]byte, error) {
	return ToOpenAIChunk(event, model, "", 0)
}

func (openaiConverter) Provider() string { return "openai" }

type ollamaConverter struct{}

func (ollamaConverter) ConvertRequest(req *ir.UnifiedChatRequest) ([]byte, error) {
	return ToOllamaRequest(req)
}

func (ollamaConverter) ToResponse(messages []ir.Message, usage *ir.Usage, model string) ([]byte, error) {
	return ToOllamaChatResponse(messages, usage, model)
}

func (ollamaConverter) ToChunk(event ir.UnifiedEvent, model string) ([]byte, error) {
	return ToOllamaChatChunk(event, model)
}

func (ollamaConverter) Provider() string { return "ollama" }

type kiroConverter struct{}

func (kiroConverter) ConvertRequest(req *ir.UnifiedChatRequest) ([]byte, error) {
	return (&KiroProvider{}).ConvertRequest(req)
}

func (kiroConverter) ToResponse(messages []ir.Message, usage *ir.Usage, model string) ([]byte, error) {
	return ToOpenAIChatCompletion(messages, usage, model, "")
}

func (kiroConverter) ToChunk(event ir.UnifiedEvent, model string) ([]byte, error) {
	return ToOpenAIChunk(event, model, "", 0)
}

func (kiroConverter) Provider() string { return "kiro" }

func init() {
	translator.RegisterFromIR("gemini", geminiConverter{})
	translator.RegisterFromIR("claude", claudeConverter{})
	translator.RegisterFromIR("openai", openaiConverter{})
	translator.RegisterFromIR("ollama", ollamaConverter{})
	translator.RegisterFromIR("kiro", kiroConverter{})
}
