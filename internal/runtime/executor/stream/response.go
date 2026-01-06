// Package stream provides request/response translation between API formats.
package stream

import (
	"github.com/nghyane/llm-mux/internal/config"
	"github.com/nghyane/llm-mux/internal/provider"
	"github.com/nghyane/llm-mux/internal/translator/from_ir"
	"github.com/nghyane/llm-mux/internal/translator/ir"
	"github.com/nghyane/llm-mux/internal/translator/to_ir"
)

// =============================================================================
// Response Translator
// =============================================================================

// ResponseTranslator handles unified IR-to-format conversion for non-streaming responses.
type ResponseTranslator struct {
	cfg       *config.Config
	to        string
	model     string
	messageID string
}

// NewResponseTranslator creates a translator for non-streaming responses.
func NewResponseTranslator(cfg *config.Config, to, model string) *ResponseTranslator {
	return &ResponseTranslator{
		cfg:       cfg,
		to:        to,
		model:     model,
		messageID: generateMessageID(to, model),
	}
}

func generateMessageID(to, model string) string {
	switch to {
	case "codex", "openai-response":
		return "resp-" + model
	case "claude":
		return "msg-" + model
	default:
		return "chatcmpl-" + model
	}
}

// Translate converts IR candidates to target format.
func (t *ResponseTranslator) Translate(candidates []ir.CandidateResult, usage *ir.Usage, meta *ir.OpenAIMeta) ([]byte, error) {
	if meta != nil && meta.ResponseID != "" {
		t.messageID = meta.ResponseID
	}

	// Extract messages from first candidate for formats that don't support multi-candidate
	var messages []ir.Message
	if len(candidates) > 0 {
		messages = candidates[0].Messages
	}

	switch {
	case t.to == "openai" || t.to == "cline":
		return from_ir.ToOpenAIChatCompletionCandidates(candidates, usage, t.model, t.messageID, meta)
	case t.to == "claude":
		return from_ir.ToClaudeResponse(messages, usage, t.model, t.messageID)
	case t.to == "ollama":
		return from_ir.ToOllamaChatResponse(messages, usage, t.model)
	case provider.IsGeminiFormat(t.to):
		return from_ir.ToGeminiResponseMeta(messages, usage, t.model, meta)
	case t.to == "codex" || t.to == "openai-response":
		return from_ir.ToResponsesAPIResponse(messages, usage, t.model, meta)
	default:
		return nil, nil
	}
}

// =============================================================================
// Source Format Parsers
// =============================================================================

// ParsedResponse contains parsed IR data from source format.
type ParsedResponse struct {
	Candidates []ir.CandidateResult
	Usage      *ir.Usage
	Meta       *ir.OpenAIMeta
}

// parseOpenAIResponse parses OpenAI/Codex format to IR.
func parseOpenAIResponse(response []byte) (*ParsedResponse, error) {
	messages, usage, err := to_ir.ParseOpenAIResponse(response)
	if err != nil {
		return nil, err
	}
	// Wrap in single candidate
	candidates := []ir.CandidateResult{{Index: 0, Messages: messages, FinishReason: ir.FinishReasonStop}}
	return &ParsedResponse{Candidates: candidates, Usage: usage}, nil
}

// parseClaudeResponse parses Claude format to IR.
func parseClaudeResponse(response []byte) (*ParsedResponse, error) {
	messages, usage, err := to_ir.ParseClaudeResponse(response)
	if err != nil {
		return nil, err
	}
	candidates := []ir.CandidateResult{{Index: 0, Messages: messages, FinishReason: ir.FinishReasonStop}}
	return &ParsedResponse{Candidates: candidates, Usage: usage}, nil
}

// parseGeminiResponse parses Gemini format to IR.
func parseGeminiResponse(response []byte) (*ParsedResponse, error) {
	candidates, usage, meta, err := to_ir.ParseGeminiResponseCandidates(response, nil)
	if err != nil {
		return nil, err
	}
	return &ParsedResponse{Candidates: candidates, Usage: usage, Meta: meta}, nil
}

// parseSourceResponse parses response based on source format.
func parseSourceResponse(from string, response []byte) (*ParsedResponse, error) {
	switch {
	case from == "openai" || from == "cline" || from == "codex" || from == "openai-response":
		return parseOpenAIResponse(response)
	case from == "claude":
		return parseClaudeResponse(response)
	case provider.IsGeminiFormat(from):
		return parseGeminiResponse(response)
	default:
		return nil, nil
	}
}

// =============================================================================
// Unified Entry Point
// =============================================================================

// TranslateResponseNonStream is the unified entry point for non-streaming response translation.
func TranslateResponseNonStream(cfg *config.Config, from, to provider.Format, response []byte, model string) ([]byte, error) {
	fromStr := from.String()
	toStr := to.String()

	// Handle passthrough cases
	if passthrough := handlePassthrough(fromStr, toStr, response); passthrough != nil {
		return passthrough, nil
	}

	// Parse source format to IR
	parsed, err := parseSourceResponse(fromStr, response)
	if err != nil {
		return nil, err
	}
	if parsed == nil {
		return response, nil
	}

	// Convert IR to target format
	translator := NewResponseTranslator(cfg, toStr, model)
	return translator.Translate(parsed.Candidates, parsed.Usage, parsed.Meta)
}

// handlePassthrough returns response bytes if passthrough is needed, nil otherwise.
func handlePassthrough(from, to string, response []byte) []byte {
	switch {
	case from == to:
		return response
	case (to == "codex" || to == "openai-response") && (from == "codex" || from == "openai-response"):
		return response
	case to == "claude" && from == "claude":
		return response
	}
	return nil
}
