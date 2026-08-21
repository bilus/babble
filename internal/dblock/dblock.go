// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Package dblock owns the book side: the two dynamic-block writers and
// the splice that puts what they produce back into the book's bytes.
package dblock

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bilus/babble/internal/book"
)

// NamedBody hands back the body of a block by name, shaped the way
// org-babel hands it over. dblock takes it as an argument because
// the shaping is the tangler's rule and a second copy of it could
// drift away from the file the diff describes.
type NamedBody func(name string) (string, error)

// Refresh splices a fresh interior into every dynamic block and
// returns the book's new bytes together with whether they differ from
// the old ones, so the caller writes the file only when there is
// something to write.
func Refresh(d *book.Document, body NamedBody) ([]byte, bool, error) {
	var blocks []*book.DynamicBlock
	d.Walk(func(n book.Node) bool {
		if db, ok := n.(*book.DynamicBlock); ok {
			blocks = append(blocks, db)
		}
		return true
	})
	if len(blocks) == 0 {
		return d.Source, false, nil
	}
	src := d.Source
	file := filepath.Base(d.Path)
	out := make([]byte, 0, len(src))
	cursor := 0
	for _, db := range blocks {
		out = append(out, src[cursor:db.Interior.Start]...)
		staged := make([]byte, 0, len(out)+1+len(src)-db.Interior.End)
		staged = append(staged, out...)
		staged = append(staged, '\n')
		staged = append(staged, src[db.Interior.End:]...)
		text, err := write(db, staged, bytes.Count(out, newline)+1, file, body)
		if err != nil {
			return nil, false, fmt.Errorf("%s:%d: %v", d.Path, db.Line, err)
		}
		out = append(out, text...)
		out = append(out, '\n')
		cursor = db.Interior.End
	}
	out = append(out, src[cursor:]...)
	if out[len(out)-1] != '\n' {
		out = append(out, '\n')
	}
	return out, !bytes.Equal(out, src), nil
}

var newline = []byte("\n")

// write dispatches an interior to the writer its opener names, or
// refuses a name no writer answers to.
func write(db *book.DynamicBlock, staged []byte, self int, file string, body NamedBody) (string, error) {
	switch db.Name {
	case "plan-toc":
		return planToc(staged, self, file), nil
	case "src-block-diff":
		return srcBlockDiff(db.Args, body)
	}
	return "", fmt.Errorf("no dynamic-block writer named %s; the two names are plan-toc and src-block-diff", db.Name)
}

// planToc renders the jump list: one line-number link per plan-markup
// site in the staged book.
func planToc(staged []byte, self int, file string) string {
	type site struct {
		line  int
		label string
	}
	var hits []site
	for i, text := range strings.Split(string(staged), "\n") {
		if label := planSite(text); label != "" {
			hits = append(hits, site{i + 1, label})
		}
	}
	var b strings.Builder
	shift := len(hits)
	for _, h := range hits {
		n := h.line
		if n > self {
			n += shift
		}
		fmt.Fprintf(&b, "- [[file:%s::%d][%d]] %s\n", file, n, n, h.label)
	}
	return b.String()
}

// The census patterns, ported from the driver. Every one folds case
// because Emacs folds case, and the folding is not cosmetic: it
// turns [A-Z] in the entry pattern into any letter, which is what
// makes a lowercase heading an entry.
//
// The CriticMarkup openers are written as character classes rather
// than escaped literals. The weave rewrites CriticMarkup by running
// over the book's whole text, code bodies included, so a pattern
// that spelled its own sigil would be read as an edit to the
// sentence around it and would swallow the rest of the page.
var (
	planSiteRe   = regexp.MustCompile(`(?i)^\*+ (?:PLANNED|FILLING|HOLE|LIVE|DROPPED) |^\*{15,} |HOLE\(|\{[+][+]|\{[-][-]|\{[~][~]|\{[>][>]|^#\+begin: src-block-diff|--planned`)
	planTodoRe   = regexp.MustCompile(`(?i)^#\+todo:`)
	planBadgeEnd = regexp.MustCompile(`(?i)^\*{15,} END`)
	planBadgeRe  = regexp.MustCompile(`(?i)^\*{15,} (.*)`)
	planEntryRe  = regexp.MustCompile(`(?i)^\*+ ([A-Z]+ .*)`)
	planDiffRe   = regexp.MustCompile(`(?i)^#\+begin: src-block-diff`)
	planHoleRe   = regexp.MustCompile(`(?i)HOLE\(([0-9]+)\): ([^"]*)`)
	planInsRe    = regexp.MustCompile(`\{[+][+](.*)`)
	planDelRe    = regexp.MustCompile(`\{[-][-](.*)`)
	planRepRe    = regexp.MustCompile(`\{[~][~](.*)`)
	planNoteRe   = regexp.MustCompile(`\{[>][>](.*)`)
	planTwinRe   = regexp.MustCompile(`(?i)#\+name: (.*)--planned`)
)

