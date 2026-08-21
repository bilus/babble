// Code generated from BOOK.org by make tangle. DO NOT EDIT.

package dblock_test

import (
	"testing"

	"github.com/bilus/babble/internal/dblock"
	"github.com/bilus/babble/internal/org"
)

func refresh(t *testing.T, src string) (string, bool) {
	t.Helper()
	d, err := org.ParseBytes([]byte(src), "b.org")
	if err != nil {
		t.Fatal(err)
	}
	out, changed, err := dblock.Refresh(d, func(string) (string, error) {
		t.Fatal("no diff in these books, so no body should be asked for")
		return "", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(out), changed
}

func TestFinalNewlineFollowsTheSave(t *testing.T) {
	const withBlock = "#+begin: plan-toc\n#+end:\n\ntail with no newline"
	out, changed := refresh(t, withBlock)
	if !changed {
		t.Error("a book holding a dynamic block was reported unchanged")
	}
	if out[len(out)-1] != '\n' {
		t.Errorf("the refresh left the book without a final newline: %q", out)
	}

	const noBlock = "#+begin_src go :tangle no\nx\n#+end_src\n\ntail with no newline"
	out, changed = refresh(t, noBlock)
	if changed {
		t.Error("a book with no dynamic block was reported changed")
	}
	if out != noBlock {
		t.Errorf("the refresh touched a book with no dynamic block:\n got %q\nwant %q", out, noBlock)
	}
}
