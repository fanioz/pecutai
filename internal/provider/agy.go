package provider

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/nex-crm/wuphf/internal/agent"
	"github.com/nex-crm/wuphf/internal/config"
	"github.com/nex-crm/wuphf/internal/runtimebin"
)

var (
	agyLookPath       = runtimebin.LookPath
	agyCommand        = exec.Command
	agyCommandContext = exec.CommandContext
	agyGetwd          = os.Getwd
)

func init() {
	Register(&Entry{
		Kind:       KindAgy,
		StreamFn:   CreateAgyCLIStreamFn,
		OneShot:    RunAgyOneShot,
		OneShotCtx: RunAgyOneShotCtx,
		Capabilities: Capabilities{
			PaneEligible:    false,
			SupportsOneShot: true,
		},
	})
}

// CreateAgyCLIStreamFn returns a StreamFn that runs the AGY CLI non-interactively.
// Each invocation is ephemeral: WUPHF owns the conversation history and hands
// AGY a fresh prompt every turn.
//
// AGY emits plain text on stdout, so we stream stdout line-by-line as text chunks.
func CreateAgyCLIStreamFn(agentSlug string) agent.StreamFn {
	return func(msgs []agent.Message, tools []agent.AgentTool) <-chan agent.StreamChunk {
		ch := make(chan agent.StreamChunk, 64)
		go func() {
			defer close(ch)

			if _, err := agyLookPath("agy"); err != nil {
				ch <- agent.StreamChunk{Type: "error", Content: "AGY CLI not found. Install agy or use /provider to choose a different provider."}
				return
			}

			cwd, err := agyGetwd()
			if err != nil {
				ch <- agent.StreamChunk{Type: "error", Content: fmt.Sprintf("resolve working directory: %v", err)}
				return
			}

			systemPrompt, prompt := buildClaudePrompts(msgs)
			if prompt == "" {
				prompt = "Proceed with the task."
			}

			startedAt := time.Now()
			var firstEventAt time.Time
			var firstTextAt time.Time
			text, err := runAgyOnce(systemPrompt, prompt, cwd, func(line string) {
				if firstEventAt.IsZero() {
					firstEventAt = time.Now()
				}
				if strings.TrimSpace(line) == "" {
					return
				}
				if firstTextAt.IsZero() {
					firstTextAt = time.Now()
				}
				ch <- agent.StreamChunk{Type: "text", Content: line}
			})
			if err != nil {
				appendAgyLatencyLog(agentSlug, fmt.Sprintf("status=error total_ms=%d first_event_ms=%d first_text_ms=%d detail=%q",
					time.Since(startedAt).Milliseconds(),
					durationMillis(startedAt, firstEventAt),
					durationMillis(startedAt, firstTextAt),
					err.Error(),
				))
				ch <- agent.StreamChunk{Type: "error", Content: describeAgyFailure(err)}
				return
			}
			appendAgyLatencyLog(agentSlug, fmt.Sprintf("status=ok total_ms=%d first_event_ms=%d first_text_ms=%d final_chars=%d",
				time.Since(startedAt).Milliseconds(),
				durationMillis(startedAt, firstEventAt),
				durationMillis(startedAt, firstTextAt),
				len(text),
			))
			if firstTextAt.IsZero() && strings.TrimSpace(text) != "" {
				streamTextChunks(ch, text)
			}
		}()
		return ch
	}
}

// RunAgyOneShot runs AGY once with the given system prompt and user prompt.
func RunAgyOneShot(systemPrompt, prompt, cwd string) (string, error) {
	if cwd == "" {
		var err error
		cwd, err = agyGetwd()
		if err != nil {
			return "", err
		}
	}
	return runAgyOnce(systemPrompt, prompt, cwd, nil)
}

// RunAgyOneShotCtx runs AGY once and binds the child process lifetime to ctx.
func RunAgyOneShotCtx(ctx context.Context, systemPrompt, prompt, cwd string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if cwd == "" {
		var err error
		cwd, err = agyGetwd()
		if err != nil {
			return "", err
		}
	}
	return runAgyOnceCtx(ctx, systemPrompt, prompt, cwd, nil)
}

