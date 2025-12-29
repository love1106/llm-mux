package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	clineauth "github.com/nghyane/llm-mux/internal/auth/cline"
	"github.com/nghyane/llm-mux/internal/config"
	cliproxyauth "github.com/nghyane/llm-mux/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/nghyane/llm-mux/sdk/cliproxy/executor"
	sdktranslator "github.com/nghyane/llm-mux/sdk/translator"
	log "github.com/sirupsen/logrus"
)

type ClineExecutor struct {
	cfg *config.Config
}

func NewClineExecutor(cfg *config.Config) *ClineExecutor {
	return &ClineExecutor{cfg: cfg}
}

func (e *ClineExecutor) Identifier() string {
	return "cline"
}

func (e *ClineExecutor) PrepareRequest(_ *http.Request, _ *cliproxyauth.Auth) error {
	return nil
}

func (e *ClineExecutor) Execute(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (resp cliproxyexecutor.Response, err error) {
	token, baseURL := clineCredentials(auth)
	if token == "" {
		return resp, fmt.Errorf("cline access token not available")
	}

	if baseURL == "" {
		baseURL = ClineDefaultBaseURL
	}

	reporter := newUsageReporter(ctx, e.Identifier(), req.Model, auth)
	defer reporter.trackFailure(ctx, &err)

	from := opts.SourceFormat
	body, err := TranslateToOpenAI(e.cfg, from, req.Model, req.Payload, false, nil)
	if err != nil {
		return resp, err
	}
	body = applyPayloadConfig(e.cfg, req.Model, body)

	url := strings.TrimSuffix(baseURL, "/") + "/api/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return resp, err
	}

	e.applyClineHeaders(httpReq, token, false)

	httpClient := newProxyAwareHTTPClient(ctx, e.cfg, auth, 0)
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return resp, NewTimeoutError("request timed out")
		}
		return resp, err
	}
	defer func() {
		if errClose := httpResp.Body.Close(); errClose != nil {
			log.Errorf("cline executor: close response body error: %v", errClose)
		}
	}()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		result := HandleHTTPError(httpResp, "cline executor")
		return resp, result.Error
	}

	data, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return resp, err
	}

	reporter.publish(ctx, extractUsageFromOpenAIResponse(data))

	fromOpenAI := sdktranslator.FromString("openai")
	translatedResp, err := TranslateResponseNonStream(e.cfg, fromOpenAI, from, data, req.Model)
	if err != nil {
		return resp, err
	}
	if translatedResp != nil {
		resp = cliproxyexecutor.Response{Payload: translatedResp}
	} else {
		resp = cliproxyexecutor.Response{Payload: data}
	}
	return resp, nil
}

func (e *ClineExecutor) ExecuteStream(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (stream <-chan cliproxyexecutor.StreamChunk, err error) {
	token, baseURL := clineCredentials(auth)
	if token == "" {
		return nil, fmt.Errorf("cline access token not available")
	}

	if baseURL == "" {
		baseURL = ClineDefaultBaseURL
	}

	reporter := newUsageReporter(ctx, e.Identifier(), req.Model, auth)
	defer reporter.trackFailure(ctx, &err)

	from := opts.SourceFormat
	body, err := TranslateToOpenAI(e.cfg, from, req.Model, req.Payload, true, nil)
	if err != nil {
		return nil, err
	}
	body = applyPayloadConfig(e.cfg, req.Model, body)

	url := strings.TrimSuffix(baseURL, "/") + "/api/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	e.applyClineHeaders(httpReq, token, true)

	httpClient := newProxyAwareHTTPClient(ctx, e.cfg, auth, 0)
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, NewTimeoutError("request timed out")
		}
		return nil, err
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		result := HandleHTTPError(httpResp, "cline executor")
		_ = httpResp.Body.Close()
		return nil, result.Error
	}

	messageID := "chatcmpl-" + req.Model
	processor := NewOpenAIStreamProcessor(e.cfg, from, req.Model, messageID)
	processor.Preprocess = clinePreprocess

	return RunSSEStream(ctx, httpResp.Body, reporter, processor, StreamConfig{
		ExecutorName:       "cline executor",
		Preprocessor:       ClineDataTagPreprocessor(),
		SkipDoneInData:     true,
		PassthroughOnEmpty: true,
	}), nil
}

