// Code generated from BOOK.org by make tangle. DO NOT EDIT.

package main

import (
	"io"
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

// TestScripts skips when no fixture exists. That mattered when the
// suite was empty and every body was a hole; the directory has been
// full since the first fixtures landed, so the skip has not fired in
// a long time. It stays because a checkout with an empty testdata
// directory should say so rather than pass silently.
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

// The lit.mk comparison is the only check that the embedded filesystem
// carries real bytes. An embed that matched nothing would still
// compile, still walk, and still write no error, so without this the
// first sign of trouble would be a new project with an empty lit/.
func TestInitWrites(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "fresh")
	p, err := newProject(dir, "example.com/fresh", io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range []func(*project) error{
		(*project).writeDevbox, (*project).copyLit, (*project).writeMakefile,
	} {
		if err := step(p); err != nil {
			t.Fatal(err)
		}
	}
	for _, want := range []string{
		"devbox.json", "Makefile", ".gitignore", "lit/lit.mk",
		"lit/tangle.el", "lit/preprocess.py", "lit/templates/go/BOOK.org",
	} {
		if _, err := os.Stat(filepath.Join(dir, want)); err != nil {
			t.Errorf("%s: %v", want, err)
		}
	}
	got, err := os.ReadFile(filepath.Join(dir, "lit/lit.mk"))
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("lit/lit.mk")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Error("the embedded lit.mk is not the one in lit/")
	}
	if _, err := os.Stat(filepath.Join(dir, "lit/BOOK.org")); err == nil {
		t.Error("lit/BOOK.org was embedded; a consumer never tangles it")
	}
}

// A module nobody names is named after the directory, which is the one
// guess that is never wrong about where the code sits. And every write
// refuses to overwrite, even though newProject has already cleared the
// three files that matter, because a writer that trusts its caller is
// a writer that truncates somebody's file the first time it is reused.
func TestInitDefaults(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "mybook")
	p, err := newProject(dir, "", io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if p.module != "mybook" {
		t.Errorf("module = %q, want mybook", p.module)
	}
	if err := p.writeDevbox(); err != nil {
		t.Fatal(err)
	}
	if err := p.writeDevbox(); err == nil {
		t.Error("writing devbox.json twice succeeded; it must never clobber")
	}
}
