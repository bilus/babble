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
// diff is a real block the collection walk must see. Greater blocks
// (quote, center, and any name org does not reserve) hold children
// too; example, export, comment, and verse blocks are verbatim,
// which is org's own split between blocks whose contents are
// elements and blocks whose contents are text.
type DynamicBlock struct {
	Name     string `json:"name"`
	Args     string `json:"args,omitempty"`
	Interior Span   `json:"interior"` // bytes between the begin and end lines
	Line     int    `json:"line"`
	Full     Span   `json:"full"`
	Children []Node `json:"-"`
}

type QuoteBlock struct {
	Kind     string `json:"kind"` // quote, center, or any other greater block
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
	nodes, err := dumpNodes(d.Nodes)
	if err != nil {
		return nil, err
	}
	v := struct {
		Path     string              `json:"path"`
		Todo     []string            `json:"todo"`
		Keywords map[string][]string `json:"keywords,omitempty"`
		Nodes    []json.Marshaler    `json:"nodes"`
	}{d.Path, d.Todo, d.Keywords, nodes}
	out, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

// Each node type carries its own JSON tags, so the dump never
// restates a field list. One wrapper adds what tags cannot: the type
// discriminator in front and the wrapped children behind. A leaf is
// the same wrapper with no children, which is why there is one type
// here and not two.
type dumped struct {
	kind     string
	node     Node
	children []Node // empty for a leaf
}

// The wrapper marshals itself because Go cannot flatten a field into
// its enclosing object: an embedded type parameter is illegal and an
// embedded interface marshals under its own key, so a typed wrapper
// that keeps the flat shape has to splice the node's object itself.
// Marshaling the children here rather than at construction lets an
// unknown node type surface as an error instead of a null.
func (d dumped) MarshalJSON() ([]byte, error) {
	body, err := json.Marshal(d.node)
	if err != nil {
		return nil, err
	}
	if len(body) < 2 || body[0] != '{' || body[len(body)-1] != '}' {
		return nil, fmt.Errorf("dump: %s node marshaled as %s, want an object", d.kind, body)
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, `{"type":%q`, d.kind)
	if inner := body[1 : len(body)-1]; len(inner) > 0 {
		b.WriteByte(',')
		b.Write(inner)
	}
	if len(d.children) > 0 {
		kids, err := dumpNodes(d.children)
		if err != nil {
			return nil, err
		}
		raw, err := json.Marshal(kids)
		if err != nil {
			return nil, err
		}
		b.WriteString(`,"children":`)
		b.Write(raw)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

func dumpNodes(nodes []Node) ([]json.Marshaler, error) {
	out := make([]json.Marshaler, 0, len(nodes))
	for _, n := range nodes {
		m, err := dumpNode(n)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

func dumpNode(n Node) (json.Marshaler, error) {
	switch n := n.(type) {
	case *Headline:
		return dumped{"headline", n, n.Children}, nil
	case *Prose:
		return dumped{"prose", n, nil}, nil
	case *Keyword:
		return dumped{"keyword", n, nil}, nil
	case *SrcBlock:
		return dumped{"src", n, nil}, nil
	case *DynamicBlock:
		return dumped{"dynamic", n, n.Children}, nil
	case *QuoteBlock:
		return dumped{"quote", n, n.Children}, nil
	case *VerbatimBlock:
		return dumped{"verbatim", n, nil}, nil
	}
	return nil, fmt.Errorf("dump: unknown node type %T", n)
}
