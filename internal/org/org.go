// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Parse reads a file; ParseBytes does the work, so tests and the
// refresh reparse can feed bytes directly.
package org

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
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
	if bytes.Contains(src, []byte("\r\n")) {
		return nil, fmt.Errorf("%s: CRLF line endings; org rewrites generated lines in the file's own ending and babble splices bytes, so the two cannot agree on a mixed file", path)
	}
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
	if err := checkDynamicBlocks(d); err != nil {
		return nil, err
	}
	return d, nil
}

// checkDynamicBlocks refuses a book where org's raw search for a
// dynamic-block opener and the tree disagree about where the dynamic
// blocks are.
var dblockStart = regexp.MustCompile(`(?i)^[ \t]*#\+begin:[ \t]+\S`)

func checkDynamicBlocks(d *book.Document) error {
	tree := map[int]bool{}
	d.Walk(func(n book.Node) bool {
		if db, ok := n.(*book.DynamicBlock); ok {
			tree[db.Line] = true
		}
		return true
	})
	num := 0
	for off := 0; off < len(d.Source); {
		text, next := line(d.Source, off)
		num++
		switch found := dblockStart.MatchString(text); {
		case found && !tree[num]:
			return fmt.Errorf("%s:%d: %s: org reads this line as a dynamic block and babble does not, because org's search ignores whatever block the line sits in; escape it with a comma", d.Path, num, strings.TrimSpace(text))
		case !found && tree[num]:
			return fmt.Errorf("%s:%d: %s: babble reads this line as a dynamic block and org does not, because org's search wants whitespace after the colon; write the space", d.Path, num, strings.TrimSpace(text))
		}
		off = next
	}
	return nil
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
// counterpart here, and not because the line is unrecognized: it is,
// and the drawer below it is read. The difference is where. Org
// threads a mode to make the next line's parse depend on the last
// one; babble handles the planning line inside the headline parser,
// which steps over it and then asks for the drawer. Same result,
// one fewer mode.
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

// The order differs from org's table in [[#the-dispatch][The dispatch]], which tests
// node-property before headline. Org can afford that because it has
// modes babble does not: item, table-row and node-property each
// narrow what a line is allowed to be. babble carries one such mode,
// for the property drawer, and a star run is unambiguous against it
// because a drawer cannot contain a headline. So the two orders agree
// on every input either can be given, and the difference is in what
// org must still rule out at that point and babble already has.
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
	if blockAt, name, header, ok := p.affiliatedRun(start, limit); ok {
		aff.name, aff.header = name, header
		p.off = blockAt
		text, next = p.line(blockAt)
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
// subset steps over org's whole affiliated set and acts on three of
// them, ~#+name:~, ~#+header:~ and ~#+headers:~, and the offsets a run
// of them produces are load bearing: the element's span opens at
// the first line of the run, and the tangler's anchor stays on the
// block's own first line.
type affiliated struct {
	name   string // the #+name: value, or ""
	header string // the #+header: values, in document order
	begin  int    // where the element's span starts
}

// affiliatedKey answers whether a keyword belongs to the block below
// it, and gives back the key org would file it under.
var affiliatedKeys = map[string]bool{
	"caption": true, "data": true, "header": true, "headers": true,
	"label": true, "name": true, "plot": true, "resname": true,
	"result": true, "results": true, "source": true, "srcname": true,
	"tblname": true,
}

func affiliatedKey(key string) (string, bool) {
	if j := strings.IndexByte(key, '['); j >= 0 {
		if !strings.HasSuffix(key, "]") {
			return "", false
		}
		base := key[:j]
		return base, base == "caption" || base == "results"
	}
	if strings.HasPrefix(key, "attr_") {
		return key, len(key) > len("attr_")
	}
	return key, affiliatedKeys[key]
}

// affiliatedRun gathers the affiliated lines above a block and reports
// where the block starts, or nothing when the run reaches no block.
func (p *parser) affiliatedRun(from, limit int) (blockAt int, name, header string, ok bool) {
	for off := from; off < limit; {
		text, next := p.line(off)
		if opensBlock(text) {
			return off, name, header, off != from
		}
		key, value, isKeyword := keywordLine(text)
		if !isKeyword {
			return 0, "", "", false
		}
		key, affiliated := affiliatedKey(key)
		if !affiliated {
			return 0, "", "", false
		}
		switch key {
		case kwName:
			name = value
		case kwHeader, kwHeaders:
			header += " " + value
		}
		off = next
	}
	return 0, "", "", false
}

// A headline owns its subtree, which ends at the next headline of
// its level or shallower. A planning line may stand between the
// heading and its property drawer, and the drawer is the headline's
// own data either way, so the parser steps over the one and asks for
// the other before handing the rest of the range to the loop.
// Everything after the stars is optional except the title.
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
	if after, ok := p.planningLine(next, end); ok {
		contents = after
	}
	p.off = contents
	if el, after, err := p.currentElement(end, modePropertyDrawer); err == nil && el.props != nil {
		h.Properties = el.props
		contents = after
	}
	return parsed{node: h, contents: book.Span{Start: contents, End: end},
		adopt:     func(kids []book.Node) { h.Children = kids },
		childMode: modeSection}, end, nil
}

// A planning line carries the timestamps a heading schedules itself
// with. Nothing in a tangle reads them, but they have to be stepped
// over, because org lets one stand between a heading and the drawer
// whose properties the blocks below inherit.
func (p *parser) planningLine(from, limit int) (int, bool) {
	if from >= limit {
		return from, false
	}
	text, next := p.line(from)
	head := strings.TrimLeft(text, " \t")
	for _, kw := range []string{"SCHEDULED:", "DEADLINE:", "CLOSED:"} {
		if strings.HasPrefix(head, kw) {
			return next, true
		}
	}
	return from, false
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
// none.
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
// is also what org does. The shape.
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

// A src block records the two raw anchors beside its parsed
// attributes: where its own first line starts, and the offset just
// past the literal end keyword, which is where org's backward search
// stops. The terminator has to name the same block and carry nothing
// after it, and the bound keeps the search inside one section.
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
		Switches: switches, RawHeader: header + aff.header,
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
// kinds are example, export, comment and verse.
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
// any name org does not reserve lands here too.
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

// A dynamic block keeps its interior twice over: the span refresh
// rewrites, and the nodes inside it, because a generated diff holds
// a real src block that collection has to see. The opener carries
// the writer's name and its arguments.
// org-dblock-end-re, which is looser than the keyword reader on
// three counts: the colon is optional, anything after it is
// ignored, and the line may be indented. The refresh overwrites
// whatever this predicate calls the interior, so it follows org.
var dblockEnd = regexp.MustCompile(`(?i)^[ \t]*#\+end([: \t\r]|$)`)

func (p *parser) dynamicBlockParser(args string, bodyStart, limit int, aff affiliated) (parsed, int, error) {
	beginAt := p.off
	if text, _ := p.line(beginAt); text != strings.TrimLeft(text, " \t") {
		return parsed{}, 0, fmt.Errorf("%s:%d: indented dynamic block; org indents the refreshed interior to match when the book is in a window and leaves it alone when it is not, so write it at column zero", p.path, p.lineNum(beginAt))
	}
	endOff, endText, ok := p.findEnd(bodyStart, limit, dblockEnd.MatchString)
	if !ok {
		return parsed{}, 0, fmt.Errorf("%s:%d: unterminated dynamic block", p.path, p.lineNum(beginAt))
	}
	// org's opener pattern takes the name as a run of non-whitespace,
	// so a tab between the name and the arguments ends it the way a
	// space does; cutting on a space alone would fold the first
	// argument into the name and lose the writer.
	name, rest := cutWord(args)
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

// The shape tests below are the only place a line's spelling is
// examined, and every one of them is a pure function of the line
// plus a position. They are read at the point of decision, never
// ahead of time, which is what stops a lexical phase from deciding
// grammar. Each answers one question about one line, in the order
// below: how many stars open it, whether it is a block delimiter,
// whether it is a keyword line, and whether it is a colon line.
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
			c >= '0' && c <= '9', c == '_', c == '-',
			c == '[', c == ']':
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
	kwHeader  = "header"
	kwBegin   = "begin" // the dynamic-block opener, #+begin:
	kwEnd     = "end"   // and its closer
	kwHeaders = "headers" // org's other spelling of #+header:
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

// The =begin_src= line splits three ways: the language, then a run of
// switches, then the header string. Only the switch run needs
// grammar, since org accepts exactly -i, -k, -r, -n and +n with an
// optional count, and -l with a quoted format. Whatever follows the
// run is the header string, which stays raw until it is resolved. In
// the tail below the language is go, the switch run is "-i", and the
// header begins at the first colon.
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
	var sc scope
	for _, v := range d.Keywords["property"] {
		name, rest, _ := strings.Cut(v, " ")
		next, err := sc.with(map[string]string{strings.ToLower(name): strings.TrimSpace(rest)})
		if err != nil {
			return fmt.Errorf("%s: %w", d.Path, err)
		}
		sc = next
	}
	return resolveNodes(d, d.Nodes, sc)
}

// Inheritance is nearest-wins, and wholesale. Org reads an inherited
// property with org-entry-get, which returns the closest ancestor's
// value and stops, so a nearer drawer replaces a farther one
// entirely rather than adding to it; a key the nearer one omits is
// not inherited from further up. The scope below therefore carries
// one generic value and one per language, each overwritten as the
// walk descends.
type scope struct {
	generic string            // header-args
	lang    map[string]string // header-args:LANG
}

func (s scope) with(props map[string]string) (scope, error) {
	out := scope{generic: s.generic, lang: s.lang}
	for k, v := range props {
		if strings.HasSuffix(k, "+") {
			return out, fmt.Errorf("accumulating property %s is not in the subset; set the whole value", k)
		}
		if k == "header-args" {
			out.generic = v
			continue
		}
		if l, ok := strings.CutPrefix(k, "header-args:"); ok {
			next := map[string]string{}
			for key, val := range out.lang {
				next[key] = val
			}
			next[l] = v
			out.lang = next
		}
	}
	return out, nil
}

// The walk hands each block the scope in force where it sits, and
// descends into everything that holds elements.
func resolveNodes(d *book.Document, nodes []book.Node, sc scope) error {
	for _, n := range nodes {
		switch n := n.(type) {
		case *book.Headline:
			next, err := sc.with(n.Properties)
			if err != nil {
				return fmt.Errorf("%s:%d: %w", d.Path, n.Line, err)
			}
			if err := resolveNodes(d, n.Children, next); err != nil {
				return err
			}
		case *book.SrcBlock:
			params, err := resolveBlock(n, sc)
			if err != nil {
				return fmt.Errorf("%s:%d: %w", d.Path, n.Line, err)
			}
			n.Params = params
		case *book.DynamicBlock:
			if err := resolveNodes(d, n.Children, sc); err != nil {
				return err
			}
		case *book.QuoteBlock:
			if err := resolveNodes(d, n.Children, sc); err != nil {
				return err
			}
		}
	}
	return nil
}

// One block's parameters are the driver's defaults, then what it
// inherits, then its own header line, each later source overwriting
// the keys it names. The switch run lands last, since a switch is
// part of the block's own line.
func resolveBlock(b *book.SrcBlock, sc scope) (book.Params, error) {
	p := book.Params{Tangle: "no", Comments: "no", Mkdirp: true, Padline: true, NowebSep: "\n"}
	for _, header := range []string{sc.generic, sc.lang[b.Lang], b.RawHeader} {
		if err := applyHeader(&p, header); err != nil {
			return p, err
		}
	}
	if err := applySwitches(&p, b.Switches); err != nil {
		return p, err
	}
	return p, nil
}

// applyHeader is where the table earns its keep: every key is looked
// up, an unknown one stops the tangle, a rejected one names itself,
// and an inert one is dropped on the floor.
func applyHeader(p *book.Params, header string) error {
	for _, arg := range balancedSplit(header) {
		if !strings.HasPrefix(arg, ":") {
			continue
		}
		key, rest, _ := strings.Cut(arg[1:], " ")
		key = strings.ToLower(key)
		bin, known := classify(key)
		switch {
		case !known:
			return fmt.Errorf("unknown header argument :%s", key)
		case bin == binRejected:
			return fmt.Errorf(":%s is not in the subset; see the deviations", key)
		case bin == binInert:
			continue
		}
		value, err := readValue(rest)
		if err != nil {
			return err
		}
		switch key {
		case "tangle":
			p.Tangle = value
		case "mkdirp":
			p.Mkdirp = value != "no"
		case "padline":
			p.Padline = value != "no"
		case "comments":
			if value != "no" && value != "org" {
				return fmt.Errorf(":comments %s is not in the subset; use no or org", value)
			}
			p.Comments = value
		case "noweb":
			for _, word := range strings.Fields(value) {
				switch word {
				case "yes":
					p.Noweb = true
				case "no":
				case "inline":
					if !p.Noweb {
						return fmt.Errorf(":noweb inline only means something beside yes; alone org reads it as no and leaves the references in the file")
					}
				default:
					return fmt.Errorf(":noweb %s is not in the subset; use yes or no, and inline beside yes", word)
				}
			}
		case "noweb-ref":
			p.NowebRef = value
		case "noweb-sep":
			p.NowebSep = value
		}
	}
	return nil
}

// Switches classify the same three ways. Only ~-i~ changes what
// babble does; the line-numbering pair is inert because tangling
// never numbers lines, and the coderef pair is refused because it
// would rewrite the body.
func applySwitches(p *book.Params, switches string) error {
	for _, sw := range strings.Fields(switches) {
		switch sw {
		case "-i":
			p.PreserveIndent = true
		case "-n", "+n", "-k":
		case "-r", "-l":
			return fmt.Errorf("switch %s is not in the subset; coderefs rewrite the body", sw)
		default:
			if _, err := strconv.Atoi(sw); err != nil {
				return fmt.Errorf("unknown switch %s", sw)
			}
		}
	}
	return nil
}

// Every key on every src block lands in one of three bins, and a key
// the table does not know is an error rather than a guess. That is
// the whole fence around org's option space, and it is a lookup, not
// a heuristic.
type headerBin int

const (
	binImplemented headerBin = iota // babble acts on it
	binRejected                     // babble refuses it by name
	binInert                        // org's tangler never reads it
)

func classify(key string) (headerBin, bool) {
	bin, ok := headerBins[key]
	return bin, ok
}

// The table is half the fence, so it is written out rather than derived.
// The inert column is the interesting one: those keys steer
// evaluation, and the tangle path never reads them, which is why
// accepting and ignoring them is safe rather than merely convenient.
var headerBins = map[string]headerBin{
	"tangle": binImplemented, "mkdirp": binImplemented,
	"comments": binImplemented, "padline": binImplemented,
	"noweb": binImplemented, "noweb-ref": binImplemented,
	"noweb-sep": binImplemented,

	"var": binRejected, "shebang": binRejected,
	"tangle-mode": binRejected, "no-expand": binRejected,
	"noweb-prefix": binRejected, "prologue": binRejected,
	"epilogue": binRejected, "dir": binRejected,

	"cache": binInert, "cmdline": binInert, "colnames": binInert,
	"eval": binInert, "exports": binInert, "file": binInert,
	"file-desc": binInert, "file-ext": binInert, "hlines": binInert,
	"noeval": binInert, "output-dir": binInert, "post": binInert,
	"results": binInert, "rownames": binInert, "sep": binInert,
	"session": binInert, "wrap": binInert,
}

func balancedSplit(s string) []string {
	var out []string
	start, depth, quoted := 0, 0, false
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case quoted:
			if c == '\\' {
				i++
			} else if c == '"' {
				quoted = false
			}
		case c == '"':
			quoted = true
		case c == '(' || c == '[':
			depth++
		case c == ')' || c == ']':
			if depth > 0 {
				depth--
			}
		case c == ':' && depth == 0 && i > 0 && (s[i-1] == ' ' || s[i-1] == '\t'):
			out = append(out, strings.TrimSpace(s[start:i]))
			start = i
		}
	}
	if last := strings.TrimSpace(s[start:]); last != "" {
		out = append(out, last)
	}
	if len(out) > 0 && out[0] == "" {
		out = out[1:]
	}
	return out
}

