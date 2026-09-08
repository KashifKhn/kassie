package ai

import (
	"fmt"
)

var detectionOrder = []string{"opencode", "claude", "codex", "gemini"}

func NewProvider(name string, model string, timeoutSeconds int) (Provider, error) {
	switch name {
	case "opencode":
		return &OpencodeProvider{Model: model, TimeoutSeconds: timeoutSeconds}, nil
	case "claude":
		return &ClaudeProvider{Model: model, TimeoutSeconds: timeoutSeconds}, nil
	case "codex":
		return &CodexProvider{Model: model, TimeoutSeconds: timeoutSeconds}, nil
	case "gemini":
		return &GeminiProvider{Model: model, TimeoutSeconds: timeoutSeconds}, nil
	case "auto", "":
		for _, name := range detectionOrder {
			p, err := NewProvider(name, model, timeoutSeconds)
			if err != nil {
				return nil, err
			}
			if p.Available() {
				return p, nil
			}
		}
		return nil, fmt.Errorf("%w: none of opencode, claude, codex, gemini found in PATH", ErrProviderUnavailable)
	}
	return nil, fmt.Errorf("unknown ai provider: %s", name)
}

func DetectAvailable() []string {
	out := make([]string, 0, len(detectionOrder))
	for _, name := range detectionOrder {
		p, err := NewProvider(name, "", 0)
		if err != nil {
			continue
		}
		if p.Available() {
			out = append(out, name)
		}
	}
	return out
}
