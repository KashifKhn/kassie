package config

import (
	"testing"

	"github.com/gocql/gocql"
)

func TestParseConsistencyLevel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  gocql.Consistency
		err   bool
	}{
		{"empty defaults to quorum", "", gocql.Quorum, false},
		{"one", "ONE", gocql.One, false},
		{"lowercase local_quorum", "local_quorum", gocql.LocalQuorum, false},
		{"mixed case spaces", "  Local_Quorum ", gocql.LocalQuorum, false},
		{"each_quorum", "EACH_QUORUM", gocql.EachQuorum, false},
		{"all", "ALL", gocql.All, false},
		{"any", "ANY", gocql.Any, false},
		{"three", "THREE", gocql.Three, false},
		{"bogus", "MOSTLY", gocql.Quorum, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConsistencyLevel(tt.input)
			if (err != nil) != tt.err {
				t.Fatalf("err = %v, wantErr %v", err, tt.err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConnTuningValidate(t *testing.T) {
	tests := []struct {
		name    string
		tuning  *ConnTuningConfig
		wantErr bool
	}{
		{"nil is fine", nil, false},
		{"empty block", &ConnTuningConfig{}, false},
		{"valid everything", &ConnTuningConfig{Consistency: "LOCAL_QUORUM", Timeout: "30s", PoolSize: 10}, false},
		{"bad consistency", &ConnTuningConfig{Consistency: "SOME"}, true},
		{"bad timeout format", &ConnTuningConfig{Timeout: "soon"}, true},
		{"negative timeout", &ConnTuningConfig{Timeout: "-5s"}, true},
		{"timeout over cap", &ConnTuningConfig{Timeout: "6m"}, true},
		{"pool size zero ok", &ConnTuningConfig{PoolSize: 0}, false},
		{"pool size negative", &ConnTuningConfig{PoolSize: -1}, true},
		{"pool size over cap", &ConnTuningConfig{PoolSize: 101}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tuning.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProfileValidateRejectsBadTuning(t *testing.T) {
	p := &Profile{Name: "x", Hosts: []string{"h"}, Port: 9042, Connection: &ConnTuningConfig{PoolSize: -5}}
	if err := p.Validate(); err != ErrInvalidPoolSize {
		t.Fatalf("err = %v, want ErrInvalidPoolSize", err)
	}
}
