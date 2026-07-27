// Code generated from BOOK.org by make tangle. DO NOT EDIT.

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"babble": func() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) },
	})
}

// TestScripts skips while no fixture exists, so the skeleton's suite
// stays green with every body a hole; the first stage-2 fixture
// retires the skip on its own.
func TestScripts(t *testing.T) {
	dir := filepath.Join("testdata", "script")
	// Glob's only error is a malformed pattern, and this one is fixed.
	fixtures, _ := filepath.Glob(filepath.Join(dir, "*.txtar"))
	if len(fixtures) == 0 {
		t.Skip("no fixtures yet")
	}
	testscript.Run(t, testscript.Params{Dir: dir})
}
