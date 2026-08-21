// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Package tangle turns a tree into files. Run is the whole pass:
// collect the eligible blocks, resolve their targets, assemble each
// file with its banner, and write only what changed.
package tangle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bilus/babble/internal/book"
)

func Run(d *book.Document) error {
	comments, err := loadCommentTable(d.Path)
	if err != nil {
		return err
	}
	units, err := collect(d)
	if err != nil {
		return err
	}
	for _, u := range units {
		content, err := assemble(d, u, comments)
		if err != nil {
			return err
		}
		if err := writeTarget(d.Path, u, content); err != nil {
			return err
		}
	}
	return nil
}

// It keeps the first error and walks the whole tree anyway. That
// looks like a bug and is the opposite: collect writes nothing, so
// finishing the walk costs nothing, and returning the error at the
// end means Run stops before its write loop starts. Every refusal
// raised here therefore precedes every write. The refusals raised
// later, while a body is being assembled, do not have that property,
// and a failure on the third target leaves the first two on disk.
type unit struct {
	target string
	lang   string
	blocks []*book.SrcBlock
}

func collect(d *book.Document) ([]unit, error) {
	var units []unit
	var firstErr error
	at := map[string]int{}
	fail := func(format string, args ...any) {
		if firstErr == nil {
			firstErr = fmt.Errorf(format, args...)
		}
	}
	d.Walk(func(n book.Node) bool {
		switch n := n.(type) {
		case *book.Headline:
			return !n.Commented && !n.Archived
		case *book.SrcBlock:
			if n.Lang == "" {
				fail("%s:%d: src block without a language", d.Path, n.Line)
				return true
			}
			if n.Params.Tangle == "no" || n.Params.Tangle == "" {
				return true
			}
			target, err := targetFor(d, n)
			if err != nil {
				fail("%s", err)
				return true
			}
			if sameFile(d.Path, target) {
				fail("%s:%d: refusing to tangle into the book itself: %s", d.Path, n.Line, target)
				return true
			}
			i, seen := at[target]
			if !seen {
				i = len(units)
				at[target] = i
				units = append(units, unit{target: target, lang: n.Lang})
			}
			if units[i].lang != n.Lang {
				fail("%s: blocks in %s and %s; one file takes one comment syntax",
					target, units[i].lang, n.Lang)
				return true
			}
			units[i].blocks = append(units[i].blocks, n)
		}
		return true
	})
	return units, firstErr
}

// A target is a path, a derived name, or nothing. The derived name
// is the book's own, with the language's extension, and the map of
// extensions is org's: two entries and the language's own name for
// everything else, which is why a python block asking for a derived
// name lands in a file named for the language.
func targetFor(d *book.Document, b *book.SrcBlock) (string, error) {
	dir := filepath.Dir(d.Path)
	switch target := b.Params.Tangle; target {
	case "yes":
		base := filepath.Base(d.Path)
		base = strings.TrimSuffix(base, filepath.Ext(base))
		return filepath.Join(dir, base+"."+langExt(b.Lang)), nil
	default:
		if filepath.IsAbs(target) {
			return target, nil
		}
		return filepath.Join(dir, target), nil
	}
}

func langExt(lang string) string {
	switch lang {
	case "emacs-lisp", "elisp":
		return "el"
	}
	return lang
}

// bodyText is the body pipeline in the contract's order; the second
// dedent and the final trim run after noweb expansion, exactly as
// the oracle orders them. A block that preserves its indentation
// skips both dedents and keeps its lead through the trim, which is
// the one branch in the pipeline and the reason the trim takes a
// flag.
func bodyText(d *book.Document, b *book.SrcBlock) (string, error) {
	body := unescapeCommas(string(d.Source[b.Body.Start:b.Body.End]))
	body = strings.TrimSuffix(body, "\n")
	keepLead := b.Params.PreserveIndent
	if !keepLead {
		body = dedent(body)
	}
	if b.Params.Noweb {
		expanded, err := expandNoweb(d, b, body)
		if err != nil {
			return "", err
		}
		body = expanded
	}
	if keepLead {
		return orgTrim(body, keepLead), nil
	}
	return orgTrim(dedent(body), keepLead), nil
}

// "\n\n    package main\n\n"  keeping the lead  ->  "    package main"
//   "\n\n    package main\n\n"  otherwise         ->  "package main"
func orgTrim(s string, keepLead bool) string {
	s = strings.TrimRight(s, orgSpace)
	if !keepLead {
		return strings.TrimLeft(s, orgSpace)
	}
	for {
		nl := strings.IndexByte(s, '\n')
		if nl < 0 || strings.Trim(s[:nl], " \t") != "" {
			return s
		}
		s = s[nl+1:]
	}
}

const orgSpace = " \t\n"

// Org's own escape: a line that would otherwise open a block or a
// heading wears a comma, and tangling takes one comma off. The run
// may be longer than one, and only the comma nearest the marker
// belongs to the escape.
func unescapeCommas(body string) string {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		j := 0
		for j < len(line) && (line[j] == ' ' || line[j] == '\t') {
			j++
		}
		k := j
		for k < len(line) && line[k] == ',' {
			k++
		}
		if k == j {
			continue
		}
		if rest := line[k:]; strings.HasPrefix(rest, "*") || strings.HasPrefix(rest, "#+") {
			lines[i] = line[:k-1] + rest
		}
	}
	return strings.Join(lines, "\n")
}