func runAgyOnce(systemPrompt, prompt, cwd string, onLine func(string)) (string, error) {
	promptText := buildAgyPrompt(systemPrompt, prompt)
	args := buildAgyArgs(config.ResolveAgyModel(), promptText)
	cmd := agyCommand("agy", args...)
	return runAgyCommand(context.Background(), cmd, cwd, onLine)
}

func runAgyOnceCtx(ctx context.Context, systemPrompt, prompt, cwd string, onLine func(string)) (string, error) {
	promptText := buildAgyPrompt(systemPrompt, prompt)
	args := buildAgyArgs(config.ResolveAgyModel(), promptText)
	cmd := agyCommandContext(ctx, "agy", args...)
	return runAgyCommand(ctx, cmd, cwd, onLine)
}

func runAgyCommand(ctx context.Context, cmd *exec.Cmd, cwd string, onLine func(string)) (string, error) {
	cmd.Dir = cwd
	cmd.Env = append(filteredEnv(nil), "NO_COLOR=1")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("attach agy stdout: %w", err)
	}

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", err
	}

	output, readErr := readAgyStream(stdout, onLine)
	if err := cmd.Wait(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", ctxErr
		}
		detail := strings.TrimSpace(stderr.String())
		if detail != "" {
			return "", fmt.Errorf("%w: %s", err, detail)
		}
		return "", err
	}
	if readErr != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", ctxErr
		}
		return "", readErr
	}
	text := strings.TrimSpace(output)
	if text == "" {
		return "", fmt.Errorf("agy returned no output")
	}
	return text, nil
}

func readAgyStream(r io.Reader, onLine func(string)) (string, error) {
	var sb strings.Builder
	err := DrainStreamLines(r, func(raw string) {
		line := strings.TrimRight(raw, "\r\n")
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(line)
		if onLine != nil {
			onLine(line)
		}
	})
	if err != nil {
		return sb.String(), fmt.Errorf("read agy stream: %w", err)
	}
	return sb.String(), nil
}

func buildAgyArgs(model, prompt string) []string {
	args := []string{"--print"}
	if strings.TrimSpace(model) != "" {
		args = append(args, "--model", strings.TrimSpace(model))
	}
	if strings.TrimSpace(prompt) != "" {
		args = append(args, prompt)
	}
	return args
}

func buildAgyPrompt(systemPrompt, prompt string) string {
	var parts []string
	if s := strings.TrimSpace(systemPrompt); s != "" {
		parts = append(parts, "<system>\n"+escapeAgySystemWrapper(s)+"\n</system>")
	}
	if p := strings.TrimSpace(prompt); p != "" {
		parts = append(parts, escapeAgySystemWrapper(p))
	}
	return strings.Join(parts, "\n\n")
}

func escapeAgySystemWrapper(s string) string {
	s = strings.ReplaceAll(s, "</system>", "</\u200bsystem>")
	s = strings.ReplaceAll(s, "<system>", "<\u200bsystem>")
	return s
}

func describeAgyFailure(err error) string {
	text := strings.ToLower(strings.TrimSpace(err.Error()))
	if strings.Contains(text, "login") || strings.Contains(text, "auth") || strings.Contains(text, "unauthorized") || strings.Contains(text, "credentials") {
		return "AGY CLI is not authenticated. Please run 'agy install' or sign in to your AGY account."
	}
	return fmt.Sprintf("agy exited with error: %v", err)
}

func appendAgyLatencyLog(agentSlug string, line string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	logDir := filepath.Join(home, ".wuphf", "logs")
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return
	}
	path := filepath.Join(logDir, "agy-latency.log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	_, _ = fmt.Fprintf(f, "[%s] agent=%s %s\n", time.Now().Format(time.RFC3339), strings.TrimSpace(agentSlug), strings.TrimSpace(line))
}
