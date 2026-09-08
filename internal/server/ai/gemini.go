package ai

import (
	"context"
	"os/exec"
)

type GeminiProvider struct {
	Model          string
	TimeoutSeconds int
}

func (p *GeminiProvider) Name() string {
	return "gemini"
}

func (p *GeminiProvider) Available() bool {
	_, err := exec.LookPath("gemini")
	return err == nil
}

func (p *GeminiProvider) Complete(ctx context.Context, system, user string) (string, error) {
	if !p.Available() {
		return "", ErrProviderUnavailable
	}
	args := append([]string{"-p"}, modelArgs("-m", p.Model)...)
	return runCLI(ctx, "gemini", append(args, combinePrompt(system, user)), timeoutSeconds(p.TimeoutSeconds))
}
