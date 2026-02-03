package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid debug level",
			cfg: Config{
				Level:  DebugLevel,
				Pretty: false,
			},
			wantErr: false,
		},
		{
			name: "valid info level",
			cfg: Config{
				Level:  InfoLevel,
				Pretty: false,
			},
			wantErr: false,
		},
		{
			name: "valid warn level",
			cfg: Config{
				Level:  WarnLevel,
				Pretty: false,
			},
			wantErr: false,
		},
		{
			name: "valid error level",
			cfg: Config{
				Level:  ErrorLevel,
				Pretty: false,
			},
			wantErr: false,
		},
		{
			name: "invalid level",
			cfg: Config{
				Level:  "invalid",
				Pretty: false,
			},
			wantErr: true,
		},
		{
			name: "pretty output",
			cfg: Config{
				Level:  InfoLevel,
				Pretty: true,
			},
			wantErr: false,
		},
		{
			name: "custom output",
			cfg: Config{
				Level:  InfoLevel,
				Pretty: false,
				Output: &bytes.Buffer{},
			},
			wantErr: false,
		},
		{
			name: "custom time format",
			cfg: Config{
				Level:      InfoLevel,
				Pretty:     false,
				TimeFormat: "2006-01-02",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && logger == nil {
				t.Error("New() returned nil logger")
			}

			if !tt.wantErr {
				if logger.cfg.Level != tt.cfg.Level {
					t.Errorf("New() level = %v, want %v", logger.cfg.Level, tt.cfg.Level)
				}
			}
		})
	}
}

func TestDefault(t *testing.T) {
	logger := Default()

	if logger == nil {
		t.Fatal("Default() returned nil logger")
	}

	if logger.cfg.Level != InfoLevel {
		t.Errorf("Default() level = %v, want %v", logger.cfg.Level, InfoLevel)
	}

	if logger.cfg.Pretty != false {
		t.Error("Default() pretty should be false")
	}
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		logFunc  func(*Logger, *bytes.Buffer)
		wantLog  bool
		wantText string
	}{
		{
			name:  "debug level logs debug",
			level: DebugLevel,
			logFunc: func(l *Logger, buf *bytes.Buffer) {
				l.Debug("test message")
			},
			wantLog:  true,
			wantText: "debug",
		},
		{
			name:  "info level does not log debug",
			level: InfoLevel,
			logFunc: func(l *Logger, buf *bytes.Buffer) {
				l.Debug("test message")
			},
			wantLog: false,
		},
		{
			name:  "info level logs info",
			level: InfoLevel,
			logFunc: func(l *Logger, buf *bytes.Buffer) {
				l.Info("test message")
			},
			wantLog:  true,
			wantText: "info",
		},
		{
			name:  "warn level does not log info",
			level: WarnLevel,
			logFunc: func(l *Logger, buf *bytes.Buffer) {
				l.Info("test message")
			},
			wantLog: false,
		},
		{
			name:  "warn level logs warn",
			level: WarnLevel,
			logFunc: func(l *Logger, buf *bytes.Buffer) {
				l.Warn("test message")
			},
			wantLog:  true,
			wantText: "warn",
		},
		{
			name:  "error level does not log warn",
			level: ErrorLevel,
			logFunc: func(l *Logger, buf *bytes.Buffer) {
				l.Warn("test message")
			},
			wantLog: false,
		},
		{
			name:  "error level logs error",
			level: ErrorLevel,
			logFunc: func(l *Logger, buf *bytes.Buffer) {
				l.Error("test message")
			},
			wantLog:  true,
			wantText: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger, err := New(Config{
				Level:  tt.level,
				Pretty: false,
				Output: buf,
			})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			tt.logFunc(logger, buf)

			output := buf.String()
			hasLog := len(output) > 0

			if hasLog != tt.wantLog {
				t.Errorf("Log output present = %v, want %v", hasLog, tt.wantLog)
			}

			if tt.wantLog && !strings.Contains(output, tt.wantText) {
				t.Errorf("Log output missing expected text %q, got: %s", tt.wantText, output)
			}

			if tt.wantLog && !strings.Contains(output, "test message") {
				t.Errorf("Log output missing message, got: %s", output)
			}
		})
	}
}

func TestLogFormatted(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func(*Logger)
		wantText string
	}{
		{
			name: "debugf",
			logFunc: func(l *Logger) {
				l.Debugf("formatted %s %d", "message", 42)
			},
			wantText: "formatted message 42",
		},
		{
			name: "infof",
			logFunc: func(l *Logger) {
				l.Infof("formatted %s %d", "message", 42)
			},
			wantText: "formatted message 42",
		},
		{
			name: "warnf",
			logFunc: func(l *Logger) {
				l.Warnf("formatted %s %d", "message", 42)
			},
			wantText: "formatted message 42",
		},
		{
			name: "errorf",
			logFunc: func(l *Logger) {
				l.Errorf("formatted %s %d", "message", 42)
			},
			wantText: "formatted message 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger, err := New(Config{
				Level:  DebugLevel,
				Pretty: false,
				Output: buf,
			})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			tt.logFunc(logger)

			output := buf.String()
			if !strings.Contains(output, tt.wantText) {
				t.Errorf("Log output missing expected text %q, got: %s", tt.wantText, output)
			}
		})
	}
}

