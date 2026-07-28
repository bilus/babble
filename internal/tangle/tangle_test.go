// Code generated from BOOK.org by make tangle. DO NOT EDIT.

package tangle_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bilus/babble/internal/org"
	"github.com/bilus/babble/internal/tangle"
)

// TestSkipsIdenticalWrites pins the one place babble parts company
// with the oracle on purpose. A target whose bytes already match is
// not rewritten, so its timestamp survives and nothing downstream
// rebuilds; a target whose bytes changed is written. The test sets
// the timestamp back rather than sleeping, so it measures the write
// and not the clock.
func TestSkipsIdenticalWrites(t *testing.T) {
	dir := t.TempDir()
	book := filepath.Join(dir, "b.org")
	target := filepath.Join(dir, "x.go")
	write := func(body string) {
		src := "#+begin_src go :tangle x.go\n" + body + "\n#+end_src\n"
		if err := os.WriteFile(book, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		d, err := org.Parse(book)
		if err != nil {
			t.Fatal(err)
		}
		if err := org.ResolveAll(d); err != nil {
			t.Fatal(err)
		}
		if err := tangle.Run(d); err != nil {
			t.Fatal(err)
		}
	}
	write("package x")
	old := time.Now().Add(-time.Hour).Truncate(time.Second)
	if err := os.Chtimes(target, old, old); err != nil {
		t.Fatal(err)
	}
	write("package x")
	st, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if !st.ModTime().Equal(old) {
		t.Errorf("rewrote a target whose bytes did not change")
	}
	write("package y")
	if st, err = os.Stat(target); err != nil {
		t.Fatal(err)
	} else if st.ModTime().Equal(old) {
		t.Errorf("did not write a target whose bytes changed")
	}
}
