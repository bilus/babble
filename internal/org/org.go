// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Package org is the frontend: everything that knows the input
// syntax lives here, and it hands the engine a finished tree. Parse
// reads a file; ParseBytes does the work, so tests and the refresh
// reparse can feed bytes directly.
package org

import (
	"bytes"
	"fmt"
	"os"
	"sort"
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
	p := &parser{src: src, path: path, d: d,
		todo: map[string]bool{}, lines: lineStarts(src)}
	for _, w := range d.Todo {
		p.todo[w] = true
	}
	nodes, err := p.parseElements(0, len(src), modeFirstSection)
	if err != nil {
		return nil, err
	}
	d.Nodes = nodes
	return d, nil
}

// The parser holds the source, a line index and an offset, which is
// the buffer position org calls point. Every element parser reads
// the line at that offset when it needs to decide something, which
// is what keeps the byte anchors exact: the parser always knows
// where it is looking.
type parser struct {
	src   []byte
	path  string
	d     *book.Document
	todo  map[string]bool
	lines []int // offset of every line start, for line numbers
	off   int   // point
}

func lineStarts(src []byte) []int {
	starts := []int{0}
	for i, b := range src {
		if b == '\n' {
			starts = append(starts, i+1)
		}
	}
	return starts
}

func (p *parser) lineNum(off int) int {
	return sort.SearchInts(p.lines, off+1)
}

func (p *parser) line(off int) (string, int) { return line(p.src, off) }

func line(src []byte, off int) (string, int) {
	if i := bytes.IndexByte(src[off:], '\n'); i >= 0 {
		return string(src[off : off+i]), off + i + 1
	}
	return string(src[off:]), len(src)
}

// Mode is the context a parent hands its children. Org threads seven
// of them; the three below are the ones this subset can reach, and
// the dispatcher branches on the third. Org's planning mode has no
// counterpart here, because a planning line between a heading and
// its drawer is not recognized.
type mode int

const (
	modeFirstSection mode = iota
	modeSection
	modePropertyDrawer
)

// An element parser returns the node, and, when the element holds
// other elements, the byte range they live in and the way to adopt
// them. A verbatim block leaves both empty, so the loop cannot
// recurse into text that must stay text. Properties ride here too,
// because a drawer is an element in org's model but a field of the
// headline in ours.
type parsed struct {
	node      book.Node
	props     map[string]string // a property drawer's payload
	contents  book.Span         // where child elements live
	adopt     func([]book.Node)
	childMode mode
}

// parseElements is the loop from [[#the-loop][The loop]]: identify what starts
// here, recurse when it holds elements, jump to its end, repeat. The
// bound matters twice over. Elements are parsed within it, and a
// non-headline element gets the tighter bound of the next headline,
// which is org's section, and the reason a block cannot reach past
// a heading to find its terminator.
func (p *parser) parseElements(beg, end int, m mode) ([]book.Node, error) {
	var out []book.Node
	for off := beg; off < end; {
		p.off = off
		text, _ := p.line(off)
		limit := end
		if starRun(text) == 0 {
			limit = p.nextHeadline(off, end, anyLevel)
		}
		el, next, err := p.currentElement(limit, m)
		if err != nil {
			return nil, err
		}
		if el.node != nil {
			if el.adopt != nil && el.contents.End > el.contents.Start {
				kids, err := p.parseElements(el.contents.Start, el.contents.End, el.childMode)
				if err != nil {
					return nil, err
				}
				el.adopt(kids)
			}
			out = append(out, el.node)
		}
		off = next
		m = modeSection
	}
	return out, nil
}

// currentElement is the dispatcher, and its order is the table in
// [[#the-dispatch][The dispatch]]: the mode branch, then a star run, then the
// affiliated name line, then the block delimiters, then the keyword
// shape, then a paragraph for everything left. The first branch that
// matches wins, which is why the order is the contract and not a
// detail.
func (p *parser) currentElement(limit int, m mode) (parsed, int, error) {
	start := p.off
	text, next := p.line(start)
	if starRun(text) > 0 {
		return p.headlineParser(text, next, limit)
	}
	if m == modePropertyDrawer {
		if props, after, ok := p.propertyDrawerParser(start, limit); ok {
			return parsed{props: props}, after, nil
		}
	}
	aff := affiliated{begin: start}
	if key, value, ok := keywordLine(text); ok && key == kwName && next < limit {
		if t2, n2 := p.line(next); opensBlock(t2) {
			aff.name = value
			p.off, text, next = next, t2, n2
		}
	}
	if name, rest, _, ok := blockDelim(text, "begin"); ok {
		switch name {
		case "src":
			return p.srcBlockParser(rest, next, limit, aff)
		case "example", "export", "comment", "verse":
			return p.verbatimBlockParser(name, next, limit, aff)
		default:
			return p.greaterBlockParser(name, next, limit, aff)
		}
	}
	if key, value, ok := keywordLine(text); ok {
		if key == kwBegin {
			return p.dynamicBlockParser(value, next, limit, aff)
		}
		return p.keywordParser(key, value, text, next)
	}
	return p.paragraphParser(limit)
}

