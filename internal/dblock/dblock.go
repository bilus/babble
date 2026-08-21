// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Package dblock owns the book side: the two dynamic-block writers and
// the splice that puts what they produce back into the book's bytes.
package dblock

import "github.com/bilus/babble/internal/book"

// NamedBody hands back the body of a block by name, shaped the way
// org-babel hands it over. dblock takes it as an argument because
// the shaping is the tangler's rule and a second copy of it could
// drift away from the file the diff describes.
type NamedBody func(name string) (string, error)

// Refresh splices a fresh interior into every dynamic block and
// returns the book's new bytes together with whether they differ from
// the old ones, so the caller writes the file only when there is
// something to write.
func Refresh(d *book.Document, body NamedBody) ([]byte, bool, error) {
	panic("HOLE(9): splice fresh interiors into the retained bytes, add a final newline, report change")
}

// write dispatches an interior to the writer its opener names, or
// refuses a name no writer answers to.
func write(name, args string, staged []byte, self int, file string, body NamedBody) (string, error) {
	panic("HOLE(9): dispatch on the writer's name, or refuse a name with no writer")
}

// planToc renders the jump list: one line-number link per plan-markup
// site in the staged book.
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

// srcBlockDiff renders a unified diff of two named blocks' bodies, or
// the words "no differences".
func srcBlockDiff(args string, body NamedBody) (string, error) {
	panic("HOLE(9): system diff -u over the two named bodies, headers dropped, a closer in the output refused")
}

// dblockArgs reads the opener's arguments as a keyword-to-name map.
func dblockArgs(args string) (map[string]string, error) {
	panic("HOLE(9): keyword and name tokens only, and an error on anything else")
}
