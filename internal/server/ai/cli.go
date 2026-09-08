package ai

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const defaultTimeout = 60 * time.Second

func timeoutSeconds(cfgSeconds int) time.Duration {
	if cfgSeconds <= 0 {
		return defaultTimeout
	}
	return time.Duration(cfgSeconds) * time.Second
}

func runCLI(ctx context.Context, bin string, args []string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	wd, err := os.MkdirTemp("", "kassie-ai")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(wd)

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = wd

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", ErrTimeout
		}
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return "", fmt.Errorf("%w: %s", ErrProviderUnavailable, msg)
		}
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

func combinePrompt(system, user string) string {
	return system + "\n\n" + user
}

func modelArgs(flag, model string) []string {
	if strings.TrimSpace(model) == "" {
		return nil
	}
	return []string{flag, model}
}
