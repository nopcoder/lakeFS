package logging

import (
	"io"
	"log/slog"
)

func Dummy() Logger {
	return Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
}
