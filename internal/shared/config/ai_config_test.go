package config

import "testing"

func TestAIConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *AIConfig
		wantErr bool
	}{
		{name: "nil config", cfg: nil, wantErr: false},
		{name: "empty defaults", cfg: &AIConfig{}, wantErr: false},
		{name: "auto provider", cfg: &AIConfig{Provider: "auto"}, wantErr: false},
		{name: "opencode provider", cfg: &AIConfig{Provider: "opencode", TimeoutSeconds: 60}, wantErr: false},
		{name: "claude provider", cfg: &AIConfig{Provider: "claude"}, wantErr: false},
		{name: "codex provider", cfg: &AIConfig{Provider: "codex"}, wantErr: false},
		{name: "gemini provider", cfg: &AIConfig{Provider: "gemini"}, wantErr: false},
		{name: "unknown provider", cfg: &AIConfig{Provider: "chatgpt"}, wantErr: true},
		{name: "privacy names", cfg: &AIConfig{Privacy: "names"}, wantErr: false},
		{name: "privacy names+types", cfg: &AIConfig{Privacy: "names+types"}, wantErr: false},
		{name: "privacy invalid", cfg: &AIConfig{Privacy: "full"}, wantErr: true},
		{name: "timeout valid", cfg: &AIConfig{TimeoutSeconds: 120}, wantErr: false},
		{name: "timeout zero", cfg: &AIConfig{TimeoutSeconds: 0}, wantErr: false},
		{name: "timeout too large", cfg: &AIConfig{TimeoutSeconds: 601}, wantErr: true},
		{name: "timeout negative", cfg: &AIConfig{TimeoutSeconds: -1}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProfileValidateAI(t *testing.T) {
	if err := (&Profile{Name: "test", Hosts: []string{"localhost"}, Port: 9042}).Validate(); err != nil {
		t.Fatalf("base profile should be valid, got %v", err)
	}

	p := Profile{Name: "test", Hosts: []string{"localhost"}, Port: 9042}
	p.AI = &AIConfig{Provider: "opencode", Privacy: "names", TimeoutSeconds: 30}
	if err := p.Validate(); err != nil {
		t.Fatalf("profile with valid ai config should pass, got %v", err)
	}
	p.AI = &AIConfig{Provider: "bad"}
	if err := p.Validate(); err == nil {
		t.Fatal("profile with invalid ai provider should fail")
	}
}
