package db

import (
	"testing"
	"time"

	"github.com/KashifKhn/kassie/internal/shared/config"
	"github.com/gocql/gocql"
)

func TestProfileToConnectionConfig_Tuning(t *testing.T) {
	tests := []struct {
		name            string
		tuning          *config.ConnTuningConfig
		wantConsistency gocql.Consistency
		wantTimeout     time.Duration
		wantPool        int
	}{
		{
			name:            "defaults when absent",
			tuning:          nil,
			wantConsistency: gocql.Quorum,
			wantTimeout:     10 * time.Second,
			wantPool:        5,
		},
		{
			name:            "full override",
			tuning:          &config.ConnTuningConfig{Consistency: "ONE", Timeout: "45s", PoolSize: 12},
			wantConsistency: gocql.One,
			wantTimeout:     45 * time.Second,
			wantPool:        12,
		},
		{
			name:            "partial override keeps defaults",
			tuning:          &config.ConnTuningConfig{PoolSize: 3},
			wantConsistency: gocql.Quorum,
			wantTimeout:     10 * time.Second,
			wantPool:        3,
		},
		{
			name:            "invalid values fall back to defaults",
			tuning:          &config.ConnTuningConfig{Consistency: "NOPE", Timeout: "never"},
			wantConsistency: gocql.Quorum,
			wantTimeout:     10 * time.Second,
			wantPool:        5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := &config.Profile{Name: "p", Hosts: []string{"127.0.0.1"}, Port: 9042, Connection: tt.tuning}
			cfg := ProfileToConnectionConfig(profile)

			if cfg.Consistency != tt.wantConsistency {
				t.Errorf("consistency = %v, want %v", cfg.Consistency, tt.wantConsistency)
			}
			if cfg.Timeout != tt.wantTimeout {
				t.Errorf("timeout = %v, want %v", cfg.Timeout, tt.wantTimeout)
			}
			if cfg.PoolSize != tt.wantPool {
				t.Errorf("pool = %d, want %d", cfg.PoolSize, tt.wantPool)
			}
		})
	}
}