// A value is text. A quoted one is read as elisp reads it, so the
// escapes inside it survive into the value; anything shaped like
// lisp is refused, since org's tangle path evaluates it and babble
// will not.
func readValue(s string) (string, error) {
	v := strings.TrimSpace(s)
	if v == "" {
		return "", nil
	}
	switch v[0] {
	case '(', '\'', '`', '[':
		return "", fmt.Errorf("lisp-valued header argument %s; the subset takes plain values", v)
	case '"':
		return unquote(v)
	}
	return v, nil
}

func unquote(v string) (string, error) {
	if len(v) < 2 || v[len(v)-1] != '"' {
		return "", fmt.Errorf("unterminated string in header argument %s", v)
	}
	var b strings.Builder
	for i := 1; i < len(v)-1; i++ {
		if v[i] != '\\' {
			b.WriteByte(v[i])
			continue
		}
		i++
		if i >= len(v)-1 {
			return "", fmt.Errorf("trailing escape in header argument %s", v)
		}
		switch v[i] {
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		default:
			b.WriteByte(v[i])
		}
	}
	return b.String(), nil
}

// Judging either one needs every block in hand. A block carrying a
// name cannot tell whether another block carries it too, and a name
// cannot tell whether some group elsewhere answers to it, so these
// cannot live in the header table with the rules a single block
// settles.
func Lint(d *book.Document) error {
	panic("HOLE(10): a name two src blocks carry, and a name that is also a group's")
}

// The shape is refused rather than reconciled. Reproducing org's
// search would mean giving up the tree for the one case that abuses
// it, and no book needs to quote a delimiter bare. A comma in front of
// it, the escape org already provides, leaves the meaning and removes
// the ambiguity. Indenting does not: none of the raw searches org runs
// for these are anchored at a column, so an indented delimiter is one
// org still finds.
var quotedDelimiter = regexp.MustCompile(`^#\+(begin|end)_src`)

func LintQuotedDelimiters(d *book.Document) error {
	var err error
	d.Walk(func(n book.Node) bool {
		v, ok := n.(*book.VerbatimBlock)
		if !ok || err != nil {
			return true
		}
		line := v.Line
		for _, text := range strings.Split(string(d.Source[v.Body.Start:v.Body.End]), "\n") {
			line++
			if quotedDelimiter.MatchString(text) {
				err = fmt.Errorf("%s:%d: %s quoted at column zero inside a %s block; org reads it as a real one and babble does not, so indent it or escape it with a comma",
					d.Path, line, strings.TrimSpace(text), v.Kind)
				return false
			}
		}
		return true
	})
	return err
}