func ClineDataTagPreprocessor() StreamPreprocessor {
	return func(line []byte) (payload []byte, skip bool) {
		payload = line
		if bytes.HasPrefix(line, []byte("data: ")) {
			payload = bytes.TrimSpace(line[6:])
		} else if bytes.HasPrefix(line, []byte("data:")) {
			payload = bytes.TrimSpace(line[5:])
		}

		if len(payload) == 0 || bytes.Equal(payload, []byte("[DONE]")) {
			return nil, true
		}
		return payload, false
	}
}

func clinePreprocess(line []byte, firstChunk bool) []byte {
	payload := convertClineReasoningToOpenAI(line, firstChunk)

	if !firstChunk && shouldSkipEmptyContentChunk(payload) {
		return nil
	}

	return append([]byte("data: "), payload...)
}

func clineCredentials(a *cliproxyauth.Auth) (token, baseURL string) {
	return ExtractCreds(a, ClineCredsConfig)
}

func (e *ClineExecutor) applyClineHeaders(req *http.Request, token string, stream bool) {
	ApplyAPIHeaders(req, HeaderConfig{
		Token: token,
		StreamHeaders: map[string]string{
			"Cache-Control": "no-cache",
			"Connection":    "keep-alive",
		},
	}, stream)
}

func convertClineReasoningToOpenAI(payload []byte, isFirstChunk bool) []byte {
	if bytes.Contains(payload, []byte(`"reasoning":`)) && !bytes.Contains(payload, []byte(`"reasoning_content":`)) {
		payload = bytes.ReplaceAll(payload, []byte(`"reasoning":`), []byte(`"reasoning_content":`))
	}

	if !isFirstChunk && bytes.Contains(payload, []byte(`"role":"assistant"`)) {
		payload = bytes.ReplaceAll(payload, []byte(`"role":"assistant",`), []byte{})
		payload = bytes.ReplaceAll(payload, []byte(`,"role":"assistant"`), []byte{})
	}

	return payload
}

func shouldSkipEmptyContentChunk(payload []byte) bool {
	hasEmptyContent := bytes.Contains(payload, []byte(`"content":""`)) ||
		bytes.Contains(payload, []byte(`"content":null`))

	if !hasEmptyContent {
		return false
	}

	if bytes.Contains(payload, []byte(`"tool_calls"`)) {
		return false
	}

	if bytes.Contains(payload, []byte(`"finish_reason"`)) {
		return false
	}

	if bytes.Contains(payload, []byte(`"usage"`)) {
		return false
	}

	return true
}

func (e *ClineExecutor) CountTokens(ctx context.Context, auth *cliproxyauth.Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	return CountTokensForOpenAIProvider(ctx, e.cfg, "cline executor", opts.SourceFormat, req.Model, req.Payload, nil)
}

func (e *ClineExecutor) Refresh(ctx context.Context, auth *cliproxyauth.Auth) (*cliproxyauth.Auth, error) {
	if auth == nil {
		return nil, fmt.Errorf("cline executor: auth is nil")
	}

	refreshToken, ok := ExtractRefreshToken(auth)
	if !ok {
		return auth, nil
	}

	svc := clineauth.NewClineAuth(e.cfg)
	td, err := svc.RefreshTokens(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	UpdateRefreshMetadata(auth, map[string]any{
		"access_token":  td.AccessToken,
		"refresh_token": td.RefreshToken,
		"email":         td.Email,
		"expired":       td.Expire,
	}, "cline")

	return auth, nil
}
