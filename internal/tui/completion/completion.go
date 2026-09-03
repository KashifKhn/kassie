package completion

import (
	"strings"
	"unicode/utf8"
)

type SuggestionKind int

const (
	KindKeyword SuggestionKind = iota
	KindKeyspace
	KindTable
	KindColumn
)

func (k SuggestionKind) String() string {
	switch k {
	case KindKeyword:
		return "keyword"
	case KindKeyspace:
		return "keyspace"
	case KindTable:
		return "table"
	case KindColumn:
		return "column"
	default:
		return ""
	}
}

type Suggestion struct {
	Label  string
	Detail string
	Kind   SuggestionKind
}

var cqlKeywords = []string{
	"SELECT", "FROM", "WHERE", "AND", "OR", "IN", "CONTAINS", "LIMIT",
	"ALLOW", "FILTERING", "ORDER", "BY", "ASC", "DESC", "NULL", "TOKEN",
	"COUNT", "MIN", "MAX", "SUM", "AVG", "DISTINCT", "AS", "JSON",
	"PER", "PARTITION",
}

// Context describes the current token position in a SELECT statement.
type Context int

const (
	ContextStart Context = iota
	ContextAfterSelect
	ContextAfterFrom
	ContextAfterWhere
	ContextColumnRef
	ContextTableRef
	ContextOther
)

// Completion sources supplied by the caller per activation.
type Sources struct {
	Keyspaces  []string
	TablesFor  func(keyspace string) []string
	ColumnsFor func(keyspace, table string) []Column
	// DefaultKeyspace/DefaultTable are the UI's current selection, used
	// to resolve columns before an explicit FROM clause is typed.
	DefaultKeyspace string
	DefaultTable    string
}

type Column struct {
	Name    string
	CqlType string
}

// Complete returns ranked suggestions for the text typed so far, using
// word-start matching against context-appropriate candidates.
// text is the full editor content; the token being completed ends at the end of text.
func Complete(text string, sources Sources) []Suggestion {
	prefix, ctx := Analyze(text)

	switch ctx {
	case ContextStart:
		return filterByPrefix(cqlKeywordSuggestions(), prefix)
	case ContextAfterSelect:
		return filterByPrefix(cqlKeywordSuggestions(), prefix)
	case ContextAfterFrom, ContextTableRef:
		return tableSuggestions(text, prefix, sources)
	case ContextAfterWhere, ContextColumnRef:
		return columnSuggestions(text, prefix, sources)
	default:
		return filterByPrefix(cqlKeywordSuggestions(), prefix)
	}
}

// Analyze returns the word being completed and the semantic context.
func Analyze(text string) (string, Context) {
	if strings.TrimSpace(text) == "" {
		return "", ContextStart
	}

	// The original text's tail decides whether we are mid-word.
	lastRune, _ := utf8.DecodeLastRuneInString(text)
	endsMidWord := isWordChar(lastRune)

	trimmedRight := strings.TrimRight(text, " \t\n\r")
	tokens := tokenize(trimmedRight)
	if len(tokens) == 0 {
		return "", ContextStart
	}

	// Cursor right after "ks." — completing a table.
	if strings.HasSuffix(trimmedRight, ".") {
		return "", ContextTableRef
	}

	if !endsMidWord {
		// Cursor after a complete word + space: fresh token, context from history
		return "", contextFromTokens(tokens)
	}

	// Mid-word: prefix is the trailing run of word chars.
	prefix := trailingWord(trimmedRight)
	history := tokens[:len(tokens)-1]
	if len(history) == 0 {
		return prefix, ContextStart
	}

	if strings.HasSuffix(trimmedRight, "."+prefix) {
		return prefix, ContextTableRef
	}

	return prefix, contextFromTokens(history)
}

// contextFromTokens infers context from the token history; when midWord is
// true the next token is being typed (so "SELECT id F" → after-SELECT with
// prefix F against FROM keyword context).
func contextFromTokens(tokens []token) Context {
	if len(tokens) == 0 {
		return ContextStart
	}

	last := strings.ToUpper(tokens[len(tokens)-1].word)
	secondLast := ""
	if len(tokens) >= 2 {
		secondLast = strings.ToUpper(tokens[len(tokens)-2].word)
	}

	switch {
	case last == "SELECT":
		return ContextColumnRef
	case last == "FROM":
		return ContextTableRef
	case last == "WHERE" || last == "AND" || last == "OR":
		return ContextColumnRef
	case last == "BY":
		return ContextColumnRef
	case secondLast == "FROM" && last == ",":
		return ContextTableRef
	}

	if hasFrom(tokens) {
		if hasWhere(tokens) {
			return ContextColumnRef
		}
		return ContextOther
	}

	// Still typing the SELECT list (no FROM yet): columns of the default table.
	if len(tokens) > 0 && strings.EqualFold(tokens[0].word, "SELECT") {
		return ContextColumnRef
	}

	return ContextOther
}

func isWordChar(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '.', ',', '(', ')', '=', '<', '>', '*', ';', '\'':
		return false
	}
	return true
}

func trailingWord(text string) string {
	start := len(text)
	for start > 0 && isWordChar(rune(text[start-1])) {
		start--
	}
	return text[start:]
}

func hasFrom(tokens []token) bool {
	for _, t := range tokens {
		if strings.EqualFold(t.word, "FROM") {
			return true
		}
	}
	return false
}

