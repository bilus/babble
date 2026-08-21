// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Refresh takes a function rather than reaching for the bodies itself.
// A src-block-diff needs the body of a named block shaped the way
// org-babel shapes it, and those rules belong to the tangler. Passing
// them in keeps dblock resting on the document alone and keeps one
// copy of the shaping.
package dblock

import "github.com/bilus/babble/internal/book"

// NamedBody hands back the body of a block by name, shaped the way
// org-babel hands it over. dblock takes it as an argument because
// the shaping is the tangler's rule and a second copy of it could
// drift away from the file the diff describes.
type NamedBody func(name string) (string, error)

func Refresh(d *book.Document, body NamedBody) ([]byte, bool, error) {
	panic("HOLE(9): splice fresh interiors into the retained bytes, report change")
}

// The second is the buffer each writer reads. Org refreshes from the
// top down, so a writer sees a book where the blocks above it are
// already rewritten, its own interior is a single empty line, and the
// blocks below it still hold their old contents. plan-toc counts lines
// in that buffer. Handing it the original book, or the finished one,
// gives different numbers for the same input. So the loop stages that
// exact byte sequence for each writer: the output built so far, one
// newline, and the rest of the original from the closing line onward.
func write(name, args string, staged []byte, self int, file string, body NamedBody) (string, error) {
	panic("HOLE(9): dispatch on the writer's name, or refuse a name with no writer")
}

// Labels are spelled without the sigils on purpose. A label that
// carried them would be a plan site itself, and the index would start
// indexing itself on the next refresh.
func planToc(staged []byte, self int, file string) string {
	panic("HOLE(9): a line link per plan site, sites below self shifted by the insert count")
}

// planSite answers what a line is called in the index, or nothing
// when the line is not a site. A line has to match the census
// pattern first: the naming rules below are wider than it is, and
// the driver applies them only to lines the census already found.
func planSite(text string) string {
	panic("HOLE(9): the census pattern, then the driver's naming rules in the driver's order")
}

// planStrip is the driver's label cleanup: trim, replace org links
// with their descriptions, drop a trailing tag block, and cut to
// sixty characters with an ellipsis. The count is characters and
// not bytes, which is what the driver counts.
func planStrip(s string) string {
	panic("HOLE(9): trim, unlink, untag, and truncate the way the driver does")
}

// Arguments are read strictly. Org hands the argument text to the lisp
// reader, which accepts anything lisp accepts; babble takes keyword
// and name tokens and refuses the rest, so the subset does not inherit
// a reader it has no use for.
func srcBlockDiff(args string, body NamedBody) (string, error) {
	panic("HOLE(9): system diff -u over the two named bodies, temp-file headers dropped")
}

// dblockArgs reads the opener's arguments as a keyword-to-name map.
func dblockArgs(args string) (map[string]string, error) {
	panic("HOLE(9): keyword and name tokens only, and an error on anything else")
}
