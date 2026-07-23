package slogger

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// check that fmt.Stringer is implemented
var _ fmt.Stringer = &LoggerTypeVar{}

// The type that defines which log handler to create via [NewLogger]
type LoggerType string

// String constants for each log handler type.
const (
	LoggerTypeAuto    LoggerType = "auto"
	LoggerTypeDiscard LoggerType = "discard"
	LoggerTypeJson    LoggerType = "json"
	LoggerTypeSystemd LoggerType = "systemd"
	LoggerTypeText    LoggerType = "text"
)

// This type satisfies the [github.com/spf13/pflag.Value] interface so
// [LoggerType] can be set via a command line flag.
type LoggerTypeVar struct {
	val LoggerType
}

// String method for [LoggerTypeVar] to satisfy the
// [github.com/spf13/pflag.Value] and [fmt.Stringer] interfaces.
func (lt *LoggerTypeVar) String() string {
	switch lt.val {
	case LoggerTypeAuto, "":
		return "auto"
	case LoggerTypeDiscard:
		return "discard"
	case LoggerTypeJson:
		return "json"
	case LoggerTypeSystemd:
		return "systemd"
	case LoggerTypeText:
		return "text"
	}

	return "unknown"
}

// The Set method is to satisfy the [github.com/spf13/pflag.Value] interface.
func (lt *LoggerTypeVar) Set(s string) error {
	switch s {
	case "auto", "":
		lt.val = LoggerTypeAuto
	case "discard":
		lt.val = LoggerTypeDiscard
	case "json":
		lt.val = LoggerTypeJson
	case "systemd":
		lt.val = LoggerTypeSystemd
	case "text":
		lt.val = LoggerTypeText
	default:
		return ErrUnhandledLoggerType
	}

	return nil
}

// The Type method is to satisfy the [github.com/spf13/pflag.Value] interface.
func (lt LoggerTypeVar) Type() string {
	return "string"
}

// Returns a [LoggerOption] suitable for passing to [NewLogger].
func (lt LoggerTypeVar) LoggerTypeOption() LoggerOption {
	return WithLoggerType(lt.val)
}

var (
	// This error is returned when an invalid [LoggerType] is provided to
	// [NewLogger] via [WithLoggerType].
	ErrUnhandledLoggerType = errors.New("unhandled logger type")
)

type logger struct {
	lt LoggerType
	w  io.Writer
}

// NewLogger creates a new [slog.Logger] with the provided [slog.Leveler].
// By default the [slog.Logger] outputs to [os.Stderr] and will output using
// either [slog.NewTextHandler] or a [SystemdHandler].
func NewLogger(leveler slog.Leveler, opts ...LoggerOption) (*slog.Logger, error) {
	l := &logger{
		lt: LoggerTypeAuto,
		w:  os.Stderr,
	}

	for _, o := range opts {
		o(l)
	}

	if l.lt == LoggerTypeAuto || l.lt == "" {
		if isSystemd() {
			l.lt = LoggerTypeSystemd
		} else {
			l.lt = LoggerTypeText
		}
	}

	switch l.lt {
	case LoggerTypeDiscard:
		// discard handler
		return slog.New(slog.DiscardHandler), nil
	case LoggerTypeJson:
		// standard JSON handler
		return slog.New(slog.NewJSONHandler(l.w, &slog.HandlerOptions{Level: leveler})), nil
	case LoggerTypeSystemd:
		// output as <LEVEL> message key=value... under systemd
		h := NewSystemdHandler(l.w, &slog.HandlerOptions{Level: leveler})
		return slog.New(h), nil
	case LoggerTypeText:
		// standard text handler
		return slog.New(slog.NewTextHandler(l.w, &slog.HandlerOptions{Level: leveler})), nil
	}

	return nil, fmt.Errorf("%w: %s", ErrUnhandledLoggerType, l.lt)
}

// LoggerOption allows passing options to [NewLogger]
type LoggerOption func(*logger)

// WithLoggerType allows overriding the default [LoggerType] of
// [LoggerTypeAuto]
func WithLoggerType(lt LoggerType) LoggerOption {
	return func(l *logger) {
		l.lt = lt
	}
}

// WithWriter accepts an [io.Writer] in order to override the default of
// sending logs to [os.Stderr]
func WithWriter(w io.Writer) LoggerOption {
	return func(l *logger) {
		l.w = w
	}
}

func isSystemd() bool {
	ppid := os.Getppid()

	return commIs(ppid, "systemd")
}

func commIs(pid int, name string) bool {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/comm")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == name
}
