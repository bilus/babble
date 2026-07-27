// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Package dblock owns the book side: Refresh splices a fresh
// interior into every dynamic block and reports whether the bytes
// changed, so the caller writes the book only when they did.
package dblock

import "github.com/bilus/babble/internal/book"

func Refresh(d *book.Document) ([]byte, bool, error) {
	panic("HOLE(9): splice fresh interiors into the retained bytes, report change")
}

// The two writers, one per dynamic-block name. planToc runs the
// ported census patterns over the retained source lines; the tree
// only supplies the splice points and the line index.
func planToc(d *book.Document, at *book.DynamicBlock) (string, error) {
	panic("HOLE(9): a line link per plan site, sites below the block shifted by the insert count")
}

func srcBlockDiff(d *book.Document, at *book.DynamicBlock) (string, error) {
	panic("HOLE(9): system diff -u over the two named bodies, temp-file headers dropped")
}
