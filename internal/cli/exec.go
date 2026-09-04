package cli

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	pb "github.com/KashifKhn/kassie/api/gen/go"
	"github.com/KashifKhn/kassie/internal/client"
	"github.com/KashifKhn/kassie/internal/server"
	"github.com/KashifKhn/kassie/internal/shared/logger"
	"github.com/spf13/cobra"
)

var (
	execProfile string
	execFormat  string
	execOutput  string
	execDryRun  bool
	execServer  string
)

func newExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec [file]",
		Short: "Execute a CQL script against a profile (read-only SELECT)",
		Long: `Execute a CQL script file (or stdin when file is '-') against a database profile.

Statements are separated by semicolons; comments (-- and /* */) are ignored.
Only read-only SELECT statements are allowed, validated server-side.

Output formats: table (default, human), json (rows per statement),
csv (one csv stream per statement, header row included).

Exit codes: 0 success, 1 query execution error, 2 validation failure.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runExec,
	}

	cmd.Flags().StringVarP(&execProfile, "profile", "p", "", "database profile to use")
	cmd.Flags().StringVarP(&execFormat, "format", "f", "table", "output format: table, json, csv")
	cmd.Flags().StringVarP(&execOutput, "output", "o", "", "write output to file instead of stdout")
	cmd.Flags().BoolVar(&execDryRun, "dry-run", false, "parse and validate statements without executing")
	cmd.Flags().StringVar(&execServer, "server", "", "kassie gRPC server address (starts embedded when empty)")

	return cmd
}

func runExec(cmd *cobra.Command, args []string) error {
	if execProfile == "" {
		profile, err := appConfig.GetDefaultProfile()
		if err != nil {
			return fmt.Errorf("no profile specified and no default configured: %w", err)
		}
		execProfile = profile.Name
	}

	script, err := readScript(args)
	if err != nil {
		return err
	}

	statements := splitStatements(script)
	if len(statements) == 0 {
		return fmt.Errorf("no statements found in script")
	}

	grpcAddr := execServer
	if grpcAddr == "" {
		quietLogger, err := logger.New(logger.Config{Level: logger.ErrorLevel, Pretty: false, Output: io.Discard})
		if err != nil {
			return err
		}
		appLogger = quietLogger
		jwtSecret := os.Getenv("KASSIE_JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = generateSecret()
		}

		embedded, err := server.NewEmbeddedServer(appConfig, &server.EmbeddedServerConfig{
			JWTSecret: jwtSecret,
			GRPCPort:  0,
			HTTPPort:  0,
		}, appLogger)
		if err != nil {
			return fmt.Errorf("failed to create embedded server: %w", err)
		}
		if err := embedded.Start(); err != nil {
			return fmt.Errorf("failed to start embedded server: %w", err)
		}
		defer embedded.Stop()
		grpcAddr = embedded.GRPCAddress()
	}

	conn, err := client.New(grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer func() { _ = conn.Close() }()

	ctxLogin, cancelLogin := context.WithTimeout(context.Background(), 10*time.Second)
	_, err = conn.Login(ctxLogin, execProfile)
	cancelLogin()
	if err != nil {
		return fmt.Errorf("failed to login with profile %s: %w", execProfile, err)
	}

	out := io.Writer(os.Stdout)
	if execOutput != "" {
		f, err := os.Create(execOutput)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer func() { _ = f.Close() }()
		out = f
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for i, stmt := range statements {
		if execDryRun {
			fmt.Fprintf(out, "-- [dry-run] statement %d: %s\n", i+1, preview(stmt))
			continue
		}

		resp, err := conn.ExecuteQuery(sigCtx, stmt, 1000)
		if err != nil {
			if strings.Contains(err.Error(), "only SELECT") || strings.Contains(err.Error(), "multiple statements") || strings.Contains(err.Error(), "disallowed keywords") {
				fmt.Fprintf(os.Stderr, "validation error on statement %d: %v\n", i+1, err)
				os.Exit(2)
			}
			fmt.Fprintf(os.Stderr, "query error on statement %d: %v\n", i+1, err)
			os.Exit(1)
		}

		if err := writeStatementOutput(out, i, stmt, resp); err != nil {
			return err
		}
	}

	return nil
}

func readScript(args []string) (string, error) {
	if len(args) == 0 || args[0] == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read stdin: %w", err)
		}
		return string(data), nil
	}

	data, err := os.ReadFile(args[0])
	if err != nil {
		return "", fmt.Errorf("failed to read script: %w", err)
	}
	return string(data), nil
}

func splitStatements(script string) []string {
	var statements []string
	var current strings.Builder
	inLineComment := false
	inBlockComment := false
	inString := false

	flush := func() {
		stmt := strings.TrimSpace(current.String())
		stmt = strings.TrimRight(stmt, ";")
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			statements = append(statements, stmt)
		}
		current.Reset()
	}

	runes := []rune(script)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		next := rune(0)
		if i+1 < len(runes) {
			next = runes[i+1]
		}

		switch {
		case inLineComment:
			if r == '\n' {
				inLineComment = false
			}
			continue
		case inBlockComment:
			if r == '*' && next == '/' {
				inBlockComment = false
				i++
			}
			continue
		case inString:
			if r == '\'' {
				inString = false
			}
			current.WriteRune(r)
			continue
		case r == '-' && next == '-':
			inLineComment = true
			continue
		case r == '/' && next == '*':
			inBlockComment = true
			i++
			continue
		case r == '\'':
			inString = true
			current.WriteRune(r)
			continue
		case r == ';':
			flush()
			continue
		}

		current.WriteRune(r)
	}
	flush()

	return statements
}

func preview(stmt string) string {
	one := strings.Join(strings.Fields(stmt), " ")
	if len(one) > 72 {
		one = one[:69] + "..."
	}
	return one
}

func writeStatementOutput(out io.Writer, index int, stmt string, resp *pb.ExecuteQueryResponse) error {
	switch execFormat {
	case "json":
		rows := make([]map[string]interface{}, 0, len(resp.Rows))
		for _, row := range resp.Rows {
			obj := make(map[string]interface{}, len(row.Cells))
			for k, cell := range row.Cells {
				obj[k] = cellJSON(cell)
			}
			rows = append(rows, obj)
		}
		data, err := json.MarshalIndent(map[string]interface{}{
			"statement":     preview(stmt),
			"total_fetched": resp.TotalFetched,
			"rows":          rows,
		}, "", "  ")
		if err != nil {
			return err
		}
		_, err = out.Write(append(data, '\n'))
		return err

	case "csv":
		w := csv.NewWriter(out)
		var header []string
		for _, row := range resp.Rows {
			if header == nil && len(row.Cells) > 0 {
				for _, col := range columnOrder(row) {
					header = append(header, col)
				}
				if err := w.Write(header); err != nil {
					return err
				}
			}
			record := make([]string, len(header))
			for i, col := range header {
				record[i] = cellString(row.Cells[col])
			}
			if err := w.Write(record); err != nil {
				return err
			}
		}
		w.Flush()
		return w.Error()

	default:
		fmt.Fprintf(out, "-- statement %d: %s\n", index+1, preview(stmt))
		fmt.Fprintf(out, "-- %d rows\n", resp.TotalFetched)

		if len(resp.Rows) == 0 {
			fmt.Fprintln(out, "(no rows)")
			return nil
		}

		header := columnOrder(resp.Rows[0])
		rows := make([][]string, 0, len(resp.Rows))
		rows = append(rows, header)
		for _, row := range resp.Rows {
			record := make([]string, len(header))
			for i, col := range header {
				record[i] = cellString(row.Cells[col])
			}
			rows = append(rows, record)
		}

		widths := make([]int, len(header))
		for _, record := range rows {
			for i, value := range record {
				if len(value) > widths[i] {
					widths[i] = len(value)
				}
			}
		}
		for i := range widths {
			if widths[i] > 40 {
				widths[i] = 40
			}
		}

		for _, record := range rows {
			line := "| "
			for i, value := range record {
				padded := value
				if len(padded) > widths[i] {
					padded = padded[:widths[i]-1] + "…"
				}
				line += pad(padded, widths[i]) + " | "
			}
			fmt.Fprintln(out, line)
		}
		return nil
	}
}

func columnOrder(row *pb.Row) []string {
	cols := make([]string, 0, len(row.Cells))
	for col := range row.Cells {
		cols = append(cols, col)
	}
	sortStrings(cols)
	return cols
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

func pad(s string, width int) string {
	for len(s) < width {
		s += " "
	}
	return s
}

func cellJSON(cell *pb.CellValue) interface{} {
	if cell == nil || cell.IsNull {
		return nil
	}
	switch v := cell.Value.(type) {
	case *pb.CellValue_StringVal:
		return v.StringVal
	case *pb.CellValue_IntVal:
		return v.IntVal
	case *pb.CellValue_DoubleVal:
		return v.DoubleVal
	case *pb.CellValue_BoolVal:
		return v.BoolVal
	case *pb.CellValue_BytesVal:
		return fmt.Sprintf("0x%x", v.BytesVal)
	default:
		return nil
	}
}

func cellString(cell *pb.CellValue) string {
	if cell == nil || cell.IsNull {
		return "null"
	}
	switch v := cell.Value.(type) {
	case *pb.CellValue_StringVal:
		return v.StringVal
	case *pb.CellValue_IntVal:
		return fmt.Sprintf("%d", v.IntVal)
	case *pb.CellValue_DoubleVal:
		return fmt.Sprintf("%g", v.DoubleVal)
	case *pb.CellValue_BoolVal:
		return fmt.Sprintf("%t", v.BoolVal)
	case *pb.CellValue_BytesVal:
		return fmt.Sprintf("0x%x", v.BytesVal)
	default:
		return ""
	}
}