// That is org's behaviour, and it is worth knowing why, because the
// obvious guess is wrong. Org cuts with ~move-to-column n t~, and the
// forcing flag does not split a tab: it pads with spaces in front of
// the tab until point sits at column n. The delete that follows takes
// exactly the padding it just added, so the line comes out as it went
// in. A tab is only removed when the measure lands on its far edge.
func dedent(body string) string {
	lines := strings.Split(body, "\n")
	measure := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if c := indentColumns(line); measure < 0 || c < measure {
			measure = c
		}
	}
	if measure <= 0 {
		return body
	}
	for i, line := range lines {
		lines[i] = cutColumns(line, measure)
	}
	return strings.Join(lines, "\n")
}

func indentColumns(line string) int {
	col := 0
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case ' ':
			col++
		case '\t':
			col += 8 - col%8
		default:
			return col
		}
	}
	return col
}

func cutColumns(line string, n int) string {
	col, i := 0, 0
	for i < len(line) && col < n {
		switch line[i] {
		case ' ':
			col++
		case '\t':
			col += 8 - col%8
		default:
			return line[i:]
		}
		i++
	}
	if col > n {
		return line
	}
	return line[i:]
}

// commentFor and wrapComment split the doc-comment story where the
// driver splits it: what text the extent and docstring rule choose,
// then how the language's comment syntax wears it.
func commentFor(d *book.Document, b *book.SrcBlock, comments commentTable) (string, error) {
	panic("HOLE(7): extent between the anchors, then the docstring rule")
}

func wrapComment(lang, text string, comments commentTable) (string, error) {
	panic("HOLE(7): wrap in comments.prefix(lang), blank lines as a bare marker")
}

// expandNoweb resolves references against the tree: a named block
// wins, else the noweb-ref concatenation, prefixes replicate onto
// every expansion line, recursion allowed, cycles and misses loud.
func expandNoweb(d *book.Document, b *book.SrcBlock, body string) (string, error) {
	panic("HOLE(8): named block first, else noweb-ref concatenation; cycles and misses error")
}

// A prefix that is present but empty is used as it stands. The elisp
// lookup treats the empty string as an answer; the Go habit of testing
// whether a value is empty would fall through to the hash.
type commentTable map[string]string

// The rows the toolchain's own books need. A language absent from
// the table takes the hash, which comments shells, Python, Ruby and
// make; the languages that would read a hash as code name themselves
// here or in the project's own file.
var builtinComments = commentTable{
	"go": "//", "js": "//",
	"emacs-lisp": ";;", "elisp": ";;",
	"lua": "--",
}

func loadCommentTable(bookPath string) (commentTable, error) {
	t := commentTable{}
	for lang, prefix := range builtinComments {
		t[lang] = prefix
	}
	path := filepath.Join(filepath.Dir(bookPath), ".lit", "comments.json")
	f, err := os.Open(path)
	if err != nil {
		return t, nil
	}
	defer f.Close()
	var raw map[string]any
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	for lang, v := range raw {
		prefix, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("%s: comment prefix for %s is not a string", path, lang)
		}
		t[lang] = prefix
	}
	return t, nil
}

// A prefix that is present wins even when it is empty, which is why
// this reads the second return value rather than testing for "".
func (t commentTable) prefix(lang string) string {
	if prefix, ok := t[lang]; ok {
		return prefix
	}
	return "#"
}

// A directory is created when any block in the unit asks for it, not
// when the first one does. Blocks sharing a target can disagree about
// ~:mkdirp~, and taking the first block's answer meant the file's
// directory depended on which block happened to come first.
func banner(lang string, comments commentTable) string {
	return comments.prefix(lang) + " Code generated from BOOK.org by make tangle. DO NOT EDIT.\n"
}

// The file is its blocks in order, a blank line between them unless
// one asks for none, and a newline after each. The banner and the
// blank line under it open the file, which is where the driver's
// hook puts them.
func assemble(d *book.Document, u unit, comments commentTable) ([]byte, error) {
	var out strings.Builder
	out.WriteString(banner(u.lang, comments))
	out.WriteString("\n")
	for i, b := range u.blocks {
		if i > 0 && b.Params.Padline {
			out.WriteString("\n")
		}
		if b.Params.Comments == "org" {
			comment, err := commentFor(d, b, comments)
			if err != nil {
				return nil, err
			}
			out.WriteString(comment)
		}
		body, err := bodyText(d, b)
		if err != nil {
			return nil, err
		}
		out.WriteString(body)
		out.WriteString("\n")
	}
	return []byte(out.String()), nil
}

func writeTarget(bookPath string, u unit, content []byte) error {
	mkdirp := false
	for _, b := range u.blocks {
		if b.Params.Mkdirp {
			mkdirp = true
			break
		}
	}
	if mkdirp {
		if err := os.MkdirAll(filepath.Dir(u.target), 0o755); err != nil {
			return err
		}
	}
	if old, err := os.ReadFile(u.target); err == nil && bytes.Equal(old, content) {
		return nil
	}
	return os.WriteFile(u.target, content, 0o644)
}

func sameFile(a, b string) bool {
	pa, err := filepath.Abs(a)
	if err != nil {
		return a == b
	}
	pb, err := filepath.Abs(b)
	if err != nil {
		return a == b
	}
	return pa == pb
}