// planSite answers what a line is called in the index, or nothing
// when the line is not a site. A line has to match the census
// pattern first: the naming rules below are wider than it is, and
// the driver applies them only to lines the census already found.
func planSite(text string) string {
	if !planSiteRe.MatchString(text) {
		return ""
	}
	if planTodoRe.MatchString(text) || planBadgeEnd.MatchString(text) {
		return ""
	}
	if m := planBadgeRe.FindStringSubmatch(text); m != nil {
		return "badge: " + planStrip(m[1])
	}
	if m := planEntryRe.FindStringSubmatch(text); m != nil {
		return "entry: " + planStrip(m[1])
	}
	if planDiffRe.MatchString(text) {
		return "generated diff"
	}
	if m := planHoleRe.FindStringSubmatch(text); m != nil {
		return fmt.Sprintf("hole %s: %s", m[1], planStrip(m[2]))
	}
	for _, kind := range []struct {
		re    *regexp.Regexp
		label string
	}{
		{planInsRe, "insertion: "},
		{planDelRe, "deletion: "},
		{planRepRe, "replacement: "},
		{planNoteRe, "note: "},
	} {
		if m := kind.re.FindStringSubmatch(text); m != nil {
			return kind.label + planStrip(m[1])
		}
	}
	if m := planTwinRe.FindStringSubmatch(text); m != nil {
		return "twin of " + m[1]
	}
	return ""
}

var (
	planLinkRe = regexp.MustCompile(`\[\[[^]]*\]\[([^]]*)\]\]`)
	planTagRe  = regexp.MustCompile(`\s+:[a-zA-Z0-9_@:]+:$`)
)

// planStrip is the driver's label cleanup: trim, replace org links
// with their descriptions, drop a trailing tag block, and cut to
// sixty characters with an ellipsis. The count is characters and
// not bytes, which is what the driver counts.
func planStrip(s string) string {
	s = strings.TrimSpace(s)
	s = planLinkRe.ReplaceAllString(s, "$1")
	s = planTagRe.ReplaceAllString(s, "")
	if r := []rune(s); len(r) > 60 {
		return string(r[:57]) + "..."
	}
	return s
}

// srcBlockDiff renders a unified diff of two named blocks' bodies, or
// the words "no differences".
// a body line that would end the generated block or start a dynamic
// block, seen after diff has prefixed it with a space
var dangerousLine = regexp.MustCompile(`(?i)^[ \t]*#\+(end|begin:)`)

func srcBlockDiff(args string, body NamedBody) (string, error) {
	p, err := dblockArgs(args)
	if err != nil {
		return "", err
	}
	old, err := body(p["old"])
	if err != nil {
		return "", fmt.Errorf("src-block-diff :old %s: %v", p["old"], err)
	}
	new, err := body(p["new"])
	if err != nil {
		return "", fmt.Errorf("src-block-diff :new %s: %v", p["new"], err)
	}
	dir, err := os.MkdirTemp("", "src-block-diff")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)
	oldPath := filepath.Join(dir, "old")
	newPath := filepath.Join(dir, "new")
	if err := os.WriteFile(oldPath, []byte(old+"\n"), 0o644); err != nil {
		return "", err
	}
	if err := os.WriteFile(newPath, []byte(new+"\n"), 0o644); err != nil {
		return "", err
	}
	out, err := exec.Command("diff", "-u", oldPath, newPath).CombinedOutput()
	if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
		err = nil
	}
	if err != nil {
		return "", fmt.Errorf("diff -u: %v", err)
	}
	if len(out) == 0 {
		return "no differences", nil
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) < 3 {
		return "", fmt.Errorf("diff -u printed %d lines, too few to hold its own header", len(lines))
	}
	hunks := lines[2:]
	for i, text := range hunks {
		if dangerousLine.MatchString(text) {
			return "", fmt.Errorf("the diff of %s and %s carries %q, which would end the generated block or open a dynamic block inside it; the two bodies cannot be shown this way", p["old"], p["new"], strings.TrimSpace(hunks[i]))
		}
	}
	return "#+begin_src diff\n" + strings.Join(hunks, "\n") + "#+end_src", nil
}

// dblockArgs reads the opener's arguments as a keyword-to-name map.
var nameToken = regexp.MustCompile("^[^\\s:\"'()\\[\\]`]+$")

func dblockArgs(args string) (map[string]string, error) {
	out := map[string]string{}
	fields := strings.Fields(args)
	for i := 0; i < len(fields); i += 2 {
		key := fields[i]
		if len(key) < 2 || key[0] != ':' {
			return nil, fmt.Errorf("dynamic-block argument %s is not a keyword or a name", key)
		}
		if i+1 == len(fields) {
			return nil, fmt.Errorf("dynamic-block keyword %s has no value", key)
		}
		value := fields[i+1]
		if !nameToken.MatchString(value) {
			return nil, fmt.Errorf("dynamic-block argument %s is not a keyword or a name", value)
		}
		out[key[1:]] = value
	}
	return out, nil
}
