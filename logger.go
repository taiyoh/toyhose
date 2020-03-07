package toyhose

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

// SetupLogger provides setting up logger for toyhose.
func SetupLogger() error {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	lvl, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(lvl)
	zerolog.TimeFieldFormat = time.RFC3339Nano
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05.000000"}
	output.FormatLevel = func(i interface{}) string {
		return fmt.Sprintf("%-6s", i)
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf(" %s ", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s:", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}
	output.FormatCaller = func(i interface{}) string {
		t := fmt.Sprintf("%s", i)
		s := strings.Split(t, ":")
		if 2 != len(s) {
			return t
		}
		f := filepath.Base(s[0])
		return f + ":" + s[1]
	}
	log = zerolog.New(output).With().Timestamp().Caller().Logger()
	return nil
}

// Logger returns zerolog.Logger instance.
func Logger() zerolog.Logger {
	return log
}
