// Code generated from BOOK.org by make tangle. DO NOT EDIT.

package org_test

import (
	"strings"
	"testing"

	"github.com/bilus/babble/internal/book"
	"github.com/bilus/babble/internal/org"
)

// The count it compares against comes from a scanner that shares no
// code with the parser, which is what makes it a check rather than a
// tautology. It skips the interior of a verbatim block, where a
// delimiter is text and not a delimiter, and the interior of a src
// block, where the same is true; a greater block's interior is
// walked, since blocks nest there for real.
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
			want := countSrcBlocks(string(d.Source))
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

func countSrcBlocks(src string) int {
	n, inside := 0, ""
	for _, l := range strings.Split(src, "\n") {
		t := strings.ToLower(strings.TrimLeft(l, " \t"))
		if inside != "" {
			if strings.HasPrefix(t, "#+end_"+inside) {
				inside = ""
			}
			continue
		}
		if !strings.HasPrefix(t, "#+begin_") {
			continue
		}
		kind := strings.TrimPrefix(t, "#+begin_")
		if i := strings.IndexAny(kind, " \t"); i >= 0 {
			kind = kind[:i]
		}
		switch kind {
		case "src":
			n++
			inside = kind
		case "example", "export", "comment", "verse":
			inside = kind
		}
	}
	return n
}
