package ai

import (
	"context"
	"os/exec"
)

type ClaudeProvider struct {
	Model          string
	TimeoutSeconds int
}

func (p *ClaudeProvider) Name() string {
	return "claude"
}

func (p *ClaudeProvider) Available() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

func (p *ClaudeProvider) Complete(ctx context.Context, system, user string) (string, error) {
	if !p.Available() {
		return "", ErrProviderUnavailable
	}
	args := append([]string{"-p"}, modelArgs("--model", p.Model)...)
	return runCLI(ctx, "claude", append(args, combinePrompt(system, user)), timeoutSeconds(p.TimeoutSeconds))
}
