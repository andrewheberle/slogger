package slogger_test

import (
	"log/slog"
	"os"

	"github.com/andrewheberle/slogger"
	"github.com/spf13/pflag"
)

func ExampleLogger() {
	// set up our logLevel variable to allow changing levels
	logLevel := new(slog.LevelVar)

	// create a logger to write to os.Stdout (os.Stderr is the default) in a
	// format expected by sd_journal_stream_fd
	logger, err := slogger.NewLogger(logLevel,
		slogger.WithLoggerType(slogger.LoggerTypeSystemd),
		slogger.WithWriter(os.Stdout),
	)
	if err != nil {
		panic(err)
	}

	// level is set to info by default
	logger.Debug("this will not be shown")
	logger.Info("this is an info level message", "withstring", "foo")

	// raise the log level to debug
	logLevel.Set(slog.LevelDebug)

	// output a debug message
	logger.Debug("this message is only shown if the log level is changed", "withnumber", 10)
	// Output:
	// <6> this is an info level message withstring=foo
	// <7> this message is only shown if the log level is changed withnumber=10

}

func ExampleLoggerTypeVar() {
	lt := new(slogger.LoggerTypeVar)

	// set up a flagset
	f := pflag.NewFlagSet("example", pflag.ContinueOnError)
	f.Var(lt, "type", "Logger type (auto, discard, json, systemd or text)")

	// parse as though option provided was systemd
	f.Parse([]string{"--type", "systemd"})

	// set up our logLevel variable to allow changing levels
	logLevel := new(slog.LevelVar)

	// set up the logger based on the command line flag
	logger, err := slogger.NewLogger(logLevel,
		lt.LoggerTypeOption(),
		slogger.WithWriter(os.Stdout),
	)
	if err != nil {
		panic(err)
	}

	logger.Info("output using systemd handler")
	// Output:
	// <6> output using systemd handler
}
