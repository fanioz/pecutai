package gbrain

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/nex-crm/wuphf/internal/config"
)

const DefaultTimeout = 8 * time.Second

// RemoteTimeout is the deadline for a single gbrain HTTP MCP tool call when
// the caller does not provide a context with its own deadline. Remote calls
// may involve vector search + reranking, so we allow more headroom than the
// local shell-out path.
const RemoteTimeout = 30 * time.Second

var httpClient = &http.Client{}

var ErrNotInstalled = errors.New("gbrain not installed")

type SearchResult struct {
	Slug        string  `json:"slug"`
	PageID      int     `json:"page_id"`
	Title       string  `json:"title"`
	Type        string  `json:"type"`
	ChunkText   string  `json:"chunk_text"`
	ChunkSource string  `json:"chunk_source"`
	ChunkID     int     `json:"chunk_id"`
	ChunkIndex  int     `json:"chunk_index"`
	Score       float64 `json:"score"`
	Stale       bool    `json:"stale"`
}

func BinaryPath() string {
	if candidate := strings.TrimSpace(os.Getenv("WUPHF_GBRAIN_COMMAND")); candidate != "" {
		if path, err := exec.LookPath(candidate); err == nil {
			return path
		}
	}
	if path, err := exec.LookPath("gbrain"); err == nil {
		return path
	}
	return ""
}

func IsInstalled() bool {
	return BinaryPath() != ""
}

func Run(ctx context.Context, args ...string) (string, error) {
	bin := BinaryPath()
	if bin == "" {
		return "", ErrNotInstalled
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultTimeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = gbrainEnv()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("gbrain %s: timeout after %s", strings.Join(args, " "), DefaultTimeout)
		}
		if detail := strings.TrimSpace(stderr.String()); detail != "" {
			return "", fmt.Errorf("gbrain %s: %s", strings.Join(args, " "), detail)
		}
		return "", fmt.Errorf("gbrain %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func Call(ctx context.Context, tool string, params any) (string, error) {
	tool = strings.TrimSpace(tool)
	if tool == "" {
		return "", fmt.Errorf("gbrain call: tool is required")
	}
	if IsRemoteConfigured() {
		return remoteCall(ctx, tool, params)
	}
	payload, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	return Run(ctx, "call", tool, string(payload))
}

func Query(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 5
	}
	raw, err := Call(ctx, "query", map[string]any{
		"query":  query,
		"limit":  limit,
		"detail": "low",
	})
	if err != nil {
		return nil, err
	}
	// An empty content array from the remote transport yields an empty
	// string here. Treat that as "no results" to match the local binary
	// contract (which prints "[]" for no matches).
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var results []SearchResult
	if err := json.Unmarshal([]byte(raw), &results); err != nil {
		return nil, fmt.Errorf("decode gbrain query: %w", err)
	}
	return results, nil
}

func gbrainEnv() []string {
	env := os.Environ()
	// user-global; intentionally NOT under WUPHF_RUNTIME_HOME — gbrain HOME
	// fallback is only set when HOME env is unset entirely; gbrain is user-global.
	if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(os.Getenv("HOME")) == "" {
		env = append(env, "HOME="+home)
	}
	if key := strings.TrimSpace(config.ResolveOpenAIAPIKey()); key != "" && strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) == "" {
		env = append(env, "OPENAI_API_KEY="+key)
	}
	if key := strings.TrimSpace(config.ResolveAnthropicAPIKey()); key != "" && strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")) == "" {
		env = append(env, "ANTHROPIC_API_KEY="+key)
	}
	return env
}

// --- Remote HTTP MCP transport ----------------------------------------------

// IsRemoteConfigured reports whether the gbrain remote HTTP MCP endpoint is
// configured. When true, Call() routes tool invocations over HTTP instead of
// shelling out to a local gbrain binary.
func IsRemoteConfigured() bool {
	return config.ResolveGBrainURL() != "" && config.ResolveGBrainToken() != ""
}

// remoteCall invokes a gbrain MCP tool over the Streamable HTTP transport.
// The endpoint returns SSE frames; we parse the single `event: message` data
// line and extract result.content[0].text — the JSON string the local
// `gbrain call <tool> <json>` path would have printed to stdout.
func remoteCall(ctx context.Context, tool string, params any) (string, error) {
	url := config.ResolveGBrainURL()
	token := config.ResolveGBrainToken()

	args := map[string]any{
		"name":      tool,
		"arguments": params,
	}
	payload, err := json.Marshal(args)
	if err != nil {
		return "", err
	}

	body := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params":  json.RawMessage(payload),
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, RemoteTimeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("gbrain remote %s: %w", tool, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gbrain remote %s: %w", tool, err)
	}
	defer resp.Body.Close()

	// Cap the response read to guard against a misbehaving or compromised
	// remote endpoint streaming an unbounded body. 16 MiB is generous for
	// MCP tool output (the largest expected payload is a get_page result).
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return "", fmt.Errorf("gbrain remote %s: read body: %w", tool, err)
	}

	if resp.StatusCode != http.StatusOK {
		// Truncate the error body so a large or leaky response can't blow
		// up logs or agent-facing error messages.
		detail := strings.TrimSpace(string(raw))
		if len(detail) > 512 {
			detail = detail[:512] + "..."
		}
		return "", fmt.Errorf("gbrain remote %s: HTTP %d: %s", tool, resp.StatusCode, detail)
	}

	text, err := extractMCPResultText(raw)
	if err != nil {
		return "", fmt.Errorf("gbrain remote %s: %w", tool, err)
	}
	return text, nil
}

// extractMCPResultText parses an MCP Streamable HTTP SSE response and returns
// the text content of the first result content block. The wire shape is:
//
//	event: message
//	data: {"result":{"content":[{"type":"text","text":"<json>"}]},"jsonrpc":"2.0","id":1}
//
// Some servers may respond with a plain JSON body (no SSE framing) when the
// client sends Accept: application/json only; we handle both.
func extractMCPResultText(raw []byte) (string, error) {
	// Try SSE-framed first: find a `data: ` line and parse the JSON after it.
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		jsonStr := strings.TrimPrefix(line, "data: ")
		text, err := parseMCPContentText([]byte(jsonStr))
		if err != nil {
			return "", err
		}
		return text, nil
	}

	// Fall back to plain JSON (no SSE framing).
	text, err := parseMCPContentText(raw)
	if err != nil {
		return "", fmt.Errorf("parse MCP response: %w", err)
	}
	return text, nil
}

// mcpToolResponse is the JSON-RPC response envelope for a tools/call request.
type mcpToolResponse struct {
	Result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"result"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func parseMCPContentText(data []byte) (string, error) {
	var resp mcpToolResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("decode json-rpc: %w", err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
	}
	if len(resp.Result.Content) == 0 {
		return "", nil
	}
	return resp.Result.Content[0].Text, nil
}
