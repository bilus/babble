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
	if err := p.parse(); err != nil {
		return nil, err
	}
	d.Nodes = p.roots
	return d, nil
}

// A token is one line. The shapes: a run of stars and a space, a
// block delimiter, a #+key: line, a colon-fenced line, or plain
// text. What any of them means is not the tokenizer's business; that
// a delimiter pair forms a block, and that a colon-line run forms a
// drawer, are grammatical facts the parser establishes from
// position.
type tokKind int

const (
	tokText tokKind = iota
	tokHeadline
	tokKeyword
	tokColonLine
	tokBlockBegin
	tokBlockEnd
)

type token struct {
	kind  tokKind
	line  int    // 1-based
	start int    // offset of the line's first byte
	end   int    // offset past its text, newline excluded
	next  int    // offset of the next line
	stars int    // tokHeadline: how many
	key   string // tokKeyword, tokBlockBegin, tokBlockEnd: lowercased; tokColonLine as written
	value string // everything after the key, trimmed
	after int    // tokBlockEnd: offset just past the delimiter keyword
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
		} else if name, rest, after, ok := blockDelim(text, "begin"); ok {
			t.kind, t.key, t.value = tokBlockBegin, name, rest
			t.after = off + after
		} else if name, rest, after, ok := blockDelim(text, "end"); ok {
			t.kind, t.key, t.value = tokBlockEnd, name, rest
			t.after = off + after
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
	kwName    = "name"
	kwBegin   = "begin" // the dynamic-block opener, #+begin:
	kwEnd     = "end"   // and its closer
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
	roots []book.Node
}

func (p *parser) parse() error {
	for p.pos < len(p.toks) {
		t := p.toks[p.pos]
		switch {
		case t.kind == tokHeadline:
			p.headline(t)
		case p.atBlock():
			if err := p.block(); err != nil {
				return err
			}
		case t.kind == tokKeyword:
			p.keyword(t)
		default:
			p.prose()
		}
	}
	return nil
}

// A block may wear an affiliated name line, so the cursor is at a
// block when it sits on a delimiter or on a name line with a
// delimiter behind it. That lookahead is the only place the parser
// needs two tokens at once.
func (p *parser) atBlock() bool {
	i := p.pos
	if t := p.toks[i]; t.kind == tokKeyword && t.key == kwName && i+1 < len(p.toks) {
		i++
	}
	t := p.toks[i]
	return t.kind == tokBlockBegin || (t.kind == tokKeyword && t.key == kwBegin)
}

// A keyword line is one node and one map entry. Prose swallows
// every token that opens nothing, stray colon lines included, blank
// lines included; its span runs from its first line to the start of
// whatever interrupts it. A block delimiter interrupts it, or a run
// of prose would eat every block below it.
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
		if t.kind == tokHeadline || t.kind == tokKeyword || t.kind == tokBlockBegin {
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
		p.roots = append(p.roots, n)
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

// block assembles whatever the delimiters enclose, name line
// included. The Full span opens at the name line when there is one,
// so a name sits inside the extent of the block below it, which is
// what the doc-comment rule needs later.
func (p *parser) block() error {
	start := p.toks[p.pos]
	name := ""
	if start.kind == tokKeyword && start.key == kwName {
		name = start.value
		p.pos++
	}
	begin := p.toks[p.pos]
	if begin.kind == tokKeyword {
		return p.dynamicBlock(start, begin, name)
	}
	return p.delimitedBlock(start, begin, name)
}

// The end line must match the opener's name and carry nothing but
// whitespace, and the search stops at the next headline, since org
// looks for the closer inside one section only. A block that never
// closes is an error naming its line; org would quietly reparse the
// opener as a paragraph.
func (p *parser) matchingEnd(closes func(token) bool) (int, bool) {
	for i := p.pos + 1; i < len(p.toks); i++ {
		if t := p.toks[i]; t.kind == tokHeadline {
			return 0, false
		} else if closes(t) {
			return i, true
		}
	}
	return 0, false
}

// A src block carries the two raw anchors: BeginAt where its begin
// line starts, and AfterEnd just past the literal end keyword, which
// is where org's own backward search stops. Everything else in the
// tree is spans and parsed attributes.
func (p *parser) delimitedBlock(start, begin token, name string) error {
	end, ok := p.matchingEnd(func(t token) bool {
		return t.kind == tokBlockEnd && t.key == begin.key && t.value == ""
	})
	if !ok {
		return fmt.Errorf("%s:%d: unterminated %s block", p.d.Path, begin.line, begin.key)
	}
	endTok := p.toks[end]
	full := book.Span{Start: start.start, End: endTok.end}
	body := book.Span{Start: begin.next, End: endTok.start}
	switch begin.key {
	case "src":
		lang, switches, header := srcHeader(begin.value)
		p.attach(&book.SrcBlock{Lang: lang, Name: name,
			Switches: switches, RawHeader: header, Body: body,
			BeginAt: begin.start, AfterEnd: endTok.after,
			Line: begin.line, Full: full})
	case "example", "export", "comment", "verse":
		p.attach(&book.VerbatimBlock{Kind: begin.key, Body: body,
			Line: begin.line, Full: full})
	default:
		kids, err := p.parseRange(p.toks[p.pos+1 : end])
		if err != nil {
			return err
		}
		p.attach(&book.QuoteBlock{Kind: begin.key, Line: begin.line,
			Full: full, Children: kids})
	}
	p.pos = end + 1
	return nil
}

// A dynamic block keeps its interior twice over: the span refresh
// rewrites, and the nodes inside it, because a generated diff holds
// a real src block that collection has to see.
func (p *parser) dynamicBlock(start, begin token, name string) error {
	end, ok := p.matchingEnd(func(t token) bool {
		return t.kind == tokKeyword && t.key == kwEnd && t.value == ""
	})
	if !ok {
		return fmt.Errorf("%s:%d: unterminated dynamic block", p.d.Path, begin.line)
	}
	endTok := p.toks[end]
	blockName, args, _ := strings.Cut(begin.value, " ")
	kids, err := p.parseRange(p.toks[p.pos+1 : end])
	if err != nil {
		return err
	}
	p.attach(&book.DynamicBlock{Name: blockName,
		Args:     strings.TrimSpace(args),
		Interior: book.Span{Start: begin.next, End: endTok.start},
		Line:     begin.line,
		Full:     book.Span{Start: start.start, End: endTok.end},
		Children: kids})
	p.pos = end + 1
	return nil
}

func (p *parser) parseRange(toks []token) ([]book.Node, error) {
	sub := &parser{src: p.src, d: p.d, todo: p.todo, toks: toks}
	if err := sub.parse(); err != nil {
		return nil, err
	}
	return sub.roots, nil
}

// The begin_src line splits three ways: the language, then a run of
// switches, then the header string. Only the switch run needs
// grammar, since org accepts exactly -i, -k, -r, -n and +n with an
// optional count, and -l with a quoted format. Whatever follows the
// run is the header, unresolved until stage 4.
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
