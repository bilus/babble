// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Document is one parsed book: the node tree plus the retained
// source bytes that every Span indexes.
package book

import "bytes"

type Document struct {
	Path     string              // where Source came from
	Source   []byte              // retained input bytes
	Keywords map[string][]string // file keywords, lowercased keys
	Todo     []string            // headline state words, default TODO and DONE
	Nodes    []Node              // top level, document order
}

// Span is a half-open byte range [Start, End) into Source.
type Span struct{ Start, End int }

// Node is any tree element; the concrete types below are the whole
// set, and the marker method keeps the set closed.
type Node interface{ isNode() }

func (*Headline) isNode()      {}
func (*Prose) isNode()         {}
func (*Keyword) isNode()       {}
func (*SrcBlock) isNode()      {}
func (*DynamicBlock) isNode()  {}
func (*QuoteBlock) isNode()    {}
func (*VerbatimBlock) isNode() {}

// A Headline carries everything collection and comments need:
// COMMENT and ARCHIVE for skipping, the property drawer for header
// inheritance, and AfterStars, the byte anchor the comment extents
// hang from. Badge lines land here too, as deep ordinary headlines.
type Headline struct {
	Level      int
	Todo       string            // one of Document.Todo, or ""
	Commented  bool              // COMMENT marker
	Archived   bool              // ARCHIVE tag
	Tags       []string
	Title      string
	Properties map[string]string // property drawer, nil when absent
	AfterStars int               // offset just past the stars and one space
	Line       int
	Head       Span              // the heading line
	Children   []Node
}

type Prose struct {
	Line int
	Text Span
}

type Keyword struct {
	Key   string // lowercased, without the #+ and the colon
	Value string
	Line  int
	Raw   Span
}

// SrcBlock records the two raw byte anchors next to the parsed
// attributes: BeginAt, where its begin line starts, and AfterEnd,
// just past the literal end keyword. Params stays zero until header
// resolution fills it.
type SrcBlock struct {
	Lang      string
	Name      string // affiliated name line, or ""
	Switches  string
	RawHeader string // the block's own header line, unresolved
	Params    Params // filled by resolution, zero until then
	Body      Span   // raw bytes between the delimiter lines
	BeginAt   int    // offset where the begin line starts
	AfterEnd  int    // offset just past the literal end keyword
	Line      int    // line of the begin line
	Full      Span   // name line through end line
}

// A dynamic block keeps its interior as both bytes and children:
// refresh splices the span, and the src block inside a generated
// diff is a real block the collection walk must see. Quote blocks
// hold children too; example, export, and comment blocks are
// verbatim.
type DynamicBlock struct {
	Name     string
	Args     string
	Interior Span // bytes between the begin and end lines
	Line     int
	Full     Span
	Children []Node
}

type QuoteBlock struct {
	Line     int
	Full     Span
	Children []Node
}

type VerbatimBlock struct {
	Kind string // example, export, or comment
	Body Span
	Line int
	Full Span
}

// Params is the resolved header set, the engine's side of the closed
// table. The frontend produces it or errors; the engine never sees a
// raw header string.
type Params struct {
	Tangle         string // "no", "yes", or a target path
	Mkdirp         bool
	Comments       string // "no" or "org"
	Padline        bool
	Noweb          bool
	NowebRef       string
	NowebSep       string // "" means one newline
	PreserveIndent bool   // the -i switch
}

// Walk visits every node in document order, descending into
// children; fn returning false prunes the subtree, which is how
// COMMENT skipping stays one line at the call site.
func (d *Document) Walk(fn func(Node) bool) {
	panic("HOLE(2): document-order walk, fn returning false prunes the subtree")
}

func (d *Document) LineAt(off int) int {
	return 1 + bytes.Count(d.Source[:off], []byte("\n"))
}

// Dump is parse --dump's output: stable JSON, so the fixtures can
// pin it byte for byte.
func Dump(d *Document) ([]byte, error) {
	panic("HOLE(2): stable JSON rendering of the tree")
}
