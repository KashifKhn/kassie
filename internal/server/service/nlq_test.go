package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/server/state"
	"github.com/KashifKhn/kassie/internal/shared/config"
	"github.com/KashifKhn/kassie/internal/shared/ctxutil"
)

type mockCompleter struct {
	name      string
	available bool
	responses []string
	err       error
	calls     int
	prompts   []string
}

func (m *mockCompleter) Name() string { return m.name }

func (m *mockCompleter) Available() bool { return m.available }

func (m *mockCompleter) Complete(ctx context.Context, system, user string) (string, error) {
	m.calls++
	m.prompts = append(m.prompts, user)
	if m.err != nil {
		return "", m.err
	}
	if m.calls <= len(m.responses) {
		return m.responses[m.calls-1], nil
	}
	return "", errors.New("no more responses")
}

var _ Completer = (*mockCompleter)(nil)

func TestStripFences(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "SELECT * FROM t", "SELECT * FROM t"},
		{"cql fence", "```cql\nSELECT * FROM t\n```", "SELECT * FROM t"},
		{"sql fence", "```sql\nSELECT 1\n```", "SELECT 1"},
		{"bare fence", "```\nSELECT 1\n```", "SELECT 1"},
		{"whitespace", "  SELECT 1  ", "SELECT 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripFences(tt.in); got != tt.want {
				t.Errorf("stripFences(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestBoundWithLimit(t *testing.T) {
	if got := boundWithLimit("SELECT * FROM t", 1); got != "SELECT * FROM t LIMIT 1" {
		t.Errorf("unexpected: %q", got)
	}
	if got := boundWithLimit("SELECT * FROM t LIMIT 100", 1); got != "SELECT * FROM t LIMIT 100" {
		t.Errorf("existing limit should be kept: %q", got)
	}
	if got := boundWithLimit("select * from t limit 5", 1); got != "select * from t limit 5" {
		t.Errorf("case-insensitive limit: %q", got)
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	p := buildSystemPrompt()
	if !strings.Contains(p, "SELECT") || !strings.Contains(p, "partition key") {
		t.Errorf("system prompt missing key rules: %q", p)
	}
}

func TestBuildUserPrompt(t *testing.T) {
	tables := []tableContext{
		{
			Table:          "events",
			PartitionKeys:  []string{"device_id"},
			ClusteringKeys: []string{"ts"},
			Columns: []columnContext{
				{Name: "device_id", Type: "uuid"},
				{Name: "ts", Type: "timestamp"},
				{Name: "severity", Type: "text"},
			},
		},
	}
	examples := []state.NLExample{
		{Question: "old question", CQL: "SELECT * FROM events"},
	}

	got := buildUserPrompt("critical events", tables, examples, false, "")
	for _, want := range []string{
		"Today is",
		"(UTC)",
		"table events",
		"partition key: device_id",
		"clustering: ts",
		"device_id uuid",
		"severity text",
		"Q: old question",
		"CQL: SELECT * FROM events",
		"Question: critical events",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("prompt missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "only the first") {
		t.Error("no truncation note expected")
	}
}

func TestBuildUserPromptTruncationNote(t *testing.T) {
	got := buildUserPrompt("q", nil, nil, true, "")
	if !strings.Contains(got, "only the first 50 tables") {
		t.Errorf("expected truncation note: %s", got)
	}
}

func TestBuildUserPromptTodayIsRealDate(t *testing.T) {
	got := buildUserPrompt("q", nil, nil, false, "")
	if !regexp.MustCompile(`Today is \d{4}-\d{2}-\d{2}`).MatchString(got) {
		t.Errorf("expected ISO date in prompt: %s", got)
	}
}

func TestBuildUserPromptPrivacyNamesOmitsTypes(t *testing.T) {
	tables := []tableContext{
		{
			Table:         "events",
			PartitionKeys: []string{"device_id"},
			Columns: []columnContext{
				{Name: "device_id"},
				{Name: "severity"},
			},
		},
	}
	got := buildUserPrompt("q", tables, nil, false, "")
	if strings.Contains(got, "uuid") || strings.Contains(got, "text") {
		t.Errorf("names-only privacy should omit types: %s", got)
	}
	if !strings.Contains(got, "device_id") {
		t.Errorf("column names must be present: %s", got)
	}
}

func TestBuildUserPromptFeedback(t *testing.T) {
	got := buildUserPrompt("q", nil, nil, false, "unknown table")
	if !strings.Contains(got, "unknown table") || !strings.Contains(got, "corrected SELECT") {
		t.Errorf("feedback missing from prompt: %s", got)
	}
}

func TestBuildUserPromptEmptySchema(t *testing.T) {
	got := buildUserPrompt("q", nil, nil, false, "")
	if !strings.Contains(got, "Schema:") || !strings.Contains(got, "Question: q") {
		t.Errorf("prompt should still render: %s", got)
	}
}

func TestNlToCqlValidation(t *testing.T) {
	tests := []struct {
		name    string
		resp    string
		wantErr bool
	}{
		{"valid select", "SELECT * FROM events", false},
		{"fenced select", "```cql\nSELECT * FROM events\n```", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := normalizeAdhocCQL(stripFences(tt.resp)); (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildTableContextOrdersKeysByPosition(t *testing.T) {
	rows := []map[string]interface{}{
		{"column_name": "day", "type": "date", "kind": "partition_key", "position": 1},
		{"column_name": "severity", "type": "text", "kind": "regular", "position": -1},
		{"column_name": "ts", "type": "timestamp", "kind": "clustering", "position": 1},
		{"column_name": "device_id", "type": "uuid", "kind": "partition_key", "position": 0},
	}

	tc := buildTableContext("events", rows, "names+types")

	if len(tc.PartitionKeys) != 2 || tc.PartitionKeys[0] != "device_id" || tc.PartitionKeys[1] != "day" {
		t.Fatalf("partition keys must be position-ordered, got %v", tc.PartitionKeys)
	}
	if len(tc.ClusteringKeys) != 1 || tc.ClusteringKeys[0] != "ts" {
		t.Fatalf("unexpected clustering keys: %v", tc.ClusteringKeys)
	}
	if len(tc.Columns) != 4 {
		t.Fatalf("expected 4 columns, got %d", len(tc.Columns))
	}
}

func TestBuildTableContextPrivacyNames(t *testing.T) {
	rows := []map[string]interface{}{
		{"column_name": "id", "type": "uuid", "kind": "partition_key", "position": 0},
	}

	tc := buildTableContext("t", rows, "names")
	if tc.Columns[0].Type != "" {
		t.Fatalf("names privacy must drop types, got %q", tc.Columns[0].Type)
	}

	tc = buildTableContext("t", rows, "names+types")
	if tc.Columns[0].Type != "uuid" {
		t.Fatalf("names+types must keep types, got %q", tc.Columns[0].Type)
	}
}

func TestDefaultCompleterResolvesProfile(t *testing.T) {
	pinned := DefaultCompleter(&config.Profile{AI: &config.AIConfig{Provider: "opencode"}}, nil)
	if pinned == nil || pinned.Name() != "opencode" {
		t.Fatalf("expected opencode, got %v", pinned)
	}

	if got := DefaultCompleter(&config.Profile{AI: &config.AIConfig{Provider: "bogus"}}, nil); got != nil {
		t.Fatalf("expected nil for unknown provider, got %v", got)
	}

	withModel := DefaultCompleter(&config.Profile{AI: &config.AIConfig{Provider: "claude", Model: "sonnet"}}, nil)
	if withModel == nil || withModel.Name() != "claude" {
		t.Fatalf("expected claude with model, got %v", withModel)
	}

	override := DefaultCompleter(&config.Profile{AI: &config.AIConfig{Provider: "opencode"}}, &CompleterOverride{Provider: "claude", Model: "sonnet"})
	if override == nil || override.Name() != "claude" {
		t.Fatalf("override must win, got %v", override)
	}
}

func TestAIServiceNlToCqlRequiresQuestion(t *testing.T) {
	store := newMockSessionStore()
	svc := NewAIService(store, DefaultCompleter, state.NewNLExampleStore(""))

	_, err := svc.NlToCql(context.Background(), &pb.NlToCqlRequest{})
	if err == nil {
		t.Fatal("expected error for empty question")
	}
}

func TestAIServiceNlToCqlProviderUnavailable(t *testing.T) {
	store := newMockSessionStore()
	svc := NewAIService(store, func(p *config.Profile, o *CompleterOverride) Completer { return nil }, state.NewNLExampleStore(""))

	ctx := ctxutil.WithSessionID(context.Background(), "sess")
	store.Create("sess", &config.Profile{Name: "test", Keyspace: "app"}, nil)

	_, err := svc.NlToCql(ctx, &pb.NlToCqlRequest{Question: "all events"})
	if err == nil {
		t.Fatal("expected provider unavailable error")
	}
}

func TestAIServiceNlToCqlRequiresKeyspace(t *testing.T) {
	store := newMockSessionStore()
	store.Create("sess", &config.Profile{Name: "test"}, nil)
	svc := NewAIService(store, DefaultCompleter, state.NewNLExampleStore(""))

	ctx := ctxutil.WithSessionID(context.Background(), "sess")
	_, err := svc.NlToCql(ctx, &pb.NlToCqlRequest{Question: "all events"})
	if err == nil {
		t.Fatal("expected keyspace required error")
	}
}

func TestAIServiceAcceptExampleRoundTrip(t *testing.T) {
	store := newMockSessionStore()
	store.Create("sess", &config.Profile{Name: "test"}, nil)
	svc := NewAIService(store, DefaultCompleter, state.NewNLExampleStore(""))

	ctx := ctxutil.WithSessionID(context.Background(), "sess")

	_, err := svc.AcceptExample(ctx, &pb.AcceptExampleRequest{})
	if err == nil {
		t.Fatal("expected error for empty example")
	}

	accept, err := svc.AcceptExample(ctx, &pb.AcceptExampleRequest{Question: "q", Cql: "SELECT 1", Keyspace: "app"})
	if err != nil {
		t.Fatal(err)
	}

	list, err := svc.ListExamples(ctx, &pb.ListExamplesRequest{Keyspace: "app"})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Examples) != 1 || list.Examples[0].Question != "q" {
		t.Fatalf("unexpected examples: %+v", list.Examples)
	}

	_, err = svc.DeleteExample(ctx, &pb.DeleteExampleRequest{Id: accept.Id})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := svc.DeleteExample(ctx, &pb.DeleteExampleRequest{Id: accept.Id}); err == nil {
		t.Fatal("expected delete of missing example to fail")
	}
}
