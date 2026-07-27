// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Document is one parsed book: the node tree plus the retained
// source bytes that every Span indexes.
package book

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Document struct {
	Path     string              // where Source came from
	Source   []byte              // retained input bytes
	Keywords map[string][]string // file keywords, lowercased keys
	Todo     []string            // headline state words, default TODO and DONE
	Nodes    []Node              // top level, document order
}

// Span is a half-open byte range [Start, End) into Source. It
// marshals as a compact start:end string, which keeps a dumped tree
// readable in a fixture golden.
type Span struct{ Start, End int }

func (s Span) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%d:%d"`, s.Start, s.End)), nil
}

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
	Level      int               `json:"level"`
	Todo       string            `json:"todo,omitempty"` // one of Document.Todo, or ""
	Commented  bool              `json:"commented,omitempty"` // COMMENT marker
	Archived   bool              `json:"archived,omitempty"` // ARCHIVE tag
	Tags       []string          `json:"tags,omitempty"`
	Title      string            `json:"title"`
	Properties map[string]string `json:"properties,omitempty"` // property drawer, nil when absent
	AfterStars int               `json:"afterStars"` // offset just past the stars and one space
	Line       int               `json:"line"`
	Head       Span              `json:"head"` // the heading line
	Children   []Node            `json:"-"`
}

type Prose struct {
	Line int  `json:"line"`
	Text Span `json:"span"`
}

type Keyword struct {
	Key   string `json:"key"` // lowercased, without the #+ and the colon
	Value string `json:"value"`
	Line  int    `json:"line"`
	Raw   Span   `json:"span"`
}

// SrcBlock records the two raw byte anchors next to the parsed
// attributes: BeginAt, where its begin line starts, and AfterEnd,
// just past the literal end keyword. Params stays zero until header
// resolution fills it.
type SrcBlock struct {
	Lang      string `json:"lang"`
	Name      string `json:"name,omitempty"` // affiliated name line, or ""
	Switches  string `json:"switches,omitempty"`
	RawHeader string `json:"rawHeader,omitempty"` // the block's own header line, unresolved
	Params    Params `json:"params,omitzero"` // filled by resolution, zero until then
	Body      Span   `json:"body"` // raw bytes between the delimiter lines
	BeginAt   int    `json:"beginAt"` // offset where the begin line starts
	AfterEnd  int    `json:"afterEnd"` // offset just past the literal end keyword
	Line      int    `json:"line"` // line of the begin line
	Full      Span   `json:"full"` // name line through end line
}

// A dynamic block keeps its interior as both bytes and children:
// refresh splices the span, and the src block inside a generated
// diff is a real block the collection walk must see. Quote blocks
// hold children too; example, export, and comment blocks are
// verbatim.
type DynamicBlock struct {
	Name     string `json:"name"`
	Args     string `json:"args,omitempty"`
	Interior Span   `json:"interior"` // bytes between the begin and end lines
	Line     int    `json:"line"`
	Full     Span   `json:"full"`
	Children []Node `json:"-"`
}

type QuoteBlock struct {
	Line     int    `json:"line"`
	Full     Span   `json:"full"`
	Children []Node `json:"-"`
}

type VerbatimBlock struct {
	Kind string `json:"kind"` // example, export, or comment
	Body Span   `json:"body"`
	Line int    `json:"line"`
	Full Span   `json:"full"`
}

// Params is the resolved header set, the engine's side of the closed
// table. The frontend produces it or errors; the engine never sees a
// raw header string.
type Params struct {
	Tangle         string `json:"tangle"` // "no", "yes", or a target path
	Mkdirp         bool   `json:"mkdirp,omitempty"`
	Comments       string `json:"comments,omitempty"` // "no" or "org"
	Padline        bool   `json:"padline,omitempty"`
	Noweb          bool   `json:"noweb,omitempty"`
	NowebRef       string `json:"nowebRef,omitempty"`
	NowebSep       string `json:"nowebSep,omitempty"` // "" means one newline
	PreserveIndent bool   `json:"preserveIndent,omitempty"` // the -i switch
}

// Walk visits every node in document order, descending into
// children; fn returning false prunes the subtree, which is how
// COMMENT skipping stays one line at the call site.
func (d *Document) Walk(fn func(Node) bool) {
	walk(d.Nodes, fn)
}

func walk(nodes []Node, fn func(Node) bool) {
	for _, n := range nodes {
		if !fn(n) {
			continue
		}
		switch n := n.(type) {
		case *Headline:
			walk(n.Children, fn)
		case *DynamicBlock:
			walk(n.Children, fn)
		case *QuoteBlock:
			walk(n.Children, fn)
		}
	}
}

// LineAt is total: any int is a valid argument, offsets before the
// source count as its start and offsets past it as its end, so no
// caller can turn a bad offset into a panic.
func (d *Document) LineAt(off int) int {
	if off < 0 {
		off = 0
	}
	if off > len(d.Source) {
		off = len(d.Source)
	}
	return 1 + bytes.Count(d.Source[:off], []byte("\n"))
}

// Dump is parse --dump's output: stable JSON, so the fixtures can
// pin it byte for byte. Map keys sort, empty fields drop, and spans
// render as start:end strings, which keeps a golden tree readable in
// a fixture file.
func Dump(d *Document) ([]byte, error) {
	v := struct {
		Path     string              `json:"path"`
		Todo     []string            `json:"todo"`
		Keywords map[string][]string `json:"keywords,omitempty"`
		Nodes    []any               `json:"nodes"`
	}{d.Path, d.Todo, d.Keywords, dumpNodes(d.Nodes)}
	out, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

// Each node type carries its own JSON tags, so the dump never
// restates a field list. The wrappers below add the two things tags
// cannot: the type discriminator, and children rendered through the
// same wrapping (the embedded struct's fields flatten into the
// wrapper's object).
func dumpNodes(nodes []Node) []any {
	out := make([]any, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, dumpNode(n))
	}
	return out
}

func dumpNode(n Node) any {
	switch n := n.(type) {
	case *Headline:
		return struct {
			Type string `json:"type"`
			*Headline
			Children []any `json:"children,omitempty"`
		}{"headline", n, dumpNodes(n.Children)}
	case *Prose:
		return struct {
			Type string `json:"type"`
			*Prose
		}{"prose", n}
	case *Keyword:
		return struct {
			Type string `json:"type"`
			*Keyword
		}{"keyword", n}
	case *SrcBlock:
		return struct {
			Type string `json:"type"`
			*SrcBlock
		}{"src", n}
	case *DynamicBlock:
		return struct {
			Type string `json:"type"`
			*DynamicBlock
			Children []any `json:"children,omitempty"`
		}{"dynamic", n, dumpNodes(n.Children)}
	case *QuoteBlock:
		return struct {
			Type string `json:"type"`
			*QuoteBlock
			Children []any `json:"children,omitempty"`
		}{"quote", n, dumpNodes(n.Children)}
	case *VerbatimBlock:
		return struct {
			Type string `json:"type"`
			*VerbatimBlock
		}{"verbatim", n}
	}
	return nil
}
