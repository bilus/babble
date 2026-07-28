// Code generated from BOOK.org by make tangle. DO NOT EDIT.

package main

import (
	"os"
	"path/filepath"
	"strings"
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

// TestFixtureTable keeps the fixtures chapter honest. The table
// there is written by hand, which is worth it for a column that says
// what a fixture is for, and worthless if it quietly stops matching
// the directory.
func TestFixtureTable(t *testing.T) {
	book, err := os.ReadFile("BOOK.org")
	if err != nil {
		t.Fatal(err)
	}
	listed := map[string]bool{}
	for _, line := range strings.Split(string(book), "\n") {
		cells := strings.Split(line, "|")
		if len(cells) < 3 {
			continue
		}
		if name := strings.TrimSpace(cells[1]); strings.HasSuffix(name, ".txtar") {
			listed[name] = true
		}
	}
	found, err := filepath.Glob(filepath.Join("testdata", "script", "*.txtar"))
	if err != nil {
		t.Fatal(err)
	}
	onDisk := map[string]bool{}
	for _, path := range found {
		onDisk[filepath.Base(path)] = true
	}
	for name := range onDisk {
		if !listed[name] {
			t.Errorf("%s is on disk and missing from the chapter's table", name)
		}
	}
	for name := range listed {
		if !onDisk[name] {
			t.Errorf("%s is in the chapter's table and not on disk", name)
		}
	}
}
