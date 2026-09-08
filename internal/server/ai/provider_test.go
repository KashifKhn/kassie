package ai

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeFakeProvider(t *testing.T, name, output string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	script := "#!/bin/sh\ncat <<'EOF'\n" + output + "\nEOF\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return path
}

func TestNewProviderAutoDetects(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)

	p, err := NewProvider("auto", "", 30)
	if err == nil {
		t.Fatalf("expected error when no providers in PATH, got %v", p.Name())
	}
}

func TestNewProviderAutoPrefersOpencode(t *testing.T) {
	writeFakeProvider(t, "opencode", "ok")
	p, err := NewProvider("auto", "", 30)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "opencode" {
		t.Fatalf("expected opencode, got %s", p.Name())
	}
}

func TestNewProviderAutoFallsBack(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	writeFakeProvider(t, "codex", "ok")

	p, err := NewProvider("auto", "", 30)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "codex" {
		t.Fatalf("expected codex fallback, got %s", p.Name())
	}
}

func TestNewProviderPinned(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	writeFakeProvider(t, "claude", "ok")

	_, err := NewProvider("nonexistent", "", 30)
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}

	p, err := NewProvider("claude", "", 30)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name() != "claude" {
		t.Fatalf("expected claude, got %s", p.Name())
	}
}

func TestDetectAvailable(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	writeFakeProvider(t, "gemini", "ok")
	writeFakeProvider(t, "opencode", "ok")

	got := DetectAvailable()
	if len(got) != 2 || got[0] != "opencode" || got[1] != "gemini" {
		t.Fatalf("expected [opencode gemini], got %v", got)
	}
}

func TestProviderCompleteRunsCLI(t *testing.T) {
	writeFakeProvider(t, "opencode", "SELECT * FROM test.documents")

	p, err := NewProvider("opencode", "", 30)
	if err != nil {
		t.Fatal(err)
	}

	out, err := p.Complete(context.Background(), "system rules", "user question")
	if err != nil {
		t.Fatal(err)
	}
	if out != "SELECT * FROM test.documents" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestModelArgsPassedToCLI(t *testing.T) {
	dir := t.TempDir()
	script := "#!/bin/sh\nfor a in \"$@\"; do printf '%s ' \"$a\"; done\n"
	for _, name := range []string{"opencode", "claude", "codex", "gemini"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir)

	tests := []struct {
		provider string
		wantArgs string
	}{
		{"opencode", "run -m gpt-x "},
		{"claude", "-p --model gpt-x "},
		{"codex", "exec -m gpt-x "},
		{"gemini", "-p -m gpt-x "},
	}
	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			p, err := NewProvider(tt.provider, "gpt-x", 30)
			if err != nil {
				t.Fatal(err)
			}
			if p.Name() != tt.provider {
				t.Fatalf("expected %s, got %s", tt.provider, p.Name())
			}
			got, err := p.Complete(context.Background(), "sys", "usr")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(got, tt.wantArgs) {
				t.Fatalf("%s: expected args %q, got %q", tt.provider, tt.wantArgs, got)
			}
		})
	}
}

func TestModelArgsEmptyOmitsFlag(t *testing.T) {
	if got := modelArgs("-m", ""); len(got) != 0 {
		t.Fatalf("empty model should omit flag, got %v", got)
	}
	if got := modelArgs("-m", "  "); len(got) != 0 {
		t.Fatalf("whitespace model should omit flag, got %v", got)
	}
	if got := modelArgs("-m", "gpt-x"); len(got) != 2 || got[0] != "-m" || got[1] != "gpt-x" {
		t.Fatalf("unexpected flag args: %v", got)
	}
}

func TestProviderCompleteUnavailable(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)

	p, err := NewProvider("claude", "", 30)
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.Complete(context.Background(), "s", "u")
	if err != ErrProviderUnavailable {
		t.Fatalf("expected ErrProviderUnavailable, got %v", err)
	}
}

func TestRunCLITimeout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sleeper")
	script := "#!/bin/sh\nsleep 5\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	_, err := runCLI(context.Background(), "sleeper", nil, 100*time.Millisecond)
	if err != ErrTimeout {
		t.Fatalf("expected ErrTimeout, got %v", err)
	}
}

func TestTimeoutSecondsDefaults(t *testing.T) {
	tests := []struct {
		in   int
		want time.Duration
	}{
		{0, defaultTimeout},
		{-5, defaultTimeout},
		{120, 120 * time.Second},
	}
	for _, tt := range tests {
		if got := timeoutSeconds(tt.in); got != tt.want {
			t.Errorf("timeoutSeconds(%d) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestCombinePrompt(t *testing.T) {
	got := combinePrompt("sys", "usr")
	if got != "sys\n\nusr" {
		t.Fatalf("unexpected: %q", got)
	}
}
