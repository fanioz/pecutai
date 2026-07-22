package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/nex-crm/wuphf/internal/agent"
)

type agyHelperRecord struct {
	Args  []string `json:"args"`
	Stdin string   `json:"stdin"`
}

func TestBuildAgyArgsPassesPromptAndModel(t *testing.T) {
	args := buildAgyArgs("gemini-2.5-pro", "ship the thing")
	if len(args) == 0 || args[0] != "--print" {
		t.Fatalf("expected `--print` as first arg, got %v", args)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--model gemini-2.5-pro") {
		t.Fatalf("expected explicit model flag, got %q", joined)
	}
	if args[len(args)-1] != "ship the thing" {
		t.Fatalf("expected prompt as final arg, got %q", args[len(args)-1])
	}
}

func TestBuildAgyArgsOmitsModelWhenUnset(t *testing.T) {
	args := buildAgyArgs("", "do the thing")
	for _, a := range args {
		if a == "--model" {
			t.Fatalf("did not expect --model flag when model empty, got %v", args)
		}
	}
}

func TestBuildAgyArgsOmitsPromptWhenEmpty(t *testing.T) {
	args := buildAgyArgs("gemini-2.5-pro", "")
	if len(args) != 3 {
		t.Fatalf("expected [--print --model X], got %v", args)
	}
}

func TestBuildAgyPromptWrapsSystem(t *testing.T) {
	got := buildAgyPrompt("sys instructions", "do the thing")
	if !strings.Contains(got, "<system>") {
		t.Fatalf("expected system wrapper, got %q", got)
	}
	if !strings.Contains(got, "do the thing") {
		t.Fatalf("expected user prompt, got %q", got)
	}
}

func TestReadAgyStreamConcatenatesLines(t *testing.T) {
	var received []string
	out, err := readAgyStream(bytes.NewBufferString("line one\nline two\n"), func(s string) {
		received = append(received, s)
	})
	if err != nil {
		t.Fatalf("readAgyStream: %v", err)
	}
	if out != "line one\nline two" {
		t.Fatalf("unexpected concatenated output: %q", out)
	}
	if len(received) != 2 || received[0] != "line one" || received[1] != "line two" {
		t.Fatalf("expected onLine called per line, got %v", received)
	}
}

func TestCreateAgyCLIStreamFnStreamsPlainText(t *testing.T) {
	recordFile := t.TempDir() + "/agy-record.jsonl"
	cwd := t.TempDir()

	restore := stubAgyRuntime(t, recordFile, "success", cwd)
	defer restore()

	fn := CreateAgyCLIStreamFn("ceo")
	chunks := collectStreamChunks(fn([]agent.Message{
		{Role: "system", Content: "You are the CEO."},
		{Role: "user", Content: "Ship it."},
	}, nil))

	text := joinedChunkText(chunks)
	if !strings.Contains(text, "shipped") {
		t.Fatalf("expected streamed plaintext reply, got %q", text)
	}

	records := readAgyHelperRecords(t, recordFile)
	if len(records) != 1 {
		t.Fatalf("expected 1 agy invocation, got %d", len(records))
	}
	lastArg := records[0].Args[len(records[0].Args)-1]
	if !strings.Contains(lastArg, "<system>") || !strings.Contains(lastArg, "Ship it.") {
		t.Fatalf("expected composed prompt as final argv arg, got %q", lastArg)
	}
}

func TestCreateAgyCLIStreamFnSurfacesMissingBinaryError(t *testing.T) {
	oldLookPath := agyLookPath
	agyLookPath = func(string) (string, error) { return "", exec.ErrNotFound }
	defer func() { agyLookPath = oldLookPath }()

	fn := CreateAgyCLIStreamFn("ceo")
	chunks := collectStreamChunks(fn([]agent.Message{{Role: "user", Content: "hi"}}, nil))
	if !hasErrorChunkContaining(chunks, "AGY CLI not found") {
		t.Fatalf("expected missing binary error, got %#v", chunks)
	}
}

