package team

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
	"unicode"

	"github.com/nex-crm/wuphf/internal/api"
	"github.com/nex-crm/wuphf/internal/config"
	"github.com/nex-crm/wuphf/internal/nex"
)

type MemoryBackendStatus struct {
	SelectedKind  string
	SelectedLabel string
	ActiveKind    string
	ActiveLabel   string
	Detail        string
	NextStep      string
}

type memoryMCPServer struct {
	Name    string
	Command string
	Args    []string
	Env     map[string]string
	EnvVars []string
}

type memoryBackend interface {
	Kind() string
	Label() string
	Ready() bool
	MCPServer() (*memoryMCPServer, error)
	FetchBrief(ctx context.Context, notification string) string
	QueryShared(ctx context.Context, query string, limit int) ([]ScopedMemoryHit, error)
	WriteShared(ctx context.Context, note SharedMemoryWrite) (string, error)
}

type ScopedMemoryHit struct {
	Scope      string   `json:"scope,omitempty"`
	Backend    string   `json:"backend,omitempty"`
	Identifier string   `json:"identifier,omitempty"`
	Title      string   `json:"title,omitempty"`
	Snippet    string   `json:"snippet,omitempty"`
	OwnerSlug  string   `json:"owner_slug,omitempty"`
	Slug       string   `json:"slug,omitempty"`
	PageID     int      `json:"page_id,omitempty"`
	ChunkID    int      `json:"chunk_id,omitempty"`
	ChunkIndex int      `json:"chunk_index,omitempty"`
	Source     string   `json:"source,omitempty"`
	Score      *float64 `json:"score,omitempty"`
	Stale      *bool    `json:"stale,omitempty"`
}

type SharedMemoryWrite struct {
	Actor   string
	Key     string
	Title   string
	Content string
}

type noMemoryBackend struct{}

func (noMemoryBackend) Kind() string  { return config.MemoryBackendNone }
func (noMemoryBackend) Label() string { return config.MemoryBackendLabel(config.MemoryBackendNone) }
func (noMemoryBackend) Ready() bool   { return true }
func (noMemoryBackend) MCPServer() (*memoryMCPServer, error) {
	return nil, nil
}
func (noMemoryBackend) FetchBrief(context.Context, string) string { return "" }
func (noMemoryBackend) QueryShared(context.Context, string, int) ([]ScopedMemoryHit, error) {
	return nil, nil
}
func (noMemoryBackend) WriteShared(context.Context, SharedMemoryWrite) (string, error) {
	return "", fmt.Errorf("shared external memory is not active for this run")
}

type nexMemoryBackend struct{}

func (nexMemoryBackend) Kind() string  { return config.MemoryBackendNex }
func (nexMemoryBackend) Label() string { return config.MemoryBackendLabel(config.MemoryBackendNex) }
func (nexMemoryBackend) Ready() bool {
	return strings.TrimSpace(config.ResolveAPIKey("")) != "" && nexMCPBinaryPath() != ""
}
func (nexMemoryBackend) MCPServer() (*memoryMCPServer, error) {
	bin := nexMCPBinaryPath()
	if bin == "" {
		return nil, nil
	}
	apiKey := strings.TrimSpace(config.ResolveAPIKey(""))
	if apiKey == "" {
		return nil, nil
	}
	return &memoryMCPServer{
		Name:    "nex",
		Command: bin,
		Env: map[string]string{
			"WUPHF_API_KEY": apiKey,
			"NEX_API_KEY":   apiKey,
		},
		EnvVars: []string{"WUPHF_API_KEY", "NEX_API_KEY"},
	}, nil
}
func (nexMemoryBackend) FetchBrief(ctx context.Context, notification string) string {
	if !nex.Connected() {
		return ""
	}
	query := strings.TrimSpace(notification)
	if query == "" {
		return ""
	}
	if len(query) > 400 {
		query = query[:400]
	}
	answer, err := nex.Recall(ctx, query)
	if err != nil || strings.TrimSpace(answer) == "" {
		return ""
	}
	return "== NEX CONTEXT ==\n" + strings.TrimSpace(answer) + "\n== END NEX CONTEXT =="
}
func (nexMemoryBackend) QueryShared(ctx context.Context, query string, limit int) ([]ScopedMemoryHit, error) {
	client := api.NewClient(strings.TrimSpace(config.ResolveAPIKey("")))
	if !client.IsAuthenticated() {
		return nil, fmt.Errorf("nex is not configured")
	}
	type askResponse struct {
		Answer string `json:"answer"`
	}
	resp, err := api.Post[askResponse](client, "/v1/context/ask", map[string]any{
		"query": strings.TrimSpace(query),
	}, 0)
	if err != nil || strings.TrimSpace(resp.Answer) == "" {
		return nil, err
	}
	return []ScopedMemoryHit{{
		Scope:      "shared",
		Backend:    config.MemoryBackendNex,
		Identifier: "nex-context",
		Title:      "Nex context",
		Snippet:    strings.TrimSpace(resp.Answer),
		OwnerSlug:  inferSharedMemoryOwner("", strings.TrimSpace(resp.Answer)),
	}}, nil
}
func (nexMemoryBackend) WriteShared(ctx context.Context, note SharedMemoryWrite) (string, error) {
	client := api.NewClient(strings.TrimSpace(config.ResolveAPIKey("")))
	if !client.IsAuthenticated() {
		return "", fmt.Errorf("nex is not configured")
	}
	content := renderNexSharedMemoryContent(note)
	if _, err := api.Post[map[string]any](client, "/v1/context/text", map[string]any{
		"content": content,
	}, 0); err != nil {
		return "", err
	}
	key := strings.TrimSpace(note.Key)
	if key == "" {
		key = slugify(firstNonEmpty(note.Title, note.Content))
	}
	if key == "" {
		key = "shared-note"
	}
	return key, nil
}

