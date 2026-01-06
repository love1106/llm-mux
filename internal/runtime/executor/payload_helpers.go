package executor

import (
	"github.com/nghyane/llm-mux/internal/config"
	"github.com/nghyane/llm-mux/internal/sseutil"
)

// ApplyPayloadConfig delegates to sseutil for shared implementation.
func ApplyPayloadConfig(cfg *config.Config, model string, payload []byte) []byte {
	return sseutil.ApplyPayloadConfig(cfg, model, payload)
}

// ApplyPayloadConfigWithRoot delegates to sseutil for shared implementation.
func ApplyPayloadConfigWithRoot(cfg *config.Config, model, protocol, root string, payload []byte) []byte {
	return sseutil.ApplyPayloadConfigWithRoot(cfg, model, protocol, root, payload)
}
