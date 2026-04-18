package logging

import (
	"context"
	"io"
	"testing"
)

func TestContextUnavailable(t *testing.T) {
	logger := ContextUnavailable()
	if logger.Logger == nil {
		t.Error("ContextUnavailable() returned nil logger")
	}
}

func TestFromContext(t *testing.T) {
	t.Run("context without fields", func(t *testing.T) {
		ctx := context.Background()
		logger := FromContext(ctx)
		if logger.Logger == nil {
			t.Error("FromContext() returned nil")
		}
	})

	t.Run("context with fields", func(t *testing.T) {
		ctx := AddFields(context.Background(), Fields{
			"request_id": "req-123",
			"user":       "alice",
		})
		logger := FromContext(ctx)
		if logger.Logger == nil {
			t.Error("FromContext() returned nil")
		}
	})
}

func TestGetFieldsFromContext(t *testing.T) {
	t.Run("context with no fields", func(t *testing.T) {
		ctx := context.Background()
		fields := GetFieldsFromContext(ctx)
		if fields != nil {
			t.Errorf("Expected nil fields, got %v", fields)
		}
	})

	t.Run("context with fields", func(t *testing.T) {
		ctx := AddFields(context.Background(), Fields{
			RequestIDFieldKey: "req-789",
			UserFieldKey:      "dave",
		})
		fields := GetFieldsFromContext(ctx)
		if fields == nil {
			t.Fatal("Expected fields, got nil")
		}
		if fields[RequestIDFieldKey] != "req-789" {
			t.Errorf("RequestIDFieldKey: got %v, want %v", fields[RequestIDFieldKey], "req-789")
		}
		if fields[UserFieldKey] != "dave" {
			t.Errorf("UserFieldKey: got %v, want %v", fields[UserFieldKey], "dave")
		}
	})
}

func TestAddFields(t *testing.T) {
	t.Run("add fields to empty context", func(t *testing.T) {
		ctx := AddFields(context.Background(), Fields{"key": "value"})
		fields := GetFieldsFromContext(ctx)
		if fields == nil {
			t.Fatal("Expected fields, got nil")
		}
		if fields["key"] != "value" {
			t.Errorf("key: got %v, want %v", fields["key"], "value")
		}
	})

	t.Run("add nil fields returns same context", func(t *testing.T) {
		ctx := context.Background()
		ctx2 := AddFields(ctx, nil)
		if ctx != ctx2 {
			t.Error("AddFields with nil should return same context")
		}
	})
}

func TestCopyFieldsFromContext(t *testing.T) {
	srcCtx := AddFields(context.Background(), Fields{"key": "value"})
	dstCtx := CopyFieldsFromContext(srcCtx, context.Background())
	fields := GetFieldsFromContext(dstCtx)
	if fields == nil {
		t.Fatal("Expected fields, got nil")
	}
	if fields["key"] != "value" {
		t.Errorf("key: got %v, want %v", fields["key"], "value")
	}
}

func TestLoggerWrapper(t *testing.T) {
	t.Run("WithField", func(t *testing.T) {
		logger := ContextUnavailable()
		result := logger.WithField("key", "value")
		if result.Logger == nil {
			t.Error("WithField returned nil")
		}
	})

	t.Run("WithFields", func(t *testing.T) {
		logger := ContextUnavailable()
		result := logger.WithFields(Fields{"key1": "val1", "key2": "val2"})
		if result.Logger == nil {
			t.Error("WithFields returned nil")
		}
	})

	t.Run("WithError", func(t *testing.T) {
		logger := ContextUnavailable()
		result := logger.WithError(io.ErrUnexpectedEOF)
		if result.Logger == nil {
			t.Error("WithError returned nil")
		}
	})

	t.Run("Info", func(t *testing.T) {
		logger := ContextUnavailable()
		logger.Info("test message")
	})

	t.Run("Debug", func(t *testing.T) {
		logger := ContextUnavailable()
		logger.Debug("debug message")
	})

	t.Run("Warn", func(t *testing.T) {
		logger := ContextUnavailable()
		logger.Warn("warn message")
	})

	t.Run("Error", func(t *testing.T) {
		logger := ContextUnavailable()
		logger.Error("error message")
	})

	t.Run("Infof", func(t *testing.T) {
		logger := ContextUnavailable()
		logger.Infof("formatted %s", "message")
	})

	t.Run("WithContext", func(t *testing.T) {
		ctx := AddFields(context.Background(), Fields{"key": "value"})
		logger := ContextUnavailable()
		result := logger.WithContext(ctx)
		if result.Logger == nil {
			t.Error("WithContext returned nil")
		}
	})

	t.Run("IsDebugging", func(t *testing.T) {
		logger := ContextUnavailable()
		_ = logger.IsDebugging()
	})

	t.Run("IsInfo", func(t *testing.T) {
		logger := ContextUnavailable()
		_ = logger.IsInfo()
	})
}

func TestSetLevel(t *testing.T) {
	t.Run("set debug level", func(t *testing.T) {
		SetLevel("debug")
		if Level() != "debug" {
			t.Errorf("Expected level debug, got %s", Level())
		}
	})

	t.Run("set info level", func(t *testing.T) {
		SetLevel("info")
		if Level() != "info" {
			t.Errorf("Expected level info, got %s", Level())
		}
	})

	t.Run("set warn level", func(t *testing.T) {
		SetLevel("warn")
		if Level() != "warn" {
			t.Errorf("Expected level warn, got %s", Level())
		}
	})

	t.Run("set error level", func(t *testing.T) {
		SetLevel("error")
		if Level() != "error" {
			t.Errorf("Expected level error, got %s", Level())
		}
	})
}

func TestDummy(t *testing.T) {
	logger := Dummy()
	if logger.Logger == nil {
		t.Error("Dummy() returned nil")
	}
}

func TestHasLogFileOutput(t *testing.T) {
	tests := []struct {
		name     string
		outputs  []string
		expected bool
	}{
		{"empty", []string{}, false},
		{"stdout", []string{"-"}, false},
		{"stderr", []string{"="}, false},
		{"file", []string{"/tmp/test.log"}, true},
		{"mixed", []string{"-", "/tmp/test.log"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasLogFileOutput(tt.outputs)
			if result != tt.expected {
				t.Errorf("HasLogFileOutput(%v) = %v, want %v", tt.outputs, result, tt.expected)
			}
		})
	}
}

func TestGetLogFileOutputPath(t *testing.T) {
	tests := []struct {
		name     string
		outputs  []string
		expected string
	}{
		{"empty", []string{}, ""},
		{"stdout", []string{"-"}, ""},
		{"file", []string{"/tmp/test.log"}, "/tmp/test.log"},
		{"mixed", []string{"-", "/tmp/test.log"}, "/tmp/test.log"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLogFileOutputPath(tt.outputs)
			if result != tt.expected {
				t.Errorf("GetLogFileOutputPath(%v) = %v, want %v", tt.outputs, result, tt.expected)
			}
		})
	}
}

func TestFieldsConstant(t *testing.T) {
	if RepositoryFieldKey != "repository" {
		t.Errorf("RepositoryFieldKey = %s, want repository", RepositoryFieldKey)
	}
	if RequestIDFieldKey != "request_id" {
		t.Errorf("RequestIDFieldKey = %s, want request_id", RequestIDFieldKey)
	}
	if UserFieldKey != "user" {
		t.Errorf("UserFieldKey = %s, want user", UserFieldKey)
	}
}