func TestRunAgyOnceOversizedLine(t *testing.T) {
	recordFile := t.TempDir() + "/agy-record.jsonl"
	cwd := t.TempDir()

	restore := stubAgyRuntime(t, recordFile, "oversized-line", cwd)
	defer restore()

	done := make(chan struct{})
	var (
		out string
		err error
	)
	go func() {
		out, err = RunAgyOneShot("sys", "do the thing", cwd)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatal("RunAgyOneShot wedged on >5 MiB stdout line")
	}

	if err != nil {
		t.Fatalf("RunAgyOneShot: %v", err)
	}
	if !strings.Contains(out, "shipped the update") {
		t.Fatalf("trailing line missing from output: out len=%d", len(out))
	}
}

func readAgyHelperRecords(t *testing.T, path string) []agyHelperRecord {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read record file: %v", err)
	}
	var records []agyHelperRecord
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var record agyHelperRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("unmarshal helper record: %v", err)
		}
		records = append(records, record)
	}
	return records
}

func stubAgyRuntime(t *testing.T, recordFile string, scenario string, cwd string) func() {
	t.Helper()

	oldLookPath := agyLookPath
	oldCommand := agyCommand
	oldCommandContext := agyCommandContext
	oldGetwd := agyGetwd
	t.Setenv("GO_WANT_AGY_HELPER_PROCESS", "1")
	t.Setenv("AGY_TEST_RECORD_FILE", recordFile)
	t.Setenv("AGY_TEST_SCENARIO", scenario)
	t.Setenv("HOME", t.TempDir())

	agyLookPath = func(file string) (string, error) {
		return "/usr/bin/agy", nil
	}
	agyGetwd = func() (string, error) {
		return cwd, nil
	}
	agyCommand = func(name string, args ...string) *exec.Cmd {
		cmdArgs := []string{"-test.run=TestAgyHelperProcess", "--"}
		cmdArgs = append(cmdArgs, args...)
		return exec.Command(os.Args[0], cmdArgs...)
	}
	agyCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cmdArgs := []string{"-test.run=TestAgyHelperProcess", "--"}
		cmdArgs = append(cmdArgs, args...)
		return exec.CommandContext(ctx, os.Args[0], cmdArgs...)
	}

	return func() {
		agyLookPath = oldLookPath
		agyCommand = oldCommand
		agyCommandContext = oldCommandContext
		agyGetwd = oldGetwd
	}
}

func TestAgyHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_AGY_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	doubleDash := 0
	for i, arg := range args {
		if arg == "--" {
			doubleDash = i
			break
		}
	}
	agyArgs := append([]string(nil), args[doubleDash+1:]...)
	stdin, _ := io.ReadAll(os.Stdin)

	recordPath := os.Getenv("AGY_TEST_RECORD_FILE")
	record, _ := json.Marshal(agyHelperRecord{Args: agyArgs, Stdin: string(stdin)})
	file, err := os.OpenFile(recordPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open helper record file: %v", err)
	}
	if _, err := file.Write(append(record, '\n')); err != nil {
		t.Fatalf("write helper record: %v", err)
	}
	file.Close()

	if !containsArg(agyArgs, "--print") {
		t.Fatalf("missing `--print` subcommand/flag: %#v", agyArgs)
	}

	switch os.Getenv("AGY_TEST_SCENARIO") {
	case "success":
		_, _ = os.Stdout.WriteString("shipped the update\n")
		os.Exit(0)
	case "auth-error":
		_, _ = os.Stderr.WriteString("error: unauthorized\n")
		os.Exit(1)
	case "oversized-line":
		const huge = 5*1024*1024 + 7
		buf := make([]byte, huge)
		for i := range buf {
			buf[i] = 'z'
		}
		_, _ = os.Stdout.Write(buf)
		_, _ = os.Stdout.WriteString("\nshipped the update\n")
		os.Exit(0)
	default:
		t.Fatalf("unknown helper scenario: %s", os.Getenv("AGY_TEST_SCENARIO"))
	}
}
