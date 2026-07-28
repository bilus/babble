// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Package tangle turns a tree into files. Run is the whole pass:
// collect the eligible blocks, resolve their targets, assemble each
// file with its banner, and write only what changed.
package tangle

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bilus/babble/internal/book"
)

func Run(d *book.Document) error {
	units, err := collect(d)
	if err != nil {
		return err
	}
	for _, u := range units {
		content, err := assemble(d, u)
		if err != nil {
			return err
		}
		if err := writeTarget(d.Path, u, content); err != nil {
			return err
		}
	}
	return nil
}

// A unit is one target file's worth of blocks, in document order;
// collect builds them in first-appearance order, skipping COMMENT
// and ARCHIVE subtrees, erroring on a languageless block.
type unit struct {
	target string
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
			if n.Params.Tangle == "no" {
				return true
			}
			target, err := targetFor(d, n)
			if err != nil {
				fail("%s", err)
				return true
			}
			i, seen := at[target]
			if !seen {
				i = len(units)
				at[target] = i
				units = append(units, unit{target: target})
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
// the oracle orders them.
func bodyText(d *book.Document, b *book.SrcBlock) (string, error) {
	body := unescapeCommas(string(d.Source[b.Body.Start:b.Body.End]))
	body = strings.TrimSuffix(body, "\n")
	if b.Params.PreserveIndent {
		return strings.TrimRight(body, " \t\n\r"), nil
	}
	body = dedent(body)
	if b.Params.Noweb {
		expanded, err := expandNoweb(d, b, body)
		if err != nil {
			return "", err
		}
		body = expanded
	}
	return strings.Trim(dedent(body), " \t\n\r"), nil
}

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

// Indentation is measured in columns, with a tab worth eight, and
// the common measure comes off every line. A blank line has no
// indentation to speak of and does not hold the measure down; a line
// flush with the margin does, and then nothing moves.
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
	return strings.Repeat(" ", col-n) + line[i:]
}

// commentFor and wrapComment split the doc-comment story where the
// driver splits it: what text the extent and docstring rule choose,
// then how the language's comment syntax wears it.
func commentFor(d *book.Document, b *book.SrcBlock) (string, error) {
	panic("HOLE(7): extent between the anchors, then the docstring rule")
}

func wrapComment(lang, text string) (string, error) {
	panic("HOLE(7): line-comment table per language, blank lines as a bare marker")
}

// expandNoweb resolves references against the tree: a named block
// wins, else the noweb-ref concatenation, prefixes replicate onto
// every expansion line, recursion allowed, cycles and misses loud.
func expandNoweb(d *book.Document, b *book.SrcBlock, body string) (string, error) {
	panic("HOLE(8): named block first, else noweb-ref concatenation; cycles and misses error")
}

// banner and writeTarget close the pass: the generated-file banner
// in the target's own comment syntax, parent directories on demand,
// a refusal to write the book itself, and no write at all when the
// bytes already match.
func banner(target string) string {
	prefix := "# "
	switch filepath.Ext(target) {
	case ".go":
		prefix = "// "
	case ".el":
		prefix = ";; "
	case ".lua":
		prefix = "-- "
	}
	return prefix + "Code generated from BOOK.org by make tangle. DO NOT EDIT.\n"
}

// The file is its blocks in order, a blank line between them unless
// one asks for none, and a newline after each. The banner and the
// blank line under it open the file, which is where the driver's
// hook puts them.
func assemble(d *book.Document, u unit) ([]byte, error) {
	var out strings.Builder
	out.WriteString(banner(u.target))
	out.WriteString("\n")
	for i, b := range u.blocks {
		if i > 0 && b.Params.Padline {
			out.WriteString("\n")
		}
		if b.Params.Comments == "org" {
			comment, err := commentFor(d, b)
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
	if sameFile(bookPath, u.target) {
		return fmt.Errorf("refusing to tangle into the book itself: %s", u.target)
	}
	if u.blocks[0].Params.Mkdirp {
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
