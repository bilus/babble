// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Package org is the frontend: everything that knows the input
// syntax lives here, and it hands the engine a finished tree. Parse
// reads a file; ParseBytes does the work, so tests and the refresh
// reparse can feed bytes directly.
package org

import "github.com/bilus/babble/internal/book"

func Parse(path string) (*book.Document, error) {
	panic("HOLE(2): read path and hand ParseBytes the bytes")
}

func ParseBytes(src []byte, path string) (*book.Document, error) {
	panic("HOLE(2): headlines, keywords, and prose into a spanned tree")
}

// parseBlocks is stage 3's half of the scanner: src, dynamic, quote,
// and verbatim blocks, with names, spans, and the two extent
// anchors.
func parseBlocks(src []byte, d *book.Document) error {
	panic("HOLE(3): blocks with names, spans, and the two extent anchors")
}

// ResolveAll fills every src block's Params through the closed
// table: the merge order, then the three bins, an unknown key an
// error. The splitter and reader below are its tools, ported from
// org's balanced splitter and value reader.
func ResolveAll(d *book.Document) error {
	panic("HOLE(4): every block's Params through the closed table, unknown keys error by name")
}

func balancedSplit(s string) []string {
	panic("HOLE(4): split on space-then-colon outside parens, brackets, and double quotes")
}

func readValue(s string) (string, error) {
	panic("HOLE(4): chomp, unquote an elisp string literal, reject lisp")
}

// Lint is the fence: the refusals that need the whole document
// rather than one block.
func Lint(d *book.Document) error {
	panic("HOLE(10): quoted delimiters in verbatim bodies, duplicate names, a name doubling as a noweb-ref")
}
