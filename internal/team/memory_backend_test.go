package team

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nex-crm/wuphf/internal/config"
)

func TestResolveMemoryBackendStatusNoNexFallsBackToMarkdown(t *testing.T) {
	// --no-nex with no explicit backend pick lands the user on the markdown
	// wiki — the default external memory that doesn't require Nex. Silent
	// 'none' was a worse default: the user asked to skip Nex, not to lose
	// all shared memory.
	t.Setenv("WUPHF_NO_NEX", "1")
	t.Setenv("WUPHF_MEMORY_BACKEND", "")

	status := ResolveMemoryBackendStatus()
	if status.SelectedKind != config.MemoryBackendMarkdown {
		t.Fatalf("expected selected backend markdown, got %+v", status)
	}
	if status.ActiveKind != config.MemoryBackendMarkdown {
		t.Fatalf("expected active backend markdown, got %+v", status)
	}
}

func TestShouldPollNexNotificationsOnlyWhenNexIsActive(t *testing.T) {
	t.Setenv("WUPHF_NO_NEX", "")
	t.Setenv("WUPHF_API_KEY", "nex-test-key")
	t.Setenv("WUPHF_MEMORY_BACKEND", config.MemoryBackendNex)

	binDir := t.TempDir()
	nexMCP := filepath.Join(binDir, "nex-mcp")
	if err := os.WriteFile(nexMCP, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("create fake nex-mcp: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	if !shouldPollNexNotifications() {
		t.Fatal("expected nex notification polling when nex backend is active")
	}
}
