// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Package oracletest runs a fixture's books through both engines and
// compares what each run wrote. It reads and never writes: a golden
// is minted by hand until the stage that mints them lands.
package oracletest

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/bilus/babble/internal/cmd"
	"golang.org/x/tools/txtar"
)

func Enabled() bool {
	return os.Getenv("ORACLE") == "1"
}

// A refusal needs no skip. A fixture asserts one with ~! exec~, and
// only a plain ~exec babble tangle~ names a book to compare, so the
// books babble is expected to refuse are left out by the same rule
// that finds the others.
var (
	tangleLine = regexp.MustCompile(`(?m)^exec babble tangle (\S+)`)
	skipLine   = regexp.MustCompile(`(?m)^# oracle: skip (.*)$`)
)

// books returns the books a fixture's script tangles, and the ones
// its skip line excuses.
func books(script string) (run []string, excusedAll bool) {
	excused := map[string]bool{}
	for _, m := range skipLine.FindAllStringSubmatch(script, -1) {
		named := 0
		for _, word := range strings.Fields(m[1]) {
			if !strings.HasSuffix(word, ".org") {
				break
			}
			excused[word] = true
			named++
		}
		if named == 0 {
			excusedAll = true
		}
	}
	seen := map[string]bool{}
	for _, m := range tangleLine.FindAllStringSubmatch(script, -1) {
		if seen[m[1]] || excused[m[1]] {
			continue
		}
		seen[m[1]] = true
		run = append(run, m[1])
	}
	return run, excusedAll
}

// A fixture whose skip line has no books after it excuses the whole
// fixture, which is how the two fixtures that already carry one are
// written.
func CrossCheck(t *testing.T, fixture string) {
	t.Helper()
	archive, err := txtar.ParseFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	run, excusedAll := books(string(archive.Comment))
	if excusedAll {
		t.Skip("the fixture excuses itself from the oracle")
	}
	if len(run) == 0 {
		t.Skip("no book to tangle")
	}
	oracleDir, babbleDir := t.TempDir(), t.TempDir()
	for _, dir := range []string{oracleDir, babbleDir} {
		for _, f := range archive.Files {
			write(t, dir, f.Name, f.Data)
		}
	}
	tangled := 0
	for _, b := range run {
		tangled += runOracle(t, oracleDir, b)
		runBabble(t, babbleDir, b)
	}
	if tangled == 0 {
		t.Fatalf("the oracle tangled no blocks from %v; two empty trees agree, which proves nothing", run)
	}
	compareAdditions(t, oracleDir, babbleDir)
}

func write(t *testing.T, dir, name string, body []byte) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
}

// Three things can go wrong and they have three different causes.
// Emacs may be absent, which is the machine's fault and not babble's.
// Emacs may run and fail, in which case its own message on stderr is
// the useful half. Or it may run and tangle nothing, which reads as
// success and is not: the count it prints is what tells them apart.
var tangledCount = regexp.MustCompile(`Tangled (\d+) code block`)

func runOracle(t *testing.T, dir, book string) int {
	t.Helper()
	if _, err := exec.LookPath("emacs"); err != nil {
		t.Fatal("ORACLE=1 is set and emacs is not on PATH; the cross-check cannot run")
	}
	driver, err := filepath.Abs(filepath.Join("lit", "tangle.el"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("emacs", "-Q", "--batch", "-l", driver,
		"--eval", fmt.Sprintf("(lit-tangle %q)", book))
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("the oracle failed on %s: %v\n%s", book, err, out)
	}
	total := 0
	for _, m := range tangledCount.FindAllSubmatch(out, -1) {
		n := 0
		fmt.Sscanf(string(m[1]), "%d", &n)
		total += n
	}
	return total
}

func runBabble(t *testing.T, dir, book string) {
	t.Helper()
	if err := cmd.Tangle(filepath.Join(dir, book)); err != nil {
		t.Fatalf("babble failed on %s where the oracle did not: %v", book, err)
	}
}

// The comparison walks both trees and reports every path where the
// two runs disagree: a file one wrote and the other did not, or the
// same path with different bytes. Only bytes are compared, since
// babble writes a fixed mode where the driver takes the umask, and a
// mode difference would be a difference neither engine is wrong
// about.
func compareAdditions(t *testing.T, oracleDir, babbleDir string) {
	t.Helper()
	oracle := added(t, oracleDir)
	babbled := added(t, babbleDir)
	paths := map[string]bool{}
	for p := range oracle {
		paths[p] = true
	}
	for p := range babbled {
		paths[p] = true
	}
	var names []string
	for p := range paths {
		names = append(names, p)
	}
	sort.Strings(names)
	for _, p := range names {
		want, inOracle := oracle[p]
		got, inBabble := babbled[p]
		switch {
		case !inBabble:
			t.Errorf("%s: the oracle wrote it and babble did not", p)
		case !inOracle:
			t.Errorf("%s: babble wrote it and the oracle did not", p)
		case !bytes.Equal(want, got):
			t.Errorf("%s: the two engines disagree\n oracle: %q\n babble: %q", p, want, got)
		}
	}
}

// added is every file in the tree after a run, keyed by its path. It
// does not subtract the files the fixture supplied, on purpose: a
// book is one of those and a refresh rewrites it, so leaving them in
// is what puts the book's own bytes under the comparison.
func added(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	out := map[string][]byte{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = body
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
