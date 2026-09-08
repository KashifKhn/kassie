package ai

import (
	"context"
	"errors"
)

var (
	ErrProviderUnavailable = errors.New("ai provider is not available")
	ErrTimeout            = errors.New("ai provider timed out")
)

type Provider interface {
	Name() string
	Available() bool
	Complete(ctx context.Context, system, user string) (string, error)
}
