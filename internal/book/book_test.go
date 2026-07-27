// Code generated from BOOK.org by make tangle. DO NOT EDIT.

package book_test

import (
	"testing"

	"github.com/bilus/babble/internal/book"
)

// TestLineAt pins the offset-to-line arithmetic the writers lean on,
// boundary offsets included: an offset on a newline belongs to the
// line that newline ends, and the offset one past the final byte
// lands on a virtual next line.
func TestLineAt(t *testing.T) {
	d := &book.Document{Source: []byte("one\ntwo\nthree\n")}
	for _, tc := range []struct{ off, line int }{
		{0, 1}, {3, 1}, {4, 2}, {7, 2}, {8, 3}, {13, 3}, {14, 4},
	} {
		if got := d.LineAt(tc.off); got != tc.line {
			t.Errorf("LineAt(%d) = %d, want %d", tc.off, got, tc.line)
		}
	}
}

// TestWalk pins the pruning contract: returning false skips a
// subtree's children without touching its siblings, which is exactly
// what COMMENT skipping needs.
func TestWalk(t *testing.T) {
	d := &book.Document{Nodes: []book.Node{
		&book.Headline{Line: 1, Children: []book.Node{&book.Prose{Line: 2}}},
		&book.Headline{Line: 3, Children: []book.Node{&book.Prose{Line: 4}}},
		&book.Prose{Line: 5},
	}}
	var lines []int
	d.Walk(func(n book.Node) bool {
		switch n := n.(type) {
		case *book.Headline:
			lines = append(lines, n.Line)
			return n.Line != 3
		case *book.Prose:
			lines = append(lines, n.Line)
		}
		return true
	})
	want := []int{1, 2, 3, 5}
	if len(lines) != len(want) {
		t.Fatalf("visited %v, want %v", lines, want)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("visited %v, want %v", lines, want)
		}
	}
}