const anyLevel = 1 << 30

func opensBlock(text string) bool {
	if _, _, _, ok := blockDelim(text, "begin"); ok {
		return true
	}
	key, _, ok := keywordLine(text)
	return ok && key == kwBegin
}

func startsElement(text string) bool {
	if starRun(text) > 0 || opensBlock(text) {
		return true
	}
	_, _, ok := keywordLine(text)
	return ok
}

func (p *parser) nextHeadline(from, limit, maxLevel int) int {
	for off := from; off < limit; {
		text, next := p.line(off)
		if lvl := starRun(text); lvl > 0 && lvl <= maxLevel {
			return off
		}
		off = next
	}
	return limit
}

func (p *parser) findEnd(from, limit int, closes func(string) bool) (int, string, bool) {
	for off := from; off < limit; {
		text, next := p.line(off)
		if closes(text) {
			return off, text, true
		}
		off = next
	}
	return 0, "", false
}

// An affiliated keyword line belongs to the element below it. The
// subset uses one, ~#+name:~, and the two offsets it produces are
// both load bearing: the element's extent opens at the name line,
// and the tangler's anchor stays on the block's own first line.
type affiliated struct {
	name  string // the #+name: value, or ""
	begin int    // where the element's extent starts
}

// A headline owns its subtree, which ends at the next headline of
// its level or shallower. The property drawer directly below it is
// the headline's own data, so the parser asks for it in property
// drawer mode before handing the rest of the range to the loop.
// Everything after the stars is optional except the title:
//   ** DONE [#A] COMMENT Ship it            :release:tools:
func (p *parser) headlineParser(text string, next, limit int) (parsed, int, error) {
	off := p.off
	stars := starRun(text)
	h := &book.Headline{Level: stars, Line: p.lineNum(off),
		AfterStars: off + stars + 1,
		Head:       book.Span{Start: off, End: off + len(text)}}
	rest := strings.TrimLeft(text[stars+1:], " \t")
	if w, r, ok := firstWord(rest); ok && p.todo[w] {
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
	end := p.nextHeadline(next, limit, stars)
	contents := next
	p.off = next
	if el, after, err := p.currentElement(end, modePropertyDrawer); err == nil && el.props != nil {
		h.Properties = el.props
		contents = after
	}
	return parsed{node: h, contents: book.Span{Start: contents, End: end},
		adopt:     func(kids []book.Node) { h.Children = kids },
		childMode: modeSection}, end, nil
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
// the title, so the first line below has two tags and the second has
// none:
//   ** Ship it                              :release:tools:
//   ** Ship it: a story in two parts
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

// The property drawer sits directly under its headline. A malformed
// drawer is not a drawer at all, and its lines fall to prose, which
// is also what org does. The shape:
//   :PROPERTIES:
//   :header-args: :comments org
//   :END:
func (p *parser) propertyDrawerParser(from, limit int) (map[string]string, int, bool) {
	text, next := p.line(from)
	key, value, ok := colonLine(text)
	if !ok || !strings.EqualFold(key, "PROPERTIES") || value != "" {
		return nil, from, false
	}
	props := map[string]string{}
	for off := next; off < limit; {
		text, next := p.line(off)
		key, value, ok := colonLine(text)
		if !ok {
			return nil, from, false
		}
		if strings.EqualFold(key, "END") && value == "" {
			return props, next, true
		}
		props[strings.ToLower(key)] = value
		off = next
	}
	return nil, from, false
}

func (p *parser) srcBlockParser(rest string, bodyStart, limit int, aff affiliated) (parsed, int, error) {
	beginAt := p.off
	endOff, endText, ok := p.findEnd(bodyStart, limit, closesBlock("src"))
	if !ok {
		return parsed{}, 0, fmt.Errorf("%s:%d: unterminated src block", p.path, p.lineNum(beginAt))
	}
	_, _, after, _ := blockDelim(endText, "end")
	lang, switches, header := srcHeader(rest)
	_, next := p.line(endOff)
	return parsed{node: &book.SrcBlock{Lang: lang, Name: aff.name,
		Switches: switches, RawHeader: header,
		Body:     book.Span{Start: bodyStart, End: endOff},
		BeginAt:  beginAt, AfterEnd: endOff + after,
		Line:     p.lineNum(beginAt),
		Full:     book.Span{Start: aff.begin, End: endOff + len(endText)}}}, next, nil
}

func closesBlock(name string) func(string) bool {
	return func(text string) bool {
		got, tail, _, ok := blockDelim(text, "end")
		return ok && got == name && tail == ""
	}
}

// Verbatim blocks hold text, so they report no contents range and
// the loop cannot descend into them, which is what keeps a
// comma-escaped delimiter inside one from being read as a block. The
// kinds are example, export, comment and verse:
//   #+begin_export latex
//   \emph{raw text, passed through}
//   #+end_export
func (p *parser) verbatimBlockParser(kind string, bodyStart, limit int, aff affiliated) (parsed, int, error) {
	beginAt := p.off
	endOff, endText, ok := p.findEnd(bodyStart, limit, closesBlock(kind))
	if !ok {
		return parsed{}, 0, fmt.Errorf("%s:%d: unterminated %s block", p.path, p.lineNum(beginAt), kind)
	}
	_, next := p.line(endOff)
	return parsed{node: &book.VerbatimBlock{Kind: kind,
		Body: book.Span{Start: bodyStart, End: endOff},
		Line: p.lineNum(beginAt),
		Full: book.Span{Start: aff.begin, End: endOff + len(endText)}}}, next, nil
}

// Greater blocks hold elements, so they hand the loop their interior
// and adopt what comes back. A quote block is the common one, and
// any name org does not reserve lands here too:
//   #+begin_quote
//   Prose, and blocks, parse normally in here.
//   #+end_quote
func (p *parser) greaterBlockParser(kind string, bodyStart, limit int, aff affiliated) (parsed, int, error) {
	beginAt := p.off
	endOff, endText, ok := p.findEnd(bodyStart, limit, closesBlock(kind))
	if !ok {
		return parsed{}, 0, fmt.Errorf("%s:%d: unterminated %s block", p.path, p.lineNum(beginAt), kind)
	}
	q := &book.QuoteBlock{Kind: kind, Line: p.lineNum(beginAt),
		Full: book.Span{Start: aff.begin, End: endOff + len(endText)}}
	_, next := p.line(endOff)
	return parsed{node: q, contents: book.Span{Start: bodyStart, End: endOff},
		adopt:     func(kids []book.Node) { q.Children = kids },
		childMode: modeSection}, next, nil
}

// #+end:
func (p *parser) dynamicBlockParser(args string, bodyStart, limit int, aff affiliated) (parsed, int, error) {
	beginAt := p.off
	endOff, endText, ok := p.findEnd(bodyStart, limit, func(text string) bool {
		key, value, ok := keywordLine(text)
		return ok && key == kwEnd && value == ""
	})
	if !ok {
		return parsed{}, 0, fmt.Errorf("%s:%d: unterminated dynamic block", p.path, p.lineNum(beginAt))
	}
	name, rest, _ := strings.Cut(args, " ")
	db := &book.DynamicBlock{Name: name, Args: strings.TrimSpace(rest),
		Interior: book.Span{Start: bodyStart, End: endOff},
		Line:     p.lineNum(beginAt),
		Full:     book.Span{Start: aff.begin, End: endOff + len(endText)}}
	_, next := p.line(endOff)
	return parsed{node: db, contents: book.Span{Start: bodyStart, End: endOff},
		adopt:     func(kids []book.Node) { db.Children = kids },
		childMode: modeSection}, next, nil
}

// A keyword line is one node and one map entry. Prose is everything
// left over: it runs from its first line to whatever interrupts it,
// blank lines and stray colon lines included.
//   #+title: babble
//   #+property: header-args :comments org
func (p *parser) keywordParser(key, value, text string, next int) (parsed, int, error) {
	k := &book.Keyword{Key: key, Value: value, Line: p.lineNum(p.off),
		Raw: book.Span{Start: p.off, End: p.off + len(text)}}
	p.d.Keywords[key] = append(p.d.Keywords[key], value)
	return parsed{node: k}, next, nil
}

func (p *parser) paragraphParser(limit int) (parsed, int, error) {
	start := p.off
	off := start
	for off < limit {
		text, next := p.line(off)
		if off != start && startsElement(text) {
			break
		}
		off = next
	}
	return parsed{node: &book.Prose{Line: p.lineNum(start),
		Text: book.Span{Start: start, End: off}}}, off, nil
}

// #+title: babble
//   :header-args: :comments org
func starRun(text string) int {
	i := 0
	for i < len(text) && text[i] == '*' {
		i++
	}
	if i > 0 && i < len(text) && text[i] == ' ' {
		return i
	}
	return 0
}

func blockDelim(text, word string) (name, rest string, after int, ok bool) {
	i := 0
	for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
		i++
	}
	prefix := "#+" + word + "_"
	if len(text) < i+len(prefix) || !strings.EqualFold(text[i:i+len(prefix)], prefix) {
		return "", "", 0, false
	}
	j := i + len(prefix)
	k := j
	for k < len(text) && text[k] != ' ' && text[k] != '\t' {
		k++
	}
	if k == j {
		return "", "", 0, false
	}
	return strings.ToLower(text[j:k]), strings.TrimSpace(text[k:]), k, true
}

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

func colonLine(text string) (key, value string, ok bool) {
	t := strings.TrimSpace(text)
	if len(t) < 3 || t[0] != ':' {
		return "", "", false
	}
	j := strings.IndexByte(t[1:], ':')
	if j <= 0 || strings.ContainsAny(t[1:1+j], " \t") {
		return "", "", false
	}
	return t[1 : 1+j], strings.TrimSpace(t[2+j:]), true
}

// The todo words come from the book's own header before any headline
// is read, because org applies a ~#+todo~ line to the whole file
// wherever it sits. Each word drops its shortcut and logging suffix,
// and the bar between active and done states drops too, since
// recognizing the words is all a tangler needs.
const (
	kwTodo    = "todo"
	kwSeqTodo = "seq_todo"
	kwTypTodo = "typ_todo"
	kwName    = "name"
	kwBegin   = "begin" // the dynamic-block opener, #+begin:
	kwEnd     = "end"   // and its closer
)

func todoWords(src []byte) []string {
	var words []string
	for off := 0; off < len(src); {
		text, next := line(src, off)
		if key, value, ok := keywordLine(text); ok {
			switch key {
			case kwTodo, kwSeqTodo, kwTypTodo:
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

// The begin_src line splits three ways: the language, then a run of
// switches, then the header string. Only the switch run needs
// grammar, since org accepts exactly -i, -k, -r, -n and +n with an
// optional count, and -l with a quoted format. Whatever follows the
// run is the header string, which stays raw until it is resolved. In
// the tail below the language is go, the switch run is "-i", and the
// header begins at the first colon:
//   go -i :tangle greet.go :comments org
func srcHeader(s string) (lang, switches, header string) {
	rest := strings.TrimLeft(s, " \t")
	lang, rest = cutWord(rest)
	run := rest
	taken := 0
	for {
		w, after := cutWord(rest)
		switch w {
		case "":
			return lang, strings.TrimSpace(run[:taken]), ""
		case "-i", "-k", "-r":
		case "-n", "+n":
			if n, a := cutWord(after); isNumber(n) {
				after = a
			}
		case "-l":
			if q, a, ok := cutQuoted(after); ok {
				_ = q
				after = a
			} else {
				return lang, strings.TrimSpace(run[:taken]), strings.TrimSpace(rest)
			}
		default:
			return lang, strings.TrimSpace(run[:taken]), strings.TrimSpace(rest)
		}
		rest = after
		taken = len(run) - len(rest)
	}
}

func cutWord(s string) (word, rest string) {
	s = strings.TrimLeft(s, " \t")
	if i := strings.IndexAny(s, " \t"); i >= 0 {
		return s[:i], s[i:]
	}
	return s, ""
}

func cutQuoted(s string) (quoted, rest string, ok bool) {
	t := strings.TrimLeft(s, " \t")
	if !strings.HasPrefix(t, `"`) {
		return "", s, false
	}
	if i := strings.IndexByte(t[1:], '"'); i >= 0 {
		return t[1 : 1+i], t[i+2:], true
	}
	return "", s, false
}

func isNumber(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
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
