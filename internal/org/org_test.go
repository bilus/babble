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

// TestResolveHeaders walks the merge order: the driver's defaults,
// then what a heading's drawer passes down, then the block's own
// line, with a nearer drawer replacing a farther one whole rather
// than adding to it.
func TestResolveHeaders(t *testing.T) {
	const src = `
#+property: header-args :comments org

#+begin_src go
the file property reaches here
#+end_src

* Outer
:PROPERTIES:
:header-args: :tangle outer.go :noweb yes
:END:

#+begin_src go
the drawer replaces the file property
#+end_src

** Inner
:PROPERTIES:
:header-args: :tangle inner.go
:END:

#+begin_src go :padline no
its own line wins
#+end_src
`
	blocks := srcBlocks(t, src)
	if len(blocks) != 3 {
		t.Fatalf("parsed %d blocks, want 3", len(blocks))
	}
	if got := blocks[0].Params; got.Comments != "org" || got.Tangle != "no" {
		t.Errorf("file property: %+v", got)
	}
	// A nearer drawer replaces a farther setting whole, so the file's
	// :comments does not survive into Outer.
	if got := blocks[1].Params; got.Tangle != "outer.go" || !got.Noweb || got.Comments != "no" {
		t.Errorf("outer block: %+v", got)
	}
	// Inner omits :noweb, and omitting is not inheriting.
	if got := blocks[2].Params; got.Tangle != "inner.go" || got.Noweb || got.Padline {
		t.Errorf("inner block: %+v", got)
	}
}

// TestHeaderRejects pins the fence: a key outside the table, a key
// inside it that babble refuses, a value outside the pair a key
// allows, and a value shaped like lisp. Each names itself, because
// an error a reader cannot act on is barely better than silence.
func TestHeaderRejects(t *testing.T) {
	for _, tc := range []struct{ header, want string }{
		{":tangel x.go", "unknown header argument :tangel"},
		{":var x=1", ":var is not in the subset"},
		{":comments link", ":comments link is not in the subset"},
		{":noweb strip-tangle", ":noweb strip-tangle is not in the subset"},
		{":tangle (concat \"a\" \".go\")", "lisp-valued header argument"},
	} {
		src := "#+begin_src go " + tc.header + "\nx\n#+end_src\n"
		d, err := org.ParseBytes([]byte(src), "b.org")
		if err != nil {
			t.Fatal(err)
		}
		err = org.ResolveAll(d)
		if err == nil {
			t.Errorf("%s: resolved without error", tc.header)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %q, want it to mention %q", tc.header, err, tc.want)
		}
	}
}

// TestInertHeaders pins the other half of the table: the keys org
// parses and its tangler never reads pass through without changing
// anything, which is what lets a book carry them.
func TestInertHeaders(t *testing.T) {
	const header = ":results output :exports code :session none :cache no :tangle x.go"
	blocks := srcBlocks(t, "#+begin_src sh "+header+"\necho hi\n#+end_src\n")
	if got := blocks[0].Params; got.Tangle != "x.go" {
		t.Errorf("inert keys disturbed the resolution: %+v", got)
	}
}

func srcBlocks(t *testing.T, src string) []*book.SrcBlock {
	t.Helper()
	d, err := org.ParseBytes([]byte(src), "b.org")
	if err != nil {
		t.Fatal(err)
	}
	if err := org.ResolveAll(d); err != nil {
		t.Fatal(err)
	}
	var out []*book.SrcBlock
	d.Walk(func(n book.Node) bool {
		if b, ok := n.(*book.SrcBlock); ok {
			out = append(out, b)
		}
		return true
	})
	return out
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
