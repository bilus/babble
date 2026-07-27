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

func dumpNodes(nodes []Node) []any {
	out := make([]any, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, dumpNode(n))
	}
	return out
}

func spanStr(s Span) string {
	return fmt.Sprintf("%d:%d", s.Start, s.End)
}

func dumpNode(n Node) any {
	switch n := n.(type) {
	case *Headline:
		return struct {
			Type       string            `json:"type"`
			Level      int               `json:"level"`
			Todo       string            `json:"todo,omitempty"`
			Commented  bool              `json:"commented,omitempty"`
			Archived   bool              `json:"archived,omitempty"`
			Tags       []string          `json:"tags,omitempty"`
			Title      string            `json:"title"`
			Properties map[string]string `json:"properties,omitempty"`
			AfterStars int               `json:"afterStars"`
			Line       int               `json:"line"`
			Head       string            `json:"head"`
			Children   []any             `json:"children,omitempty"`
		}{"headline", n.Level, n.Todo, n.Commented, n.Archived,
			n.Tags, n.Title, n.Properties, n.AfterStars, n.Line,
			spanStr(n.Head), dumpNodes(n.Children)}
	case *Prose:
		return struct {
			Type string `json:"type"`
			Line int    `json:"line"`
			Span string `json:"span"`
		}{"prose", n.Line, spanStr(n.Text)}
	case *Keyword:
		return struct {
			Type  string `json:"type"`
			Key   string `json:"key"`
			Value string `json:"value"`
			Line  int    `json:"line"`
			Span  string `json:"span"`
		}{"keyword", n.Key, n.Value, n.Line, spanStr(n.Raw)}
	case *SrcBlock:
		var params any
		if n.Params != (Params{}) {
			params = n.Params
		}
		return struct {
			Type      string `json:"type"`
			Lang      string `json:"lang"`
			Name      string `json:"name,omitempty"`
			Switches  string `json:"switches,omitempty"`
			RawHeader string `json:"rawHeader,omitempty"`
			Params    any    `json:"params,omitempty"`
			Body      string `json:"body"`
			BeginAt   int    `json:"beginAt"`
			AfterEnd  int    `json:"afterEnd"`
			Line      int    `json:"line"`
			Full      string `json:"full"`
		}{"src", n.Lang, n.Name, n.Switches, n.RawHeader, params,
			spanStr(n.Body), n.BeginAt, n.AfterEnd, n.Line,
			spanStr(n.Full)}
	case *DynamicBlock:
		return struct {
			Type     string `json:"type"`
			Name     string `json:"name"`
			Args     string `json:"args,omitempty"`
			Interior string `json:"interior"`
			Line     int    `json:"line"`
			Full     string `json:"full"`
			Children []any  `json:"children,omitempty"`
		}{"dynamic", n.Name, n.Args, spanStr(n.Interior), n.Line,
			spanStr(n.Full), dumpNodes(n.Children)}
	case *QuoteBlock:
		return struct {
			Type     string `json:"type"`
			Line     int    `json:"line"`
			Full     string `json:"full"`
			Children []any  `json:"children,omitempty"`
		}{"quote", n.Line, spanStr(n.Full), dumpNodes(n.Children)}
	case *VerbatimBlock:
		return struct {
			Type string `json:"type"`
			Kind string `json:"kind"`
			Body string `json:"body"`
			Line int    `json:"line"`
			Full string `json:"full"`
		}{"verbatim", n.Kind, spanStr(n.Body), n.Line, spanStr(n.Full)}
	}
	return nil
}
