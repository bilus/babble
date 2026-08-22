// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Expand returns the book as one text, with every chapter's headings
// moved to the depth its include sits at, and a map from a line of
// that text back to the file and line it came from.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bilus/babble/internal/org"
)

// Origin says which file a merged line came from, so an error can
// name a chapter and a line inside it rather than a position in a
// text that exists nowhere.
type Origin struct {
	Path string
	Line int
}

func Expand(path string) ([]byte, []Origin, error) {
	lines, origins, err := expandFile(path, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	return []byte(strings.Join(lines, "\n")), origins, nil
}

// A commented subtree ends at the next heading no deeper than the one
// that opened it, and inside one the include line stays as it is. Org
// never opens the file, so neither does this, which is what keeps a
// missing chapter under a comment as quiet here as it is there.
func expandFile(path string, open []string, todo map[string]bool) ([]string, []Origin, error) {
	for _, seen := range open {
		if seen == path {
			return nil, nil, fmt.Errorf("%s: includes itself, through %s", path, strings.Join(open, " -> "))
		}
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	if todo == nil {
		todo = map[string]bool{}
		for _, w := range org.TodoWords(src) {
			todo[w] = true
		}
	}
	var out []string
	var origins []Origin
	level, commented := 0, 0
	for n, line := range strings.Split(strings.TrimSuffix(string(src), "\n"), "\n") {
		here := Origin{Path: path, Line: n + 1}
		if stars := starRun(line); stars > 0 {
			level = stars
			if commented > 0 && stars <= commented {
				commented = 0
			}
			if commented == 0 && commentedHeading(line, todo) {
				commented = stars
			}
		}
		chapter, isChapter, isInclude := includeSpelling(line)
		if !isInclude || commented > 0 {
			out = append(out, line)
			origins = append(origins, here)
			continue
		}
		if chapter == "" {
			return nil, nil, fmt.Errorf("%s:%d: only #+include: \"path\" and #+include: \"path\" example are understood", path, n+1)
		}
		if !isChapter {
			out = append(out, line)
			origins = append(origins, here)
			continue
		}
		sub, subOrigins, err := expandFile(filepath.Join(filepath.Dir(path), chapter), append(open, path), todo)
		if err != nil {
			return nil, nil, err
		}
		shiftHeadings(sub, level+1)
		out = append(out, sub...)
		origins = append(origins, subOrigins...)
	}
	return out, origins, nil
}

// A chapter's headings move by one offset, the target depth minus the
// shallowest depth in the chapter, so the chapter keeps its own shape
// and one already at the right depth is untouched.
func shiftHeadings(lines []string, target int) {
	lowest := 0
	for _, line := range lines {
		if n := starRun(line); n > 0 && (lowest == 0 || n < lowest) {
			lowest = n
		}
	}
	if lowest == 0 || target == lowest {
		return
	}
	offset := target - lowest
	for i, line := range lines {
		if n := starRun(line); n > 0 {
			lines[i] = strings.Repeat("*", n+offset) + line[n:]
		}
	}
}

// A heading is what org's outline search calls one: stars at column
// zero and then a space. Nothing here parses, so a line inside a
// block that looks like a heading is one, which is what org does too.
func starRun(line string) int {
	n := 0
	for n < len(line) && line[n] == '*' {
		n++
	}
	if n == 0 || n == len(line) || line[n] != ' ' {
		return 0
	}
	return n
}

// Org reads a heading as commented when the word COMMENT opens its
// title, after a todo keyword and a priority if either is there. The
// expander cannot ask the parser, since it runs before there is one,
// so it repeats the test on the line.
func commentedHeading(line string, todo map[string]bool) bool {
	n := starRun(line)
	if n == 0 {
		return false
	}
	rest := strings.TrimLeft(line[n:], " \t")
	if w, r, ok := firstWord(rest); ok && todo[w] {
		rest = r
	}
	if len(rest) >= 4 && strings.HasPrefix(rest, "[#") && rest[3] == ']' {
		rest = strings.TrimLeft(rest[4:], " \t")
	}
	return rest == "COMMENT" || strings.HasPrefix(rest, "COMMENT ")
}

func firstWord(s string) (word, rest string, ok bool) {
	i := strings.IndexAny(s, " \t")
	if i < 0 {
		if s == "" {
			return "", s, false
		}
		return s, "", true
	}
	return s[:i], strings.TrimLeft(s[i:], " \t"), true
}

// Two spellings are understood and every other is refused, which is
// the deviation this stage adds. The keyword is read the way org reads
// it, without regard to case.
var includeLine = regexp.MustCompile(`(?i)^[ \t]*#\+include:(.*)$`)

func includeSpelling(line string) (path string, chapter, ok bool) {
	m := includeLine.FindStringSubmatch(line)
	if m == nil {
		return "", false, false
	}
	rest := strings.TrimSpace(m[1])
	if !strings.HasPrefix(rest, `"`) {
		return "", false, true
	}
	end := strings.IndexByte(rest[1:], '"')
	if end < 0 {
		return "", false, true
	}
	path, after := rest[1:1+end], strings.TrimSpace(rest[end+2:])
	switch after {
	case "":
		return path, true, true
	case "example":
		return path, false, true
	}
	return "", false, true
}
