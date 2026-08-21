// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Package cmd holds what a command does once its arguments are read,
// so that the command line and the oracle harness run the same steps
// in the same order.
package cmd

import "github.com/bilus/babble/internal/book"

// Tangle is the whole of what "babble tangle" does to one book: read
// it, refresh its dynamic blocks, write it back if that changed it,
// read it again, and tangle.
func Tangle(path string) error {
	panic("HOLE(9): read, refresh, write back if changed, read again, tangle")
}

// read is the front half of Tangle, separate because Tangle performs
// it twice on the same path.
func read(path string) (*book.Document, error) {
	panic("HOLE(9): parse, resolve, and run the lints the fence owns so far")
}

// replaceBook writes the book beside itself and renames it over, so a
// run that dies mid-write leaves the old book intact.
func replaceBook(path string, body []byte) error {
	panic("HOLE(9): write beside the book and rename over it")
}
