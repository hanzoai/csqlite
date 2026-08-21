package csqlite

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
)

// An entity group is many small databases queried together, so the ATTACH
// ceiling is the ceiling on that whole shape. Upstream refuses above 125 and
// breaks at 254; both are byte-width artefacts, not structure.
//
//	#error SQLITE_MAX_ATTACHED>125     a constant, now removed
//	u8  sqlite3InitInfo.iDb            wrapped at 256 -> int
//	i8  VdbeCursor.iDb                 signed, holds 127 -> int
//
// yDbMask was never the limit: above 30 attachments it is already a bit array
// sized from the constant, one bit per database.
//
// Build with -DSQLITE_MAX_ATTACHED=N to raise it. The real cost is memory:
// ~111 KB per attached database, measured, so ten thousand is about 1.1 GB on
// one connection. That number and not the constant decides a group's size.
func TestAttachBeyondTheOldCeiling(t *testing.T) {
	dir := t.TempDir()
	db, err := sql.Open("sqlite3", filepath.Join(dir, "main.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	// Attach until the build refuses, so the test reads the ceiling rather than
	// assuming one. A default build stops at 10 and has nothing to prove here.
	const want = 300 // past 254, where the u8 index used to corrupt the schema
	n := 0
	for ; n < want; n++ {
		p := filepath.Join(dir, fmt.Sprintf("e%d.db", n))
		if _, err := db.Exec(fmt.Sprintf("ATTACH DATABASE '%s' AS e%d", p, n)); err != nil {
			break
		}
		if _, err := db.Exec(fmt.Sprintf("CREATE TABLE e%d.row(id INTEGER PRIMARY KEY, v TEXT)", n)); err != nil {
			t.Fatalf("schema on e%d: %v", n, err)
		}
	}
	if n <= 125 {
		t.Skipf("this build caps ATTACH at %d; rebuild with -DSQLITE_MAX_ATTACHED=%d to exercise the widened indices", n, want)
	}

	// Round-trip at the boundaries that used to give out. A wrong answer would
	// be worse than an error, so the value is read back, not just counted.
	for _, i := range []int{0, 127, 128, 255, n - 1} {
		if i >= n {
			continue
		}
		want := fmt.Sprintf("v%d", i)
		if _, err := db.Exec(fmt.Sprintf("INSERT INTO e%d.row(v) VALUES('%s')", i, want)); err != nil {
			t.Fatalf("write e%d: %v", i, err)
		}
		var got string
		if err := db.QueryRow(fmt.Sprintf("SELECT v FROM e%d.row LIMIT 1", i)).Scan(&got); err != nil {
			t.Fatalf("read e%d: %v", i, err)
		}
		if got != want {
			t.Fatalf("e%d returned %q, want %q", i, got, want)
		}
	}

	// One question of many files — the shape the group exists for.
	var rows int
	q := fmt.Sprintf("SELECT count(*) FROM (SELECT v FROM e0.row UNION ALL SELECT v FROM e128.row UNION ALL SELECT v FROM e%d.row)", n-1)
	if err := db.QueryRow(q).Scan(&rows); err != nil {
		t.Fatalf("cross-database query: %v", err)
	}
	if rows != 3 {
		t.Fatalf("union returned %d rows, want 3", rows)
	}
}
