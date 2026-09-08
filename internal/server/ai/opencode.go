package ai

import (
	"context"
	"os/exec"
)

type OpencodeProvider struct {
	Model          string
	TimeoutSeconds int
}

func (p *OpencodeProvider) Name() string {
	return "opencode"
}

func (p *OpencodeProvider) Available() bool {
	_, err := exec.LookPath("opencode")
	return err == nil
}

func (p *OpencodeProvider) Complete(ctx context.Context, system, user string) (string, error) {
	if !p.Available() {
		return "", ErrProviderUnavailable
	}
	args := append([]string{"run"}, modelArgs("-m", p.Model)...)
	return runCLI(ctx, "opencode", append(args, combinePrompt(system, user)), timeoutSeconds(p.TimeoutSeconds))
}