func hasWhere(tokens []token) bool {
	for _, t := range tokens {
		if strings.EqualFold(t.word, "WHERE") {
			return true
		}
	}
	return false
}

type token struct {
	word string
}

func tokenize(text string) []token {
	var tokens []token
	var current strings.Builder
	inWord := false

	for _, r := range text {
		switch {
		case r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '.' || r == ',' || r == '(' || r == ')' || r == '=' || r == '<' || r == '>' || r == '*' || r == ';' || r == '\'':
			if inWord {
				tokens = append(tokens, token{word: current.String()})
				current.Reset()
				inWord = false
			}
		default:
			current.WriteRune(r)
			inWord = true
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, token{word: current.String()})
	}

	return tokens
}

func tableSuggestions(text, prefix string, sources Sources) []Suggestion {
	var suggestions []Suggestion

	// keyspace.table typing
	if idx := strings.LastIndex(text, "."); idx >= 0 {
		ks := strings.TrimSpace(text[maxInt(0, findFromEnd(text, idx)):idx])
		if ks == "" {
			// leading dot: nothing sensible
			return nil
		}
		tablePrefix := text[idx+1:]
		for _, tbl := range sources.TablesFor(ks) {
			if strings.HasPrefix(strings.ToLower(tbl), strings.ToLower(tablePrefix)) || tablePrefix == "" {
				suggestions = append(suggestions, Suggestion{
					Label: ks + "." + tbl, Kind: KindTable,
				})
			}
		}
		return suggestions
	}

	// bare prefix: keyspaces + keyword FROM-relevant
	for _, ks := range sources.Keyspaces {
		if matchWord(prefix, ks) {
			suggestions = append(suggestions, Suggestion{Label: ks, Kind: KindKeyspace})
		}
	}

	// tables of the currently selected keyspace are handled by caller
	// via TablesFor("") if provided
	for _, tbl := range sources.TablesFor("") {
		if matchWord(prefix, tbl) {
			suggestions = append(suggestions, Suggestion{Label: tbl, Kind: KindTable})
		}
	}

	return suggestions
}

func columnSuggestions(text, prefix string, sources Sources) []Suggestion {
	ks, tbl := resolveTargetTable(text)
	if ks == "" || tbl == "" {
		ks, tbl = sources.DefaultKeyspace, sources.DefaultTable
	}

	var suggestions []Suggestion
	if ks != "" && tbl != "" {
		for _, col := range sources.ColumnsFor(ks, tbl) {
			if matchWord(prefix, col.Name) {
				suggestions = append(suggestions, Suggestion{
					Label: col.Name, Detail: col.CqlType, Kind: KindColumn,
				})
			}
		}
	}

	// WHERE context also allows keywords (IN, AND, ...)
	for _, kw := range cqlKeywords {
		if matchWord(prefix, kw) && (kw == "AND" || kw == "OR" || kw == "IN" || kw == "CONTAINS") {
			suggestions = append(suggestions, Suggestion{Label: kw, Kind: KindKeyword})
		}
	}

	if len(suggestions) == 0 {
		for _, kw := range []string{"AND", "OR", "IN", "CONTAINS", "LIMIT", "ORDER", "ALLOW", "FILTERING"} {
			if matchWord(prefix, kw) {
				suggestions = append(suggestions, Suggestion{Label: kw, Kind: KindKeyword})
			}
		}
	}

	return suggestions
}

// resolveTargetTable finds "ks.table" in FROM clause of text.
func resolveTargetTable(text string) (string, string) {
	upper := strings.ToUpper(text)
	fromIdx := strings.Index(upper, " FROM ")
	if fromIdx < 0 {
		if strings.HasPrefix(strings.TrimSpace(upper), "FROM ") {
			fromIdx = -1
		} else {
			return "", ""
		}
	}

	rest := text[fromIdx+len(" FROM "):]
	if fromIdx < 0 {
		rest = strings.TrimSpace(text)[len("FROM "):]
	}
	// cut at WHERE
	if wIdx := strings.Index(strings.ToUpper(rest), " WHERE "); wIdx >= 0 {
		rest = rest[:wIdx]
	}
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return "", ""
	}

	// strip alias / trailing tokens
	fields := strings.Fields(rest)
	target := fields[0]
	target = strings.TrimSuffix(target, ";")

	if idx := strings.Index(target, "."); idx >= 0 {
		ks := strings.Trim(target[:idx], "\"")
		tbl := strings.Trim(target[idx+1:], "\"")
		return ks, tbl
	}
	return "", strings.Trim(target, "\"")
}

func findFromEnd(text string, dotIdx int) int {
	for i := dotIdx - 1; i >= 0; i-- {
		r := rune(text[i])
		if r == ' ' || r == '\t' || r == '\n' {
			return i + 1
		}
	}
	return 0
}

func matchWord(prefix, candidate string) bool {
	return strings.HasPrefix(strings.ToLower(candidate), strings.ToLower(prefix))
}

func cqlKeywordSuggestions() []Suggestion {
	s := make([]Suggestion, 0, len(cqlKeywords))
	for _, kw := range cqlKeywords {
		s = append(s, Suggestion{Label: kw, Kind: KindKeyword})
	}
	return s
}

func filterByPrefix(suggestions []Suggestion, prefix string) []Suggestion {
	if prefix == "" {
		return suggestions
	}
	var out []Suggestion
	for _, s := range suggestions {
		if strings.HasPrefix(strings.ToLower(s.Label), strings.ToLower(prefix)) {
			out = append(out, s)
		}
	}
	return out
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
