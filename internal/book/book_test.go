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
