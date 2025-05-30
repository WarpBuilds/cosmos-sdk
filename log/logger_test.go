package log_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"cosmossdk.io/log"
)

func inner() error {
	return errors.New("seems we have an error here")
}

type _MockHook string

func (h _MockHook) Run(e *zerolog.Event, l zerolog.Level, msg string) {
	e.Bool(string(h), true)
}

func TestLoggerOptionHooks(t *testing.T) {
	buf := new(bytes.Buffer)
	var (
		mockHook1 _MockHook = "mock_message1"
		mockHook2 _MockHook = "mock_message2"
	)
	logger := log.NewLogger(buf, log.HooksOption(mockHook1, mockHook2), log.ColorOption(false))
	logger.Info("hello world")
	if !strings.Contains(buf.String(), "mock_message1=true") {
		t.Fatalf("expected mock_message1=true, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "mock_message2=true") {
		t.Fatalf("expected mock_message2=true, got: %s", buf.String())
	}

	buf.Reset()
	logger = log.NewLogger(buf, log.HooksOption(), log.ColorOption(false))
	logger.Info("hello world")
	if !strings.Contains(buf.String(), "hello world") {
		t.Fatalf("expected hello world, got: %s", buf.String())
	}
}

func TestLoggerOptionStackTrace(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := log.NewLogger(buf, log.TraceOption(true), log.ColorOption(false))
	logger.Error("this log should be displayed", "error", inner())
	if strings.Count(buf.String(), "logger_test.go") != 1 {
		t.Fatalf("stack trace not found, got: %s", buf.String())
	}
	buf.Reset()

	logger = log.NewLogger(buf, log.TraceOption(false), log.ColorOption(false))
	logger.Error("this log should be displayed", "error", inner())
	if strings.Count(buf.String(), "logger_test.go") > 0 {
		t.Fatalf("stack trace found, got: %s", buf.String())
	}
}
