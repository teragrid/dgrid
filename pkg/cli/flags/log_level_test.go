package flags_test

import (
	"bytes"
	"strings"
	"testing"

	tmflags "github.com/tgrid/tgrid/pkg/cli/flags"
	"github.com/teragrid/dgrid/pkg/log"
)

const (
	defaultLogLevelValue = "info"
)

func TestParseLogLevel(t *testing.T) {
	var buf bytes.Buffer
	jsonLogger := log.NewTMJSONLogger(&buf)

	correctLogLevels := []struct {
		lvl              string
		expectedLogLines []string
	}{
		{"storage:error", []string{
			``, // if no default is given, assume info
			``,
			`{"_msg":"Mesmero","level":"error","module":"storage"}`,
			`{"_msg":"Mind","level":"info","module":"state"}`, // if no default is given, assume info
			``}},

		{"storage:error,*:debug", []string{
			`{"_msg":"Kingpin","level":"debug","module":"wire"}`,
			``,
			`{"_msg":"Mesmero","level":"error","module":"storage"}`,
			`{"_msg":"Mind","level":"info","module":"state"}`,
			`{"_msg":"Gideon","level":"debug"}`}},

		{"*:debug,wire:none", []string{
			``,
			`{"_msg":"Kitty Pryde","level":"info","module":"storage"}`,
			`{"_msg":"Mesmero","level":"error","module":"storage"}`,
			`{"_msg":"Mind","level":"info","module":"state"}`,
			`{"_msg":"Gideon","level":"debug"}`}},
	}

	for _, c := range correctLogLevels {
		logger, err := tmflags.ParseLogLevel(c.lvl, jsonLogger, defaultLogLevelValue)
		if err != nil {
			t.Fatal(err)
		}

		buf.Reset()

		logger.With("module", "storage").With("module", "wire").Debug("Kingpin")
		if have := strings.TrimSpace(buf.String()); c.expectedLogLines[0] != have {
			t.Errorf("\nwant '%s'\nhave '%s'\nlevel '%s'", c.expectedLogLines[0], have, c.lvl)
		}

		buf.Reset()

		logger.With("module", "storage").Info("Kitty Pryde")
		if have := strings.TrimSpace(buf.String()); c.expectedLogLines[1] != have {
			t.Errorf("\nwant '%s'\nhave '%s'\nlevel '%s'", c.expectedLogLines[1], have, c.lvl)
		}

		buf.Reset()

		logger.With("module", "storage").Error("Mesmero")
		if have := strings.TrimSpace(buf.String()); c.expectedLogLines[2] != have {
			t.Errorf("\nwant '%s'\nhave '%s'\nlevel '%s'", c.expectedLogLines[2], have, c.lvl)
		}

		buf.Reset()

		logger.With("module", "state").Info("Mind")
		if have := strings.TrimSpace(buf.String()); c.expectedLogLines[3] != have {
			t.Errorf("\nwant '%s'\nhave '%s'\nlevel '%s'", c.expectedLogLines[3], have, c.lvl)
		}

		buf.Reset()

		logger.Debug("Gideon")
		if have := strings.TrimSpace(buf.String()); c.expectedLogLines[4] != have {
			t.Errorf("\nwant '%s'\nhave '%s'\nlevel '%s'", c.expectedLogLines[4], have, c.lvl)
		}
	}

	incorrectLogLevel := []string{"some", "storage:some", "*:some,storage:error"}
	for _, lvl := range incorrectLogLevel {
		if _, err := tmflags.ParseLogLevel(lvl, jsonLogger, defaultLogLevelValue); err == nil {
			t.Fatalf("Expected %s to produce error", lvl)
		}
	}
}
