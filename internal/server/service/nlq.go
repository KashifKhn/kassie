package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/KashifKhn/kassie/internal/server/ai"
	"github.com/KashifKhn/kassie/internal/server/state"
	"github.com/KashifKhn/kassie/internal/shared/config"
	pb "github.com/KashifKhn/kassie/api/gen/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxPromptTables = 50
const maxFewShotExamples = 5

type tableContext struct {
	Table          string
	PartitionKeys  []string
	ClusteringKeys []string
	Columns        []columnContext
}

type columnContext struct {
	Name string
	Type string
}

type namedColumn struct {
	name     string
	typeName string
	kind     string
	position int
}

type Completer interface {
	Name() string
	Available() bool
	Complete(ctx context.Context, system, user string) (string, error)
}

type CompleterFunc func(profile *config.Profile, override *CompleterOverride) Completer

type CompleterOverride struct {
	Provider string
	Model    string
}

func DefaultCompleter(profile *config.Profile, override *CompleterOverride) Completer {
	name := "auto"
	model := ""
	timeout := 0
	if profile != nil && profile.AI != nil {
		if profile.AI.Provider != "" {
			name = profile.AI.Provider
		}
		model = profile.AI.Model
		timeout = profile.AI.TimeoutSeconds
	}
	if override != nil {
		if override.Provider != "" {
			name = override.Provider
		}
		if override.Model != "" {
			model = override.Model
		}
	}
	provider, err := ai.NewProvider(name, model, timeout)
	if err != nil {
		return nil
	}
	return provider
}

type AIService struct {
	pb.UnimplementedAIServiceServer
	store     SessionStore
	completer CompleterFunc
	examples  *state.NLExampleStore
}

func NewAIService(store SessionStore, completer CompleterFunc, examples *state.NLExampleStore) *AIService {
	return &AIService{store: store, completer: completer, examples: examples}
}

func (s *AIService) NlToCql(ctx context.Context, req *pb.NlToCqlRequest) (*pb.NlToCqlResponse, error) {
	question := strings.TrimSpace(req.Question)
	if question == "" {
		return nil, status.Error(codes.InvalidArgument, "question is required")
	}

	session, err := GetSessionFromContext(ctx, s.store)
	if err != nil {
		return nil, err
	}

	provider := s.completer(session.Profile, &CompleterOverride{
		Provider: req.Provider,
		Model:    req.Model,
	})
	if provider == nil || !provider.Available() {
		return nil, status.Error(codes.FailedPrecondition, "no ai provider available; install opencode, claude, codex, or gemini")
	}

	privacy := "names+types"
	if session.Profile.AI != nil && session.Profile.AI.Privacy != "" {
		privacy = session.Profile.AI.Privacy
	}

	tables, truncated, err := s.schemaTables(ctx, session, req.Keyspace, req.Table, privacy)
	if err != nil {
		return nil, err
	}

	examples := s.examples.List(session.Profile.Name, req.Keyspace, maxFewShotExamples)

	result, err := attempt(ctx, provider, question, tables, examples, truncated)
	if err != nil {
		return nil, err
	}

	cql, validateErr := normalizeAdhocCQL(stripFences(result))
	if validateErr != nil {
		retry, retryErr := attemptWithFeedback(ctx, provider, question, tables, examples, truncated, validateErr.Error())
		if retryErr != nil {
			return nil, retryErr
		}
		cql, validateErr = normalizeAdhocCQL(stripFences(retry))
		if validateErr != nil {
			return nil, status.Errorf(codes.InvalidArgument, "generated cql failed validation: %v", validateErr)
		}
		result = retry
	}

	if err := s.dryRun(ctx, session, cql); err != nil {
		retry, retryErr := attemptWithFeedback(ctx, provider, question, tables, examples, truncated, err.Error())
		if retryErr != nil {
			return nil, status.Errorf(codes.InvalidArgument, "generated cql failed dry-run: %v", err)
		}
		retryCQL, retryValidate := normalizeAdhocCQL(stripFences(retry))
		if retryValidate != nil {
			return nil, status.Errorf(codes.InvalidArgument, "generated cql failed validation: %v", retryValidate)
		}
		if dryErr := s.dryRun(ctx, session, retryCQL); dryErr != nil {
			return nil, status.Errorf(codes.InvalidArgument, "generated cql failed dry-run: %v", dryErr)
		}
		cql = retryCQL
		result = retry
	}

	return &pb.NlToCqlResponse{
		Cql:        cql,
		Validated:  true,
		Explanation: result,
		Provider:   provider.Name(),
	}, nil
}

