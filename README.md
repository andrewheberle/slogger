# slogger

[![codecov](https://codecov.io/gh/andrewheberle/slogger/graph/badge.svg?token=vGWAS7HoL0)](https://codecov.io/gh/andrewheberle/slogger)

This is a helper package for `log/slog` to set up a `*slog.Logger` based on
some defaults for text, JSON or output suitable for consumption by
`sd_journal_stream_fd` (i.e., when your application is run as a systemd unit).

## Usage

```go
package main

import (
	"log/slog"
	"os"

	"github.com/andrewheberle/slogger"
)

func main() {
	// set up our logLevel variable to allow changing levels
	logLevel := new(slog.LevelVar)

	// create a logger with default settings
	logger, err := slogger.NewLogger(logLevel)
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
}
```

## Output Formats

By default the `*slog.Logger` is created with a handler as follows:

1. If it is detected that the application is run by systemd (ie `systemd` is
   the parent process) logs are output as text to `os.Stderr` in the following
   format:

   ```
   <LEVEL> MESSAGE [KEY=VALUE KEY=VALUE ...]
   ```
2. Otherwise the log is output using `slog.NewTextHandler` to `os.Stderr`

It is possible override or force a particular output format by passing
`slogger.WithLoggerType` to `slogger.NewLogger` with one of the following
options:

* `slogger.LoggerTypeAuto` - chooses systemd or text output (default)
* `slogger.LoggerTypeDiscard` - no output via `slog.DiscardHandler`
* `slogger.LoggerTypeJson` - JSON output via `slog.NewJSONHandler`
* `slogger.LoggerTypeSystemd` - custom text output for by `sd_journal_stream_fd `
* `slogger.LoggerTypeText` - Text output using `slog.NewTextHandler`
