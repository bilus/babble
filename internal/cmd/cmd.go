// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Package cmd holds what a command does once its arguments are read,
// so that the command line and the oracle harness run the same steps
// in the same order.
package cmd

import (
	"os"
	"path/filepath"

	"github.com/bilus/babble/internal/book"
	"github.com/bilus/babble/internal/dblock"
	"github.com/bilus/babble/internal/org"
	"github.com/bilus/babble/internal/tangle"
)

// Tangle is the whole of what "babble tangle" does to one book: read
// it, refresh its dynamic blocks, write it back if that changed it,
// read it again, and tangle.
func Tangle(path string) error {
	d, err := read(path)
	if err != nil {
		return err
	}
	src, changed, err := dblock.Refresh(d, func(name string) (string, error) {
		return tangle.NamedBody(d, name)
	})
	if err != nil {
		return err
	}
	if changed {
		if err := replaceBook(path, src); err != nil {
			return err
		}
		if d, err = read(path); err != nil {
			return err
		}
	}
	return tangle.Run(d)
}

// read is the front half of Tangle, separate because Tangle performs
// it twice on the same path.
func read(path string) (*book.Document, error) {
	d, err := org.Parse(path)
	if err != nil {
		return nil, err
	}
	if err := org.ResolveAll(d); err != nil {
		return nil, err
	}
	if err := org.LintQuotedDelimiters(d); err != nil {
		return nil, err
	}
	return d, nil
}

// replaceBook writes the book beside itself and renames it over, so a
// run that dies mid-write leaves the old book intact.
func replaceBook(path string, body []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".babble-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(body); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
