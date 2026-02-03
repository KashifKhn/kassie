package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Level string

const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
)

var (
	ErrInvalidLevel = fmt.Errorf("invalid log level")
)

type Config struct {
	Level      Level
	Pretty     bool
	Output     io.Writer
	TimeFormat string
}

type Logger struct {
	zlog zerolog.Logger
	cfg  Config
}

func New(cfg Config) (*Logger, error) {
	if err := validateLevel(cfg.Level); err != nil {
		return nil, err
	}

	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}

	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}

	var writer io.Writer
	if cfg.Pretty {
		writer = zerolog.ConsoleWriter{
			Out:        cfg.Output,
			TimeFormat: cfg.TimeFormat,
		}
	} else {
		writer = cfg.Output
	}

	level := parseLevel(cfg.Level)
	zlog := zerolog.New(writer).Level(level).With().Timestamp().Logger()

	return &Logger{
		zlog: zlog,
		cfg:  cfg,
	}, nil
}

func Default() *Logger {
	logger, _ := New(Config{
		Level:  InfoLevel,
		Pretty: false,
		Output: os.Stderr,
	})
	return logger
}

func (l *Logger) Debug(msg string) {
	l.zlog.Debug().Msg(msg)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.zlog.Debug().Msgf(format, args...)
}

func (l *Logger) Info(msg string) {
	l.zlog.Info().Msg(msg)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.zlog.Info().Msgf(format, args...)
}

func (l *Logger) Warn(msg string) {
	l.zlog.Warn().Msg(msg)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.zlog.Warn().Msgf(format, args...)
}

func (l *Logger) Error(msg string) {
	l.zlog.Error().Msg(msg)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.zlog.Error().Msgf(format, args...)
}

func (l *Logger) With() *Event {
	return &Event{event: l.zlog.With()}
}

func (l *Logger) WithError(err error) *Event {
	return &Event{event: l.zlog.With().Err(err)}
}

func (l *Logger) SetLevel(level Level) error {
	if err := validateLevel(level); err != nil {
		return err
	}
	l.cfg.Level = level
	l.zlog = l.zlog.Level(parseLevel(level))
	return nil
}

func (l *Logger) GetLevel() Level {
	return l.cfg.Level
}

type Event struct {
	event zerolog.Context
}

func (e *Event) Str(key, val string) *Event {
	e.event = e.event.Str(key, val)
	return e
}

func (e *Event) Int(key string, val int) *Event {
	e.event = e.event.Int(key, val)
	return e
}

func (e *Event) Bool(key string, val bool) *Event {
	e.event = e.event.Bool(key, val)
	return e
}

func (e *Event) Err(err error) *Event {
	e.event = e.event.Err(err)
	return e
}

func (e *Event) Logger() *Logger {
	return &Logger{zlog: e.event.Logger()}
}

func parseLevel(level Level) zerolog.Level {
	switch level {
	case DebugLevel:
		return zerolog.DebugLevel
	case InfoLevel:
		return zerolog.InfoLevel
	case WarnLevel:
		return zerolog.WarnLevel
	case ErrorLevel:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

func validateLevel(level Level) error {
	switch level {
	case DebugLevel, InfoLevel, WarnLevel, ErrorLevel:
		return nil
	default:
		return ErrInvalidLevel
	}
}

func ParseLevel(s string) (Level, error) {
	level := Level(strings.ToLower(strings.TrimSpace(s)))
	if err := validateLevel(level); err != nil {
		return "", err
	}
	return level, nil
}
