// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Package org is the frontend: everything that knows the input
// syntax lives here, and it hands the engine a finished tree. Parse
// reads a file; ParseBytes does the work, so tests and the refresh
// reparse can feed bytes directly.
package org

import (
	"bytes"
	"os"
	"strings"

	"github.com/bilus/babble/internal/book"
)

func Parse(path string) (*book.Document, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseBytes(src, path)
}

func ParseBytes(src []byte, path string) (*book.Document, error) {
	d := &book.Document{Path: path, Source: src,
		Keywords: map[string][]string{}, Todo: todoWords(src)}
	sc := &scanner{d: d, todo: map[string]bool{}}
	for _, w := range d.Todo {
		sc.todo[w] = true
	}
	sc.scan()
	return d, nil
}

// The scanner is a single pass over lines, with one exception made
// first: the todo words. Org applies a #+todo line to the whole
// file, wherever it sits, so a prescan collects the words before any
// headline is read. Each word drops its shortcut-and-logging suffix,
// and the bar that separates active from done states drops too,
// since the tangler only needs to recognize the words.
func todoWords(src []byte) []string {
	var words []string
	for off := 0; off < len(src); {
		text, next := line(src, off)
		if key, value, ok := keywordLine(text); ok {
			if key == "todo" || key == "seq_todo" || key == "typ_todo" {
				for _, f := range strings.Fields(value) {
					if f == "|" {
						continue
					}
					if j := strings.IndexByte(f, '('); j >= 0 {
						f = f[:j]
					}
					if f != "" {
						words = append(words, f)
					}
				}
			}
		}
		off = next
	}
	if words == nil {
		words = []string{"TODO", "DONE"}
	}
	return words
}

// line yields one line's text without its newline and the offset of
// the next line; the scanner and every lookahead share it.
func line(src []byte, off int) (string, int) {
	eol := bytes.IndexByte(src[off:], '\n')
	if eol < 0 {
		return string(src[off:]), len(src)
	}
	return string(src[off : off+eol]), off + eol + 1
}

// The scanner proper. Headlines and keyword lines are structural;
// everything else joins the open prose run, blank lines included,
// and a run flushes whenever structure interrupts it. A property
// drawer is recognized only directly under its headline, which is
// org's own rule.
type scanner struct {
	d     *book.Document
	todo  map[string]bool
	stack []*book.Headline
}

func (sc *scanner) scan() {
	src := sc.d.Source
	off, ln := 0, 1
	proseStart, proseLine := -1, 0
	flush := func(end int) {
		if proseStart >= 0 {
			sc.attach(&book.Prose{Line: proseLine,
				Text: book.Span{Start: proseStart, End: end}})
			proseStart = -1
		}
	}
	for off < len(src) {
		text, next := line(src, off)
		switch {
		case headlineLine(text):
			flush(off)
			h := sc.headline(text, off, ln)
			sc.attachHeadline(h)
			if props, dOff, dLn, ok := drawer(src, next, ln+1); ok {
				h.Properties = props
				next, ln = dOff, dLn-1
			}
		case isKeyword(text):
			flush(off)
			key, value, _ := keywordLine(text)
			k := &book.Keyword{Key: key, Value: value, Line: ln,
				Raw: book.Span{Start: off, End: off + len(text)}}
			sc.d.Keywords[key] = append(sc.d.Keywords[key], value)
			sc.attach(k)
		default:
			if proseStart < 0 {
				proseStart, proseLine = off, ln
			}
		}
		off = next
		ln++
	}
	flush(len(src))
}

// Attachment is a stack of open headlines: a new headline pops
// everything at its level or deeper and becomes a child of what
// remains, and every other node joins the innermost open headline.
// A fifteen-star badge is just a very deep entry in that stack.
func (sc *scanner) attach(n book.Node) {
	if len(sc.stack) == 0 {
		sc.d.Nodes = append(sc.d.Nodes, n)
		return
	}
	top := sc.stack[len(sc.stack)-1]
	top.Children = append(top.Children, n)
}

func (sc *scanner) attachHeadline(h *book.Headline) {
	for len(sc.stack) > 0 && sc.stack[len(sc.stack)-1].Level >= h.Level {
		sc.stack = sc.stack[:len(sc.stack)-1]
	}
	sc.attach(h)
	sc.stack = append(sc.stack, h)
}

