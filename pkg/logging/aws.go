package logging

import (
	"context"
	"fmt"

	"github.com/aws/smithy-go/logging"
)

type AWSAdapter struct {
	Logger Logger
}

func (l *AWSAdapter) Logf(classification logging.Classification, format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	if classification == logging.Warn {
		l.Logger.Warn(msg)
	} else {
		l.Logger.Debug(msg)
	}
}

func (l *AWSAdapter) WithContext(ctx context.Context) Logger {
	return l.Logger
}