func ResolveMemoryBackendStatus() MemoryBackendStatus {
	selected := config.ResolveMemoryBackend("")
	status := MemoryBackendStatus{
		SelectedKind:  selected,
		SelectedLabel: config.MemoryBackendLabel(selected),
		ActiveKind:    config.MemoryBackendNone,
		ActiveLabel:   config.MemoryBackendLabel(config.MemoryBackendNone),
	}

	switch selected {
	case config.MemoryBackendNone:
		status.ActiveKind = config.MemoryBackendNone
		status.ActiveLabel = config.MemoryBackendLabel(config.MemoryBackendNone)
		if config.ResolveNoNex() {
			status.Detail = "Nex is disabled for this run, so the office is operating without an external memory backend."
			status.NextStep = "Restart without --no-nex or select --memory-backend nex when you want external context."
		} else {
			status.Detail = "External memory is disabled for this run."
			status.NextStep = "Set --memory-backend nex or --memory-backend markdown to enable organizational context."
		}
	case config.MemoryBackendNex:
		if strings.TrimSpace(config.ResolveAPIKey("")) == "" {
			status.Detail = "Nex backend selected, but no WUPHF/Nex API key is configured."
			status.NextStep = "Run /init or set WUPHF_API_KEY to enable Nex-backed context."
			return status
		}
		if nexMCPBinaryPath() == "" {
			status.Detail = "Nex backend selected, but the nex-mcp server is not installed."
			status.NextStep = "Install the latest Nex CLI bundle so the Nex MCP server is available."
			return status
		}
		status.ActiveKind = config.MemoryBackendNex
		status.ActiveLabel = config.MemoryBackendLabel(config.MemoryBackendNex)
		status.Detail = "Nex-backed organizational context is configured."
	case config.MemoryBackendMarkdown:
		// Markdown is the file-over-app default: a git repo under
		// ~/.wuphf/wiki. No API keys, no external service. Always "ready"
		// as long as `git` is on PATH — and if it isn't, Repo.Init surfaces
		// ErrGitUnavailable at launch, so we don't need to guard here.
		status.ActiveKind = config.MemoryBackendMarkdown
		status.ActiveLabel = config.MemoryBackendLabel(config.MemoryBackendMarkdown)
		status.Detail = "Markdown-backed team wiki at ~/.wuphf/wiki is configured. Every edit commits to git with per-agent authorship."
	default:
		status.SelectedKind = config.MemoryBackendNone
		status.SelectedLabel = config.MemoryBackendLabel(config.MemoryBackendNone)
		status.Detail = "External memory is disabled for this run."
		status.NextStep = "Select a supported memory backend to enable external context."
	}

	return status
}

func selectedMemoryBackend() memoryBackend {
	switch config.ResolveMemoryBackend("") {
	case config.MemoryBackendNex:
		return nexMemoryBackend{}
	default:
		return noMemoryBackend{}
	}
}

func activeMemoryBackend() memoryBackend {
	backend := selectedMemoryBackend()
	if backend.Ready() {
		return backend
	}
	return noMemoryBackend{}
}

func activeMemoryBackendKind() string {
	return activeMemoryBackend().Kind()
}

func shouldPollNexNotifications() bool {
	return activeMemoryBackendKind() == config.MemoryBackendNex
}

func fetchMemoryBrief(ctx context.Context, notification string) string {
	return activeMemoryBackend().FetchBrief(ctx, notification)
}

func QuerySharedMemory(ctx context.Context, query string, limit int) ([]ScopedMemoryHit, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 5
	}
	backend := activeMemoryBackend()
	if backend.Kind() == config.MemoryBackendNone {
		return nil, nil
	}
	return backend.QueryShared(ctx, query, limit)
}

func WriteSharedMemory(ctx context.Context, note SharedMemoryWrite) (string, error) {
	note.Content = strings.TrimSpace(note.Content)
	if note.Content == "" {
		return "", fmt.Errorf("content is required")
	}
	return activeMemoryBackend().WriteShared(ctx, note)
}

func resolvedMemoryMCPServer() (*memoryMCPServer, error) {
	return activeMemoryBackend().MCPServer()
}

func nexMCPBinaryPath() string {
	path, err := exec.LookPath("nex-mcp")
	if err != nil {
		return ""
	}
	return path
}

