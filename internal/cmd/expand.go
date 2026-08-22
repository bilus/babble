// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Expand returns the book as one text, with every chapter's headings
// moved to the depth its include sits at, and a map from a line of
// that text back to the file and line it came from.
package cmd

// Origin says which file a merged line came from, so an error can
// name a chapter and a line inside it rather than a position in a
// text that exists nowhere.
type Origin struct {
	Path string
	Line int
}

func Expand(path string) ([]byte, []Origin, error) {
	panic("HOLE(26): merge every chapter into one text, shifting headings, tracking origins")
}

// A chapter's headings move by one offset, the target depth minus the
// shallowest depth in the chapter, so the chapter keeps its own shape
// and one already at the right depth is untouched.
func shiftHeadings(text []byte, target int) []byte {
	panic("HOLE(26): move every star line by the target minus the shallowest")
}

// Two spellings are understood and every other is refused, which is
// the deviation this stage adds. The keyword is read the way org reads
// it, without regard to case.
func includeSpelling(line string) (path string, chapter bool, ok bool) {
	panic("HOLE(26): the two spellings, case-insensitively, and nothing else")
}
