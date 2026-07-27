// Code generated from BOOK.org by make tangle. DO NOT EDIT.

package org_test

import (
	"strings"
	"testing"

	"github.com/bilus/babble/internal/book"
	"github.com/bilus/babble/internal/org"
)

// TestParseCorpora parses the books this repository carries and
// checks each tree against the file it came from: one src block per
// begin_src line the file opens, every body inside its own block,
// and every anchor inside it too. Fixtures pin what their author
// thought of; three real books of a few thousand lines find the
// rest.
func TestParseCorpora(t *testing.T) {
	for _, path := range []string{
		"../../BOOK.org",
		"../../lit/BOOK.org",
		"../../lit/example/BOOK.org",
	} {
		t.Run(path, func(t *testing.T) {
			d, err := org.Parse(path)
			if err != nil {
				t.Fatal(err)
			}
			want := 0
			for _, l := range strings.Split(string(d.Source), "\n") {
				if strings.HasPrefix(strings.ToLower(strings.TrimLeft(l, " \t")), "#+begin_src") {
					want++
				}
			}
			got := 0
			d.Walk(func(n book.Node) bool {
				b, ok := n.(*book.SrcBlock)
				if !ok {
					return true
				}
				got++
				if b.Body.Start < b.Full.Start || b.Body.End > b.Full.End {
					t.Errorf("line %d: body %v outside the block %v", b.Line, b.Body, b.Full)
				}
				if b.BeginAt < b.Full.Start || b.AfterEnd > b.Full.End {
					t.Errorf("line %d: anchors outside the block %v", b.Line, b.Full)
				}
				return true
			})
			if got != want {
				t.Errorf("parsed %d src blocks, file opens %d", got, want)
			}
		})
	}
}