func directMemoryPromptBlock() string {
	switch activeMemoryBackendKind() {
	case config.MemoryBackendNex:
		return "Memory scopes:\n- team_memory_query: Read your private notes (`scope=private`) or shared org memory backed by Nex (`scope=shared`)\n- team_memory_write: Store private notes by default; only write shared memory after a durable outcome is real\n- team_memory_promote: Copy one of your private notes into shared Nex memory when it becomes canonical\n- If shared memory points at another agent, ask them in the office for fresher working detail instead of guessing\n\n"
	default:
		return "Memory scopes:\n- team_memory_query: Your private notes still work with `scope=private`\n- team_memory_write: Store private notes for yourself\n- Shared org memory is not active for this run, so `scope=shared` and team_memory_promote are unavailable\n\n"
	}
}

func directMemoryStorageRule() string {
	switch activeMemoryBackendKind() {
	case config.MemoryBackendNex:
		return "7. Keep scratch notes private by default. Only claim shared storage after team_memory_write visibility=shared or team_memory_promote actually succeeded.\n"
	default:
		return "7. Do not pretend anything was stored outside your private note scope.\n"
	}
}

func leadMemoryPromptBlock() string {
	switch activeMemoryBackendKind() {
	case config.MemoryBackendNex:
		return "Memory scopes: use team_memory_query with scope=shared for org memory backed by Nex, scope=private for your own notes, and team_memory_promote when a private note becomes durable shared knowledge. If shared memory points at another agent, ask them in the office for the freshest working context.\n\n"
	default:
		return "Shared org memory is not active for this run. You can still use private notes with team_memory_query/team_memory_write scope=private.\n\n"
	}
}

func leadMemoryFirstRule() string {
	switch activeMemoryBackendKind() {
	case config.MemoryBackendNex:
		return "1. On strategy or prior decisions, call team_memory_query early. Use scope=shared for org memory and scope=private for your own retained notes.\n"
	default:
		return "1. Coordinate inside the office channel first, and use private memory only for your own scratch history.\n"
	}
}

func leadMemoryStorageRule() string {
	switch activeMemoryBackendKind() {
	case config.MemoryBackendNex:
		return "8. When you lock a durable decision, promote it into shared memory before claiming it is stored\n"
	default:
		return "8. Summarize final decisions clearly in-channel; shared org memory is unavailable in this run\n"
	}
}

func leadMemoryFinalWarning() string {
	switch activeMemoryBackendKind() {
	case config.MemoryBackendNex:
		return "Do not pretend shared memory was updated; verify team_memory_write visibility=shared or team_memory_promote succeeded.\n"
	default:
		return "Do not claim you stored anything outside your private notes.\n"
	}
}

func specialistMemoryPromptBlock() string {
	switch activeMemoryBackendKind() {
	case config.MemoryBackendNex:
		return "Memory scopes: use team_memory_query with scope=shared for org memory backed by Nex, scope=private for your own notes, and team_memory_promote when a private note becomes durable shared knowledge. If shared memory points at another agent, ask them in the office for the freshest working context.\n\n"
	default:
		return "Shared org memory is not active for this run. You can still use private notes with team_memory_query/team_memory_write scope=private.\n\n"
	}
}

func specialistMemoryStorageRule() string {
	switch activeMemoryBackendKind() {
	case config.MemoryBackendNex:
		return "9. Use team_memory_query when prior knowledge matters. Keep notes private by default, and only promote durable conclusions into shared memory once they are real.\n\n"
	default:
		return "9. Don't fake shared memory. Surface uncertainty in-channel and keep any retained notes private.\n\n"
	}
}

func renderNexSharedMemoryContent(note SharedMemoryWrite) string {
	title := strings.TrimSpace(note.Title)
	if title == "" {
		title = firstNonEmpty(strings.TrimSpace(note.Key), "WUPHF shared memory")
	}
	actor := strings.TrimSpace(note.Actor)
	if actor == "" {
		actor = "wuphf"
	}
	return fmt.Sprintf("[WUPHF shared memory]\nTitle: %s\nAuthor: @%s\nRecorded at: %s\n\n%s",
		title,
		actor,
		time.Now().UTC().Format(time.RFC3339),
		strings.TrimSpace(note.Content),
	)
}

func inferSharedMemoryOwner(identifier string, snippet string) string {
	_ = identifier
	for _, marker := range []string{"author: @", "recorded by @"} {
		if owner := parseSharedMemoryOwnerAfterMarker(snippet, marker); owner != "" {
			return owner
		}
	}
	return ""
}

func parseSharedMemoryOwnerAfterMarker(text string, marker string) string {
	text = strings.TrimSpace(text)
	marker = strings.ToLower(strings.TrimSpace(marker))
	if text == "" || marker == "" {
		return ""
	}
	lower := strings.ToLower(text)
	idx := strings.Index(lower, marker)
	if idx < 0 {
		return ""
	}
	start := idx + len(marker)
	if start >= len(text) {
		return ""
	}
	var b strings.Builder
	for _, r := range text[start:] {
		switch {
		case unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' || r == '_':
			b.WriteRune(unicode.ToLower(r))
		default:
			value := slugify(b.String())
			if value != "" {
				return value
			}
			return ""
		}
	}
	return slugify(b.String())
}
