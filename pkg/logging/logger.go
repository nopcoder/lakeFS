package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/lmittmann/tint"
	"gopkg.in/natefinch/lumberjack.v2"
)

type contextKey string

const (
	LogFieldsContextKey = contextKey("log_fields")

	ProjectDirectoryName = "lakefs"

	RepositoryFieldKey      = "repository"
	MatchedHostFieldKey     = "matched_host"
	RefHostFieldKey         = "ref"
	PathFieldKey            = "path"
	UploadIDFieldKey        = "upload_id"
	ListTypeFieldKey        = "list_type"
	PhysicalAddressFieldKey = "physical_address"
	PartNumberFieldKey      = "part_number"
	RequestIDFieldKey       = "request_id"
	HostFieldKey            = "host"
	MethodFieldKey          = "method"
	UserFieldKey            = "user"
	ServiceNameFieldKey     = "service_name"
	LogAudit                = "log_audit"
)

type Fields map[string]any

var (
	defaultLogger *slog.Logger
	openLoggers   []io.Closer
	initOnce      sync.Once
	currentLevel  slog.Level
	currentFormat string
	loggerMu      sync.RWMutex
)

func getLogger() *slog.Logger {
	initOnce.Do(func() {
		defaultLogger = newHandlerLogger(slog.LevelInfo, "text")
		currentLevel = slog.LevelInfo
		currentFormat = "text"
	})
	return defaultLogger
}

func newHandlerLogger(level slog.Level, format string) *slog.Logger {
	var handler slog.Handler

	output := os.Stderr

	switch strings.ToLower(format) {
	case "text":
		handler = tint.NewHandler(output, &tint.Options{
			Level:       level,
			AddSource:   true,
			ReplaceAttr: trimSourceAttr,
		})
	case "json":
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level:     level,
			AddSource: true,
		})
	default:
		handler = tint.NewHandler(output, &tint.Options{
			Level:       level,
			AddSource:   true,
			ReplaceAttr: trimSourceAttr,
		})
	}

	return slog.New(handler)
}

func trimSourceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "source" {
		if v := a.Value.String(); v != "" {
			idx := findLakeFSPath(v)
			if idx > 0 {
				return slog.Attr{Key: "source", Value: slog.StringValue(v[idx:])}
			}
		}
	}
	return a
}

func findLakeFSPath(s string) int {
	lower := strings.ToLower(s)
	idx := strings.Index(lower, ProjectDirectoryName)
	if idx == -1 {
		return 0
	}
	remaining := s[idx+len(ProjectDirectoryName):]
	sepIdx := strings.Index(remaining, "/")
	if sepIdx == -1 {
		return 0
	}
	return idx + len(ProjectDirectoryName) + sepIdx
}

func Level() string {
	loggerMu.RLock()
	level := currentLevel
	loggerMu.RUnlock()

	switch level {
	case slog.LevelDebug:
		return "debug"
	case slog.LevelInfo:
		return "info"
	case slog.LevelWarn:
		return "warn"
	case slog.LevelError:
		return "error"
	default:
		return level.String()
	}
}

func SetLevel(level string) {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "trace", "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	case "panic":
		lvl = slog.LevelError + 1
	case "null", "none":
		defaultLogger = slog.New(slog.NewTextHandler(io.Discard, nil))
		loggerMu.Lock()
		currentLevel = slog.LevelError + 1
		loggerMu.Unlock()
		return
	default:
		lvl = slog.LevelInfo
	}

	loggerMu.Lock()
	defer loggerMu.Unlock()

	defaultLogger = newHandlerLogger(lvl, currentFormat)
	currentLevel = lvl
}

func CloseWriters() error {
	for _, c := range openLoggers {
		if err := c.Close(); err != nil {
			return err
		}
	}
	openLoggers = nil
	return nil
}

func SetOutputs(outputs []string, fileMaxSizeMB, filesKeep int) error {
	var writers []io.Writer
	if err := CloseWriters(); err != nil {
		return err
	}
	for _, output := range outputs {
		var w io.Writer
		switch output {
		case "":
			continue
		case "-":
			w = os.Stdout
		case "=":
			w = os.Stderr
		default:
			l := &lumberjack.Logger{
				Filename:   output,
				MaxSize:    fileMaxSizeMB,
				MaxBackups: filesKeep,
			}
			w = l
			openLoggers = append(openLoggers, l)
		}
		writers = append(writers, w)
	}

	loggerMu.Lock()
	defer loggerMu.Unlock()

	var handler slog.Handler
	if len(writers) == 1 {
		handler = newHandlerFromWriter(writers[0], currentLevel, currentFormat)
	} else if len(writers) > 1 {
		handler = newHandlerFromWriter(io.MultiWriter(writers...), currentLevel, currentFormat)
	} else {
		handler = newHandlerFromWriter(os.Stderr, currentLevel, currentFormat)
	}

	defaultLogger = slog.New(handler)
	return nil
}

func newHandlerFromWriter(w io.Writer, level slog.Level, format string) slog.Handler {
	switch strings.ToLower(format) {
	case "text":
		return tint.NewHandler(w, &tint.Options{
			Level:       level,
			AddSource:   true,
			ReplaceAttr: trimSourceAttr,
		})
	case "json":
		return slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level:     level,
			AddSource: true,
		})
	default:
		return tint.NewHandler(w, &tint.Options{
			Level:       level,
			AddSource:   true,
			ReplaceAttr: trimSourceAttr,
		})
	}
}