// Headline anatomy, in org's order: stars and one space, an optional
// todo word from the book's own set, an optional [#A] priority, an
// optional COMMENT marker, the title, and trailing tags. The
// ARCHIVE tag sets Archived on the way through.
func headlineLine(text string) bool {
	i := 0
	for i < len(text) && text[i] == '*' {
		i++
	}
	return i > 0 && i < len(text) && text[i] == ' '
}

func (sc *scanner) headline(text string, off, ln int) *book.Headline {
	stars := strings.IndexByte(text, ' ')
	h := &book.Headline{Level: stars, Line: ln,
		AfterStars: off + stars + 1,
		Head:       book.Span{Start: off, End: off + len(text)}}
	rest := strings.TrimLeft(text[stars+1:], " \t")
	if w, r, ok := firstWord(rest); ok && sc.todo[w] {
		h.Todo = w
		rest = r
	}
	if len(rest) >= 4 && strings.HasPrefix(rest, "[#") && rest[3] == ']' {
		rest = strings.TrimLeft(rest[4:], " \t")
	}
	if rest == "COMMENT" || strings.HasPrefix(rest, "COMMENT ") {
		h.Commented = true
		rest = strings.TrimLeft(strings.TrimPrefix(rest, "COMMENT"), " \t")
	}
	rest, h.Tags = splitTags(rest)
	h.Title = strings.TrimSpace(rest)
	for _, tag := range h.Tags {
		if tag == "ARCHIVE" {
			h.Archived = true
		}
	}
	return h
}

func firstWord(s string) (word, rest string, ok bool) {
	i := strings.IndexAny(s, " \t")
	if i < 0 {
		return s, "", s != ""
	}
	return s[:i], strings.TrimLeft(s[i:], " \t"), i > 0
}

// Tags are the trailing colon-fenced run; each tag is alphanumerics
// plus the four org extras. Anything that fails the shape stays in
// the title.
func splitTags(s string) (string, []string) {
	t := strings.TrimRight(s, " \t")
	if !strings.HasSuffix(t, ":") {
		return s, nil
	}
	i := strings.LastIndexAny(t, " \t")
	cand := t[i+1:]
	if len(cand) < 2 || cand[0] != ':' {
		return s, nil
	}
	parts := strings.Split(cand[1:len(cand)-1], ":")
	for _, p := range parts {
		if p == "" || !tagWord(p) {
			return s, nil
		}
	}
	return t[:i+1], parts
}

func tagWord(s string) bool {
	for _, c := range s {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z',
			c >= '0' && c <= '9', c == '_', c == '@', c == '#', c == '%':
		default:
			return false
		}
	}
	return true
}

// A keyword line is #+key: value, the key one colon-free word; the
// begin line of a block has a space before its first colon, so it
// falls through to prose until stage 3 claims it.
func keywordLine(text string) (key, value string, ok bool) {
	t := strings.TrimLeft(text, " \t")
	if !strings.HasPrefix(t, "#+") {
		return "", "", false
	}
	rest := t[2:]
	i := strings.IndexByte(rest, ':')
	if i <= 0 {
		return "", "", false
	}
	for _, c := range rest[:i] {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z',
			c >= '0' && c <= '9', c == '_', c == '-':
		default:
			return "", "", false
		}
	}
	return strings.ToLower(rest[:i]), strings.TrimSpace(rest[i+1:]), true
}

func isKeyword(text string) bool {
	_, _, ok := keywordLine(text)
	return ok
}

// The property drawer: a :PROPERTIES: line directly under the
// heading, :KEY: value lines, a closing :END:. A malformed drawer is
// not a drawer at all, and its lines stay prose, which is also what
// org does.
func drawer(src []byte, off, ln int) (map[string]string, int, int, bool) {
	text, next := line(src, off)
	if !strings.EqualFold(strings.TrimSpace(text), ":PROPERTIES:") {
		return nil, off, ln, false
	}
	props := map[string]string{}
	lines := 1
	for next < len(src) {
		text, n2 := line(src, next)
		lines++
		t := strings.TrimSpace(text)
		if strings.EqualFold(t, ":END:") {
			return props, n2, ln + lines, true
		}
		if len(t) < 2 || t[0] != ':' {
			return nil, off, ln, false
		}
		j := strings.IndexByte(t[1:], ':')
		if j < 0 {
			return nil, off, ln, false
		}
		key := strings.ToLower(t[1 : 1+j])
		props[key] = strings.TrimSpace(t[2+j:])
		next = n2
	}
	return nil, off, ln, false
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
