package ai

import (
	"context"
	"os/exec"
)

type CodexProvider struct {
	Model          string
	TimeoutSeconds int
}

func (p *CodexProvider) Name() string {
	return "codex"
}

func (p *CodexProvider) Available() bool {
	_, err := exec.LookPath("codex")
	return err == nil
}

func (p *CodexProvider) Complete(ctx context.Context, system, user string) (string, error) {
	if !p.Available() {
		return "", ErrProviderUnavailable
	}
	args := append([]string{"exec"}, modelArgs("-m", p.Model)...)
	return runCLI(ctx, "codex", append(args, combinePrompt(system, user)), timeoutSeconds(p.TimeoutSeconds))
}
