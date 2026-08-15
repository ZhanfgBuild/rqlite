package db

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	command "github.com/rqlite/rqlite/v10/command/proto"
)

func TestSlowQueryLogging(t *testing.T) {
	db, path := mustCreateOnDiskDatabase()
	defer db.Close()
	defer os.Remove(path)

	if got, want := db.slowQueryThreshold, defaultSlowQueryThreshold; got != want {
		t.Fatalf("unexpected default slow-query threshold: got %s, want %s", got, want)
	}

	var buf bytes.Buffer
	db.logger.SetOutput(&buf)
	db.logger.SetFlags(0)
	db.SetSlowQueryThreshold(time.Nanosecond)

	const createStmt = "CREATE TABLE foo (id INTEGER NOT NULL PRIMARY KEY, name TEXT)"
	mustExecute(db, createStmt)
	if got := buf.String(); !strings.Contains(got, "slow query:") || !strings.Contains(got, createStmt) {
		t.Fatalf("execute did not log the full slow SQL statement: %q", got)
	}

	buf.Reset()
	const queryStmt = "SELECT * FROM foo"
	mustQuery(db, queryStmt)
	if got := buf.String(); !strings.Contains(got, "slow query:") || !strings.Contains(got, queryStmt) {
		t.Fatalf("query did not log the full slow SQL statement: %q", got)
	}

	buf.Reset()
	_, err := db.Execute(&command.Request{
		Statements: []*command.Statement{
			{
				Sql:        "SELECT 1",
				ForceQuery: true,
			},
		},
	}, false)
	if err != nil {
		t.Fatalf("force query failed: %s", err.Error())
	}
	if got := strings.Count(buf.String(), "slow query:"); got != 1 {
		t.Fatalf("force query logged %d times, want 1: %q", got, buf.String())
	}

	buf.Reset()
	db.SetSlowQueryThreshold(0)
	mustQuery(db, "SELECT 1")
	if got := buf.String(); got != "" {
		t.Fatalf("slow-query logging was not disabled: %q", got)
	}
}
