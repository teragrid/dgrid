package log_test

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/go-logfmt/logfmt"
	"github.com/teragrid/dgrid/pkg/log"
)

func TestLoggerLogsItsErrors(t *testing.T) {
	var buf bytes.Buffer

	logger := log.NewTGLogger(&buf)
	logger.Info("foo", "baz baz", "bar")
	msg := strings.TrimSpace(buf.String())
	if !strings.Contains(msg, logfmt.ErrInvalidKey.Error()) {
		t.Errorf("Expected logger msg to contain ErrInvalidKey, got %s", msg)
	}
}

func BenchmarkTGLoggerSimple(b *testing.B) {
	benchmarkRunner(b, log.NewTGLogger(ioutil.Discard), baseInfoMessage)
}

func BenchmarkTGLoggerContextual(b *testing.B) {
	benchmarkRunner(b, log.NewTGLogger(ioutil.Discard), withInfoMessage)
}

func benchmarkRunner(b *testing.B, logger log.Logger, f func(log.Logger)) {
	lc := logger.With("common_key", "common_value")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f(lc)
	}
}

var (
	baseInfoMessage = func(logger log.Logger) { logger.Info("foo_message", "foo_key", "foo_value") }
	withInfoMessage = func(logger log.Logger) { logger.With("a", "b").Info("c", "d", "f") }
)
