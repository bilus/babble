// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Package tangle turns a tree into files. Run is the whole pass:
// collect the eligible blocks, resolve their targets, assemble each
// file with its banner, and write only what changed.
package tangle

import "github.com/bilus/babble/internal/book"

func Run(d *book.Document) error {
	panic("HOLE(5): collect, resolve targets, assemble with banners, write")
}

// A unit is one target file's worth of blocks, in document order;
// collect builds them in first-appearance order, skipping COMMENT
// and ARCHIVE subtrees, erroring on a languageless block.
type unit struct {
	target string
	blocks []*book.SrcBlock
}

func collect(d *book.Document) ([]unit, error) {
	panic("HOLE(5): eligible blocks grouped by target in first-appearance order")
}

func targetFor(d *book.Document, b *book.SrcBlock) (string, error) {
	panic("HOLE(5): no, yes, or a path relative to the book's directory")
}

// bodyText is the body pipeline in the contract's order; the second
// dedent and the final trim run after noweb expansion, exactly as
// the oracle orders them.
func bodyText(d *book.Document, b *book.SrcBlock) (string, error) {
	panic("HOLE(5): unescape commas, dedent by column, expand noweb, dedent and trim again")
}

// commentFor and wrapComment split the doc-comment story where the
// driver splits it: what text the extent and docstring rule choose,
// then how the language's comment syntax wears it.
func commentFor(d *book.Document, b *book.SrcBlock) (string, error) {
	panic("HOLE(7): extent between the anchors, then the docstring rule")
}

func wrapComment(lang, text string) (string, error) {
	panic("HOLE(7): line-comment table per language, blank lines as a bare marker")
}

// expandNoweb resolves references against the tree: a named block
// wins, else the noweb-ref concatenation, prefixes replicate onto
// every expansion line, recursion allowed, cycles and misses loud.
func expandNoweb(d *book.Document, b *book.SrcBlock, body string) (string, error) {
	panic("HOLE(8): named block first, else noweb-ref concatenation; cycles and misses error")
}

// banner and writeTarget close the pass: the generated-file banner
// in the target's own comment syntax, parent directories on demand,
// a refusal to write the book itself, and no write at all when the
// bytes already match.
func banner(target string) string {
	panic("HOLE(5): the generated-file banner in the target's own comment syntax")
}

func writeTarget(bookPath, target string, content []byte, mkdirp bool) error {
	panic("HOLE(5): refuse the book itself, mkdirp on demand, skip byte-identical writes")
}