func TestWith(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, err := New(Config{
		Level:  InfoLevel,
		Pretty: false,
		Output: buf,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	testErr := fmt.Errorf("context error")
	contextLogger := logger.With().Str("service", "test").Int("version", 1).Bool("enabled", true).Err(testErr).Logger()
	contextLogger.Info("test message")

	output := buf.String()

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if logEntry["service"] != "test" {
		t.Errorf("Expected service=test, got %v", logEntry["service"])
	}

	if logEntry["version"] != float64(1) {
		t.Errorf("Expected version=1, got %v", logEntry["version"])
	}

	if logEntry["enabled"] != true {
		t.Errorf("Expected enabled=true, got %v", logEntry["enabled"])
	}

	if logEntry["error"] != "context error" {
		t.Errorf("Expected error='context error', got %v", logEntry["error"])
	}

	if logEntry["message"] != "test message" {
		t.Errorf("Expected message='test message', got %v", logEntry["message"])
	}
}

func TestWithError(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, err := New(Config{
		Level:  InfoLevel,
		Pretty: false,
		Output: buf,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	testErr := fmt.Errorf("test error")
	contextLogger := logger.WithError(testErr).Logger()
	contextLogger.Error("operation failed")

	output := buf.String()

	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if logEntry["error"] != "test error" {
		t.Errorf("Expected error='test error', got %v", logEntry["error"])
	}

	if logEntry["message"] != "operation failed" {
		t.Errorf("Expected message='operation failed', got %v", logEntry["message"])
	}
}

func TestSetLevel(t *testing.T) {
	tests := []struct {
		name      string
		initLevel Level
		newLevel  Level
		wantErr   bool
	}{
		{
			name:      "change from info to debug",
			initLevel: InfoLevel,
			newLevel:  DebugLevel,
			wantErr:   false,
		},
		{
			name:      "change from debug to error",
			initLevel: DebugLevel,
			newLevel:  ErrorLevel,
			wantErr:   false,
		},
		{
			name:      "invalid new level",
			initLevel: InfoLevel,
			newLevel:  "invalid",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(Config{
				Level:  tt.initLevel,
				Pretty: false,
			})
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			err = logger.SetLevel(tt.newLevel)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && logger.GetLevel() != tt.newLevel {
				t.Errorf("GetLevel() = %v, want %v", logger.GetLevel(), tt.newLevel)
			}
		})
	}
}

func TestSetLevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, err := New(Config{
		Level:  InfoLevel,
		Pretty: false,
		Output: buf,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Debug("should not appear")
	if buf.Len() > 0 {
		t.Error("Debug message logged at Info level")
	}

	if err := logger.SetLevel(DebugLevel); err != nil {
		t.Fatalf("SetLevel() error = %v", err)
	}

	logger.Debug("should appear")
	if buf.Len() == 0 {
		t.Error("Debug message not logged after changing to Debug level")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Level
		wantErr bool
	}{
		{
			name:    "debug lowercase",
			input:   "debug",
			want:    DebugLevel,
			wantErr: false,
		},
		{
			name:    "info uppercase",
			input:   "INFO",
			want:    InfoLevel,
			wantErr: false,
		},
		{
			name:    "warn mixed case",
			input:   "WaRn",
			want:    WarnLevel,
			wantErr: false,
		},
		{
			name:    "error with whitespace",
			input:   "  error  ",
			want:    ErrorLevel,
			wantErr: false,
		},
		{
			name:    "invalid level",
			input:   "invalid",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLevel(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("ParseLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateLevel(t *testing.T) {
	tests := []struct {
		name    string
		level   Level
		wantErr bool
	}{
		{
			name:    "valid debug",
			level:   DebugLevel,
			wantErr: false,
		},
		{
			name:    "valid info",
			level:   InfoLevel,
			wantErr: false,
		},
		{
			name:    "valid warn",
			level:   WarnLevel,
			wantErr: false,
		},
		{
			name:    "valid error",
			level:   ErrorLevel,
			wantErr: false,
		},
		{
			name:    "invalid level",
			level:   "invalid",
			wantErr: true,
		},
		{
			name:    "empty level",
			level:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLevel(tt.level)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateLevel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJSONOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, err := New(Config{
		Level:  InfoLevel,
		Pretty: false,
		Output: buf,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Info("test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if logEntry["level"] != "info" {
		t.Errorf("Expected level=info, got %v", logEntry["level"])
	}

	if logEntry["message"] != "test message" {
		t.Errorf("Expected message='test message', got %v", logEntry["message"])
	}

	if _, ok := logEntry["time"]; !ok {
		t.Error("Expected timestamp in log entry")
	}
}
