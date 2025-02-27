package flog

import (
	"log/slog"
	"strings"

	flog "github.com/flokiorg/go-flokicoin/log"
)

// Redefine the levels here so that any package importing the original flog
// Level does not need to import both the old and new modules.
const (
	LevelTrace    = flog.LevelTrace
	LevelDebug    = flog.LevelDebug
	LevelInfo     = flog.LevelInfo
	LevelWarn     = flog.LevelWarn
	LevelError    = flog.LevelError
	LevelCritical = flog.LevelCritical
	LevelOff      = flog.LevelOff
)

// LevelFromString returns a level based on the input string s.  If the input
// can't be interpreted as a valid log level, the info level and false is
// returned.
func LevelFromString(s string) (l flog.Level, ok bool) {
	switch strings.ToLower(s) {
	case "trace", "trc":
		return LevelTrace, true
	case "debug", "dbg":
		return LevelDebug, true
	case "info", "inf":
		return LevelInfo, true
	case "warn", "wrn":
		return LevelWarn, true
	case "error", "err":
		return LevelError, true
	case "critical", "crt":
		return LevelCritical, true
	case "off":
		return LevelOff, true
	default:
		return LevelInfo, false
	}
}

// slog uses some pre-defined level integers. So we will need to sometimes map
// between the flog.Level and the slog level. The slog library defines a few
// of the commonly used levels and allows us to add a few of our own too.
const (
	levelTrace    slog.Level = -5
	levelDebug               = slog.LevelDebug
	levelInfo                = slog.LevelInfo
	levelWarn                = slog.LevelWarn
	levelError               = slog.LevelError
	levelCritical slog.Level = 9
	levelOff      slog.Level = 10
)

// toSlogLevel converts a flog.Level to the associated slog.Level type.
func toSlogLevel(l flog.Level) slog.Level {
	switch l {
	case LevelTrace:
		return levelTrace
	case LevelDebug:
		return levelDebug
	case LevelInfo:
		return levelInfo
	case LevelWarn:
		return levelWarn
	case LevelError:
		return levelError
	case LevelCritical:
		return levelCritical
	default:
		return levelOff
	}
}

// fromSlogLevel converts an slog.Level type to the associated flog.Level
// type.
func fromSlogLevel(l slog.Level) flog.Level {
	switch l {
	case levelTrace:
		return LevelTrace
	case levelDebug:
		return LevelDebug
	case levelInfo:
		return LevelInfo
	case levelWarn:
		return LevelWarn
	case levelError:
		return LevelError
	case levelCritical:
		return LevelCritical
	default:
		return LevelOff
	}
}
