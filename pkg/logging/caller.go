package logging

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
)

// This entire file is taken from logrus' implementation, the only difference is that knownLoggerFrames
// has been adjusted to skip our wrapper

var (
	// qualified package name, cached at first use
	loggingPackage string

	// Positions in the call stack when tracing to report the calling method
	// start at the bottom of the stack before the package-name cache is primed
	minimumCallerDepth = 1

	// Used for caller information initialisation
	callerInitOnce sync.Once
)

const (
	maximumCallerDepth int = 25
	knownLoggerFrames  int = 8
)

// getPackageName reduces a fully qualified function name to the package name
// There really ought to be to be a better way...
func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}

func getCaller() *runtime.Frame {
	// cache this package's fully-qualified name
	callerInitOnce.Do(func() {
		pcs := make([]uintptr, 2) //nolint: mnd
		_ = runtime.Callers(0, pcs)
		loggingPackage = getPackageName(runtime.FuncForPC(pcs[1]).Name())

		// now that we have the cache, we can skip a minimum count of known-logrus functions
		// XXX this is dubious, the number of frames may vary
		minimumCallerDepth = knownLoggerFrames
	})

	// Restrict the lookback frames to avoid runaway lookups
	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(minimumCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	for f, again := frames.Next(); again; f, again = frames.Next() {
		pkg := getPackageName(f.Function)
		// If the caller isn't part of this package, we're done
		if pkg != loggingPackage {
			return &f
		}
	}

	// if we got here, we failed to find the caller's context
	return nil
}

// PathTrimmer is used to trim the caller paths to be relative to the project root
type PathTrimmer struct {
	trimPath string
}

func NewPathTrimmer(moduleName string) *PathTrimmer {
	return &PathTrimmer{
		trimPath: moduleName[strings.LastIndex(moduleName, "/")+1:],
	}
}

// Trim trims the frame to be relative to the project root
func (t *PathTrimmer) Trim(frame *runtime.Frame, moduleName string) (function string, file string) {
	indexOfModule := strings.Index(strings.ToLower(frame.File), t.trimPath)
	if indexOfModule != -1 {
		file = frame.File[indexOfModule+len(t.trimPath):]
	} else {
		file = frame.File
	}
	file = fmt.Sprintf("%s:%d", strings.TrimPrefix(file, string(os.PathSeparator)), frame.Line)
	function = strings.TrimPrefix(frame.Function, fmt.Sprintf("%s%s", moduleName, string(os.PathSeparator)))
	return
}