func HasLogFileOutput(outputs []string) bool {
	for _, o := range outputs {
		if o != "" && o != "-" && o != "=" {
			return true
		}
	}
	return false
}

func GetLogFileOutputPath(outputs []string) string {
	for _, o := range outputs {
		if o != "" && o != "-" && o != "=" {
			return o
		}
	}
	return ""
}

func SetOutputFormat(format string) {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	currentFormat = format
	defaultLogger = newHandlerLogger(currentLevel, format)
}

func SetLogger(logger *slog.Logger) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	defaultLogger = logger
}

func ContextUnavailable() Logger {
	return Logger{Logger: getLogger()}
}

func GetFieldsFromContext(ctx context.Context) Fields {
	fields := ctx.Value(LogFieldsContextKey)
	if fields == nil {
		return nil
	}
	return fields.(Fields)
}

func FromContext(ctx context.Context) Logger {
	logger := getLogger()
	fields := GetFieldsFromContext(ctx)
	if fields == nil {
		return Logger{Logger: logger}
	}
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	return Logger{Logger: logger.With(attrs...)}
}

func AddFields(ctx context.Context, fields Fields) context.Context {
	if fields == nil {
		return ctx
	}
	existing := ctx.Value(LogFieldsContextKey)
	var loggerFields Fields
	if existing != nil {
		loggerFields = existing.(Fields)
	} else {
		loggerFields = make(Fields)
	}
	for k, v := range fields {
		loggerFields[k] = v
	}
	return context.WithValue(ctx, LogFieldsContextKey, loggerFields)
}

func CopyFieldsFromContext(srcCtx, dstCtx context.Context) context.Context {
	if fields := GetFieldsFromContext(srcCtx); fields != nil {
		return context.WithValue(dstCtx, LogFieldsContextKey, fields)
	}
	return dstCtx
}

type Logger struct {
	*slog.Logger
}

func (l Logger) WithField(key string, value any) Logger {
	return Logger{Logger: l.Logger.With(key, value)}
}

func (l Logger) WithFields(fields Fields) Logger {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	return Logger{Logger: l.Logger.With(attrs...)}
}

func (l Logger) WithError(err error) Logger {
	return Logger{Logger: l.Logger.With("error", err)}
}

func (l Logger) WithContext(ctx context.Context) Logger {
	fields := GetFieldsFromContext(ctx)
	if fields == nil {
		return l
	}
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	return Logger{Logger: l.Logger.With(attrs...)}
}

func (l Logger) IsTracing() bool {
	return l.Logger.Enabled(context.Background(), slog.LevelDebug-1)
}

func (l Logger) IsDebugging() bool {
	return l.Logger.Enabled(context.Background(), slog.LevelDebug)
}

func (l Logger) IsInfo() bool {
	return l.Logger.Enabled(context.Background(), slog.LevelInfo)
}

func (l Logger) IsError() bool {
	return l.Logger.Enabled(context.Background(), slog.LevelError)
}

func (l Logger) IsWarn() bool {
	return l.Logger.Enabled(context.Background(), slog.LevelWarn)
}

func (l Logger) Trace(msg ...any) {
	l.Logger.Log(context.Background(), slog.LevelDebug-1, fmt.Sprint(msg...))
}

func (l Logger) Debug(args ...any) {
	l.Logger.Debug(fmt.Sprint(args...))
}

func (l Logger) Info(args ...any) {
	l.Logger.Info(fmt.Sprint(args...))
}

func (l Logger) Warn(args ...any) {
	l.Logger.Warn(fmt.Sprint(args...))
}

func (l Logger) Warning(args ...any) {
	l.Logger.Warn(fmt.Sprint(args...))
}

func (l Logger) Error(args ...any) {
	l.Logger.Error(fmt.Sprint(args...))
}

func (l Logger) Fatal(args ...any) {
	l.Logger.Log(context.Background(), slog.LevelError+1, fmt.Sprint(args...))
}

func (l Logger) Panic(args ...any) {
	msg := fmt.Sprint(args...)
	l.Logger.Log(context.Background(), slog.LevelError+2, msg)
	panic(msg)
}

func (l Logger) Log(level slog.Level, args ...any) {
	l.Logger.Log(context.Background(), level, fmt.Sprint(args...))
}

func (l Logger) Tracef(format string, args ...any) {
	l.Logger.Debug(fmt.Sprintf(format, args...))
}

func (l Logger) Debugf(format string, args ...any) {
	l.Logger.Debug(fmt.Sprintf(format, args...))
}

func (l Logger) Infof(format string, args ...any) {
	l.Logger.Info(fmt.Sprintf(format, args...))
}

func (l Logger) Warnf(format string, args ...any) {
	l.Logger.Warn(fmt.Sprintf(format, args...))
}

func (l Logger) Warningf(format string, args ...any) {
	l.Logger.Warn(fmt.Sprintf(format, args...))
}

func (l Logger) Errorf(format string, args ...any) {
	l.Logger.Error(fmt.Sprintf(format, args...))
}

func (l Logger) Fatalf(format string, args ...any) {
	l.Logger.Log(context.Background(), slog.LevelError+1, fmt.Sprintf(format, args...))
}

func (l Logger) Panicf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.Logger.Log(context.Background(), slog.LevelError+2, msg)
	panic(msg)
}

func (l Logger) Logf(level slog.Level, format string, args ...any) {
	l.Logger.Log(context.Background(), level, fmt.Sprintf(format, args...))
}
