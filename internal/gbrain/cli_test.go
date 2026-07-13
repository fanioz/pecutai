package gbrain

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsRemoteConfigured(t *testing.T) {
	t.Run("false when both env vars unset", func(t *testing.T) {
		t.Setenv("GBRAIN_URL", "")
		t.Setenv("GBRAIN_TOKEN", "")
		if IsRemoteConfigured() {
			t.Fatal("expected false when no env vars set")
		}
	})

	t.Run("false when only URL set", func(t *testing.T) {
		t.Setenv("GBRAIN_URL", "https://example.com/mcp")
		t.Setenv("GBRAIN_TOKEN", "")
		if IsRemoteConfigured() {
			t.Fatal("expected false when token is missing")
		}
	})

	t.Run("false when only token set", func(t *testing.T) {
		t.Setenv("GBRAIN_URL", "")
		t.Setenv("GBRAIN_TOKEN", "gbrain_xxx")
		if IsRemoteConfigured() {
			t.Fatal("expected false when URL is missing")
		}
	})

	t.Run("true when both set", func(t *testing.T) {
		t.Setenv("GBRAIN_URL", "https://example.com/mcp")
		t.Setenv("GBRAIN_TOKEN", "gbrain_xxx")
		if !IsRemoteConfigured() {
			t.Fatal("expected true when both env vars set")
		}
	})
}

func TestExtractMCPResultText_SSE(t *testing.T) {
	// gbrain serve --http returns SSE-framed JSON-RPC responses.
	raw := []byte("event: message\n" +
		`data: {"result":{"content":[{"type":"text","text":"[{\"slug\":\"test\"}]"}]},"jsonrpc":"2.0","id":1}` + "\n\n")

	text, err := extractMCPResultText(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != `[{"slug":"test"}]` {
		t.Fatalf("got %q, want [{\"slug\":\"test\"}]", text)
	}
}

func TestExtractMCPResultText_PlainJSON(t *testing.T) {
	// Some servers may respond with plain JSON (no SSE framing).
	raw := []byte(`{"result":{"content":[{"type":"text","text":"ok"}]},"jsonrpc":"2.0","id":1}`)

	text, err := extractMCPResultText(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "ok" {
		t.Fatalf("got %q, want ok", text)
	}
}

func TestExtractMCPResultText_RPCError(t *testing.T) {
	raw := []byte(`{"error":{"code":-32602,"message":"invalid params"},"jsonrpc":"2.0","id":1}`)

	_, err := extractMCPResultText(raw)
	if err == nil {
		t.Fatal("expected error for RPC error response")
	}
}

func TestExtractMCPResultText_EmptyContent(t *testing.T) {
	raw := []byte(`{"result":{"content":[]},"jsonrpc":"2.0","id":1}`)

	text, err := extractMCPResultText(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "" {
		t.Fatalf("got %q, want empty string", text)
	}
}

func TestRemoteCall_Query(t *testing.T) {
	// Simulate a gbrain HTTP MCP server responding to a tools/call for "query".
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("expected Authorization header, got %q", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json, text/event-stream" {
			t.Errorf("expected Accept header, got %q", got)
		}

		var body map[string]json.RawMessage
		_ = json.NewDecoder(r.Body).Decode(&body)

		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		_ = json.Unmarshal(body["params"], &params)

		if params.Name != "query" {
			t.Errorf("expected tool name 'query', got %q", params.Name)
		}

		results := []SearchResult{
			{Slug: "people/alice", Title: "Alice", Type: "person", ChunkText: "Alice runs engineering"},
		}
		resultJSON, _ := json.Marshal(results)
		textJSON, _ := json.Marshal(string(resultJSON))

		w.Header().Set("Content-Type", "text/event-stream")
		resp := `{"result":{"content":[{"type":"text","text":` + string(textJSON) + `}]},"jsonrpc":"2.0","id":1}`
		_, _ = w.Write([]byte("event: message\ndata: " + resp + "\n\n"))
	}))
	defer srv.Close()

	t.Setenv("GBRAIN_URL", srv.URL+"/mcp")
	t.Setenv("GBRAIN_TOKEN", "test-token")

	results, err := Query(context.Background(), "alice", 5)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Slug != "people/alice" {
		t.Fatalf("expected slug 'people/alice', got %q", results[0].Slug)
	}
}

func TestRemoteCall_PutPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		resp := `{"result":{"content":[{"type":"text","text":"{\"slug\":\"test-page\",\"ok\":true}"}]},"jsonrpc":"2.0","id":1}`
		_, _ = w.Write([]byte("event: message\ndata: " + resp + "\n\n"))
	}))
	defer srv.Close()

	t.Setenv("GBRAIN_URL", srv.URL+"/mcp")
	t.Setenv("GBRAIN_TOKEN", "test-token")

	text, err := Call(context.Background(), "put_page", map[string]any{
		"slug":    "test-page",
		"content": "# Test",
	})
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if text == "" {
		t.Fatal("expected non-empty result text")
	}
}

func TestRemoteCall_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_token"}`))
	}))
	defer srv.Close()

	t.Setenv("GBRAIN_URL", srv.URL+"/mcp")
	t.Setenv("GBRAIN_TOKEN", "bad-token")

	_, err := Call(context.Background(), "query", map[string]any{"query": "test"})
	if err == nil {
		t.Fatal("expected error for HTTP 401")
	}
}

func TestRemoteCall_RoutesWhenConfigured(t *testing.T) {
	// When GBRAIN_URL + GBRAIN_TOKEN are set, Call must route to HTTP, not
	// try to exec a local gbrain binary.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		resp := `{"result":{"content":[{"type":"text","text":"\"routed\""}]},"jsonrpc":"2.0","id":1}`
		_, _ = w.Write([]byte("event: message\ndata: " + resp + "\n\n"))
	}))
	defer srv.Close()

	t.Setenv("GBRAIN_URL", srv.URL+"/mcp")
	t.Setenv("GBRAIN_TOKEN", "test-token")

	text, err := Call(context.Background(), "get_page", map[string]any{"slug": "test"})
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if text != `"routed"` {
		t.Fatalf("expected \"routed\", got %q", text)
	}
}

func TestRemoteCall_QueryEmptyContent(t *testing.T) {
	// When the remote returns an empty content array (no text block),
	// Query should return nil, nil — not a JSON decode error.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		resp := `{"result":{"content":[]},"jsonrpc":"2.0","id":1}`
		_, _ = w.Write([]byte("event: message\ndata: " + resp + "\n\n"))
	}))
	defer srv.Close()

	t.Setenv("GBRAIN_URL", srv.URL+"/mcp")
	t.Setenv("GBRAIN_TOKEN", "test-token")

	results, err := Query(context.Background(), "nothing-matches", 5)
	if err != nil {
		t.Fatalf("expected nil error for empty content, got: %v", err)
	}
	if results != nil {
		t.Fatalf("expected nil results, got %d items", len(results))
	}
}