func attempt(ctx context.Context, provider Completer, question string, tables []tableContext, examples []state.NLExample, truncated bool) (string, error) {
	system := buildSystemPrompt()
	user := buildUserPrompt(question, tables, examples, truncated, "")
	out, err := provider.Complete(ctx, system, user)
	if err != nil {
		return "", status.Errorf(codes.Internal, "ai provider failed: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		return "", status.Error(codes.Internal, "ai provider returned empty response")
	}
	return out, nil
}

func attemptWithFeedback(ctx context.Context, provider Completer, question string, tables []tableContext, examples []state.NLExample, truncated bool, feedback string) (string, error) {
	system := buildSystemPrompt()
	user := buildUserPrompt(question, tables, examples, truncated, feedback)
	out, err := provider.Complete(ctx, system, user)
	if err != nil {
		return "", status.Errorf(codes.Internal, "ai provider retry failed: %v", err)
	}
	return out, nil
}

func (s *AIService) schemaTables(ctx context.Context, session *state.Session, keyspace, table, privacy string) ([]tableContext, bool, error) {
	effectiveKeyspace := keyspace
	if effectiveKeyspace == "" {
		effectiveKeyspace = session.Profile.Keyspace
	}
	if effectiveKeyspace == "" {
		return nil, false, status.Error(codes.InvalidArgument, "keyspace is required for nl queries")
	}
	if session.Connection == nil {
		return nil, false, status.Error(codes.Internal, "no database connection available")
	}

	query := `SELECT table_name FROM system_schema.tables WHERE keyspace_name = ?`
	rows, err := session.Connection.FetchAll(ctx, query, effectiveKeyspace)
	if err != nil {
		return nil, false, status.Errorf(codes.Internal, "failed to fetch tables: %v", err)
	}

	names := make([]string, 0, len(rows))
	for _, row := range rows {
		name, _ := row["table_name"].(string)
		if name == "" {
			continue
		}
		if table != "" && name != table {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	truncated := false
	if len(names) > maxPromptTables {
		names = names[:maxPromptTables]
		truncated = true
	}

	tables := make([]tableContext, 0, len(names))
	for _, name := range names {
		tc, err := s.tableContext(ctx, session, effectiveKeyspace, name, privacy)
		if err != nil {
			continue
		}
		tables = append(tables, tc)
	}
	return tables, truncated, nil
}

func (s *AIService) tableContext(ctx context.Context, session *state.Session, keyspace, table, privacy string) (tableContext, error) {
	query := `SELECT column_name, type, kind, position FROM system_schema.columns WHERE keyspace_name = ? AND table_name = ?`
	rows, err := session.Connection.FetchAll(ctx, query, keyspace, table)
	if err != nil {
		return tableContext{}, err
	}
	return buildTableContext(table, rows, privacy), nil
}

func buildTableContext(table string, rows []map[string]interface{}, privacy string) tableContext {
	cols := make([]namedColumn, 0, len(rows))
	for _, row := range rows {
		colName, _ := row["column_name"].(string)
		if colName == "" {
			continue
		}
		colType, _ := row["type"].(string)
		kind, _ := row["kind"].(string)
		position, _ := row["position"].(int)
		cols = append(cols, namedColumn{name: colName, typeName: colType, kind: kind, position: position})
	}

	sort.Slice(cols, func(i, j int) bool { return cols[i].position < cols[j].position })

	tc := tableContext{Table: table}
	for _, col := range cols {
		typeName := col.typeName
		if privacy == "names" {
			typeName = ""
		}
		switch col.kind {
		case "partition_key":
			tc.PartitionKeys = append(tc.PartitionKeys, col.name)
		case "clustering":
			tc.ClusteringKeys = append(tc.ClusteringKeys, col.name)
		}
		tc.Columns = append(tc.Columns, columnContext{Name: col.name, Type: typeName})
	}
	return tc
}

func (s *AIService) dryRun(ctx context.Context, session *state.Session, cql string) error {
	bounded := boundWithLimit(cql, 1)
	_, err := session.Connection.FetchPage(ctx, bounded, 1, nil)
	return err
}

func (s *AIService) AcceptExample(ctx context.Context, req *pb.AcceptExampleRequest) (*pb.AcceptExampleResponse, error) {
	session, err := GetSessionFromContext(ctx, s.store)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.Question) == "" || strings.TrimSpace(req.Cql) == "" {
		return nil, status.Error(codes.InvalidArgument, "question and cql are required")
	}

	id := s.examples.Add(session.Profile.Name, req.Question, req.Cql, req.Keyspace)
	return &pb.AcceptExampleResponse{Id: id}, nil
}

func (s *AIService) ListExamples(ctx context.Context, req *pb.ListExamplesRequest) (*pb.ListExamplesResponse, error) {
	session, err := GetSessionFromContext(ctx, s.store)
	if err != nil {
		return nil, err
	}

	limit := int(req.Limit)
	if limit <= 0 || limit > state.MaxNLExamples {
		limit = state.MaxNLExamples
	}

	entries := s.examples.List(session.Profile.Name, req.Keyspace, limit)
	out := make([]*pb.NLExample, 0, len(entries))
	for _, e := range entries {
		out = append(out, &pb.NLExample{
			Id:        e.ID,
			Question:  e.Question,
			Cql:       e.CQL,
			Keyspace:  e.Keyspace,
		})
	}
	return &pb.ListExamplesResponse{Examples: out}, nil
}

func (s *AIService) DeleteExample(ctx context.Context, req *pb.DeleteExampleRequest) (*pb.DeleteExampleResponse, error) {
	session, err := GetSessionFromContext(ctx, s.store)
	if err != nil {
		return nil, err
	}

	if !s.examples.Delete(session.Profile.Name, req.Id) {
		return nil, status.Error(codes.NotFound, "example not found")
	}
	return &pb.DeleteExampleResponse{}, nil
}

func (s *AIService) DetectProviders(ctx context.Context, req *pb.DetectProvidersRequest) (*pb.DetectProvidersResponse, error) {
	session, err := GetSessionFromContext(ctx, s.store)
	if err != nil {
		return nil, err
	}

	configured := "auto"
	model := ""
	if session.Profile.AI != nil {
		if session.Profile.AI.Provider != "" {
			configured = session.Profile.AI.Provider
		}
		model = session.Profile.AI.Model
	}

	return &pb.DetectProvidersResponse{
		Available: ai.DetectAvailable(),
		Configured: configured,
		Model:     model,
	}, nil
}

func buildSystemPrompt() string {
	return "You translate natural language questions into Cassandra CQL. " +
		"Respond with EXACTLY ONE valid SELECT statement. " +
		"No explanation, no markdown fences, no semicolons. " +
		"Always include all partition key columns in WHERE clauses with concrete equality conditions. " +
		"Only reference tables and columns that exist in the provided schema. " +
		"Use ALLOW FILTERING only when the question requires filtering on non-key columns."
}

func buildUserPrompt(question string, tables []tableContext, examples []state.NLExample, truncated bool, feedback string) string {
	now := time.Now().UTC()
	var b strings.Builder

	fmt.Fprintf(&b, "Today is %s (UTC). Resolve relative dates like 'yesterday' or 'last week' against this date.\n\n", now.Format("2006-01-02"))

	b.WriteString("Schema:\n")
	for _, t := range tables {
		fmt.Fprintf(&b, "table %s", t.Table)
		if len(t.PartitionKeys) > 0 {
			fmt.Fprintf(&b, " (partition key: %s)", strings.Join(t.PartitionKeys, ", "))
		}
		if len(t.ClusteringKeys) > 0 {
			fmt.Fprintf(&b, " (clustering: %s)", strings.Join(t.ClusteringKeys, ", "))
		}
		b.WriteString("\n")
		for _, c := range t.Columns {
			if c.Type != "" {
				fmt.Fprintf(&b, "  %s %s\n", c.Name, c.Type)
			} else {
				fmt.Fprintf(&b, "  %s\n", c.Name)
			}
		}
	}

	if truncated {
		fmt.Fprintf(&b, "\nNote: only the first %d tables (alphabetically) are listed. If the question references a table not listed, say so instead of guessing.\n", maxPromptTables)
	}

	if len(examples) > 0 {
		b.WriteString("\nExample translations:\n")
		for _, e := range examples {
			fmt.Fprintf(&b, "Q: %s\nCQL: %s\n", e.Question, e.CQL)
		}
	}

	if feedback != "" {
		fmt.Fprintf(&b, "\nYour previous attempt failed with: %s\nFix it and respond again with a single corrected SELECT.\n", feedback)
	}

	fmt.Fprintf(&b, "\nQuestion: %s\n", question)
	return b.String()
}

func stripFences(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "```cql")
	trimmed = strings.TrimPrefix(trimmed, "```sql")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	return strings.TrimSpace(trimmed)
}

func boundWithLimit(cql string, limit int) string {
	if !strings.Contains(strings.ToUpper(cql), " LIMIT ") {
		return fmt.Sprintf("%s LIMIT %d", cql, limit)
	}
	return cql
}
