package executor

import (
	"context"
	"net/http"
	"time"

	"github.com/nghyane/llm-mux/internal/config"
	cliproxyauth "github.com/nghyane/llm-mux/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/nghyane/llm-mux/sdk/cliproxy/executor"
)

type BaseExecutor struct {
	Cfg *config.Config
}

func (b *BaseExecutor) Config() *config.Config {
	return b.Cfg
}

func (b *BaseExecutor) PrepareRequest(_ *http.Request, _ *cliproxyauth.Auth) error {
	return nil
}

func (b *BaseExecutor) NewHTTPClient(ctx context.Context, auth *cliproxyauth.Auth, timeout time.Duration) *http.Client {
	return newProxyAwareHTTPClient(ctx, b.Cfg, auth, timeout)
}

func (b *BaseExecutor) NewUsageReporter(ctx context.Context, provider, model string, auth *cliproxyauth.Auth) *usageReporter {
	return newUsageReporter(ctx, provider, model, auth)
}

func (b *BaseExecutor) ApplyPayloadConfig(model string, payload []byte) []byte {
	return applyPayloadConfig(b.Cfg, model, payload)
}

func (b *BaseExecutor) RefreshNoOp(_ context.Context, auth *cliproxyauth.Auth) (*cliproxyauth.Auth, error) {
	return auth, nil
}

func (b *BaseExecutor) CountTokensNotSupported(provider string) (cliproxyexecutor.Response, error) {
	return cliproxyexecutor.Response{}, NewNotImplementedError("count tokens not supported for " + provider)
}
