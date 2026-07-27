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
	toks := tokenize(src)
	d := &book.Document{Path: path, Source: src,
		Keywords: map[string][]string{}, Todo: todoWords(toks)}
	p := &parser{src: src, d: d, toks: toks, todo: map[string]bool{}}
	for _, w := range d.Todo {
		p.todo[w] = true
	}
	p.parse()
	return d, nil
}

// A token is one line. The shapes: a run of stars and a space, a
// means is not the tokenizer's business; a drawer is a grammatical
// fact the parser establishes from position.
type tokKind int

const (
	tokText tokKind = iota
	tokHeadline
	tokKeyword
	tokColonLine
)

type token struct {
	kind  tokKind
	line  int    // 1-based
	start int    // offset of the line's first byte
	end   int    // offset past its text, newline excluded
	next  int    // offset of the next line
	stars int    // tokHeadline: how many
	key   string // tokKeyword lowercased; tokColonLine as written
	value string // tokKeyword, tokColonLine: trimmed
}

func tokenize(src []byte) []token {
	var toks []token
	off, ln := 0, 1
	for off < len(src) {
		text, next := line(src, off)
		t := token{kind: tokText, line: ln, start: off,
			end: off + len(text), next: next}
		if stars := starRun(text); stars > 0 {
			t.kind, t.stars = tokHeadline, stars
		} else if key, value, ok := keywordLine(text); ok {
			t.kind, t.key, t.value = tokKeyword, key, value
		} else if key, value, ok := colonLine(text); ok {
			t.kind, t.key, t.value = tokColonLine, key, value
		}
		toks = append(toks, t)
		off = next
		ln++
	}
	return toks
}

func line(src []byte, off int) (string, int) {
	eol := bytes.IndexByte(src[off:], '\n')
	if eol < 0 {
		return string(src[off:]), len(src)
	}
	return string(src[off : off+eol]), off + eol + 1
}

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

// The todo words come from the token stream before the parser runs,
// because org applies a #+todo line to the whole file wherever it
// sits. Each word drops its shortcut-and-logging suffix, and the bar
// between active and done states drops too, since recognizing the
// words is all the tangler needs. The three keyword spellings are
// the org set.
const (
	kwTodo    = "todo"
	kwSeqTodo = "seq_todo"
	kwTypTodo = "typ_todo"
)

func todoWords(toks []token) []string {
	var words []string
	for _, t := range toks {
		if t.kind != tokKeyword {
			continue
		}
		switch t.key {
		case kwTodo, kwSeqTodo, kwTypTodo:
		default:
			continue
		}
		for _, f := range strings.Fields(t.value) {
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
	if words == nil {
		words = []string{"TODO", "DONE"}
	}
	return words
}

// The parser owns the grammar: which star lines open which
// subtrees, which colon lines form a drawer, and where a prose run
// ends. It walks the token slice with one cursor; every branch of
// parse advances it.
type parser struct {
	src   []byte
	d     *book.Document
	todo  map[string]bool
	toks  []token
	pos   int
	stack []*book.Headline
}

func (p *parser) parse() {
	for p.pos < len(p.toks) {
		switch t := p.toks[p.pos]; t.kind {
		case tokHeadline:
			p.headline(t)
		case tokKeyword:
			p.keyword(t)
		default:
			p.prose()
		}
	}
}

// A keyword line is one node and one map entry. Prose swallows
// every token that is neither a headline nor a keyword, stray colon
// lines included, blank lines included; its span runs from its first
// line to the start of whatever interrupts it.
func (p *parser) keyword(t token) {
	k := &book.Keyword{Key: t.key, Value: t.value, Line: t.line,
		Raw: book.Span{Start: t.start, End: t.end}}
	p.d.Keywords[t.key] = append(p.d.Keywords[t.key], t.value)
	p.attach(k)
	p.pos++
}

func (p *parser) prose() {
	first := p.toks[p.pos]
	last := first
	for p.pos < len(p.toks) {
		t := p.toks[p.pos]
		if t.kind == tokHeadline || t.kind == tokKeyword {
			break
		}
		last = t
		p.pos++
	}
	p.attach(&book.Prose{Line: first.line,
		Text: book.Span{Start: first.start, End: last.next}})
}

// Attachment is a stack of open headlines: a new headline pops
// everything at its level or deeper and becomes a child of what
// remains, and every other node joins the innermost open headline.
// A fifteen-star badge is just a very deep entry in that stack.
func (p *parser) attach(n book.Node) {
	if len(p.stack) == 0 {
		p.d.Nodes = append(p.d.Nodes, n)
		return
	}
	top := p.stack[len(p.stack)-1]
	top.Children = append(top.Children, n)
}

func (p *parser) attachHeadline(h *book.Headline) {
	for len(p.stack) > 0 && p.stack[len(p.stack)-1].Level >= h.Level {
		p.stack = p.stack[:len(p.stack)-1]
	}
	p.attach(h)
	p.stack = append(p.stack, h)
}

// Headline anatomy, in org's order: stars and one space (the token
// carries the count), an optional todo word from the book's own set,
// an optional [#A] priority, an optional COMMENT marker, the title,
// and trailing tags. The ARCHIVE tag sets Archived on the way
// through, and a well-formed property drawer directly below attaches
// before the parser moves on.
func (p *parser) headline(t token) {
	h := &book.Headline{Level: t.stars, Line: t.line,
		AfterStars: t.start + t.stars + 1,
		Head:       book.Span{Start: t.start, End: t.end}}
	rest := strings.TrimLeft(string(p.src[t.start+t.stars+1:t.end]), " \t")
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
	p.attachHeadline(h)
	p.pos++
	if props, ok := p.drawer(); ok {
		h.Properties = props
	}
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

// The keyword shape is #+key: value, the key one colon-free word;
// the begin line of a block has a space before its first colon, so
// it tokenizes as text until stage 3 teaches the tokenizer the
// delimiter shapes.
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

// The property drawer is grammar, not shape: an unbroken run of
// colon-line tokens opening with :PROPERTIES: and closing with
// :END:, directly under a heading. Anything else leaves the cursor
// untouched, and those tokens fall to prose, which is also what org
// does. No line arithmetic survives here; the tokens carry it.
func (p *parser) drawer() (map[string]string, bool) {
	i := p.pos
	if i >= len(p.toks) {
		return nil, false
	}
	t := p.toks[i]
	if t.kind != tokColonLine || !strings.EqualFold(t.key, "PROPERTIES") || t.value != "" {
		return nil, false
	}
	props := map[string]string{}
	for i++; i < len(p.toks); i++ {
		t := p.toks[i]
		if t.kind != tokColonLine {
			return nil, false
		}
		if strings.EqualFold(t.key, "END") && t.value == "" {
			p.pos = i + 1
			return props, true
		}
		props[strings.ToLower(t.key)] = t.value
	}
	return nil, false
}

// parseBlocks is stage 3's seam: the tokenizer gains the block
// delimiter shapes, and this parser method assembles src, dynamic,
// quote, and verbatim blocks from them, with names, spans, and the
// two extent anchors.
func (p *parser) parseBlocks() error {
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
