// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// Package tangle turns a tree into files. Run is the whole pass:
// collect the eligible blocks, resolve their targets, assemble each
// file with its banner, and write only what changed.
package tangle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bilus/babble/internal/book"
)

func Run(d *book.Document) error {
	comments, err := loadCommentTable(d.Path)
	if err != nil {
		return err
	}
	units, err := collect(d)
	if err != nil {
		return err
	}
	for _, u := range units {
		content, err := assemble(d, u, comments)
		if err != nil {
			return err
		}
		if err := writeTarget(d.Path, u, content); err != nil {
			return err
		}
	}
	return nil
}

// The order is also why this is a function and not a paragraph in two
// command implementations. ~babble tangle~ and the oracle harness both
// need it, and when each carried its own copy they disagreed: the
// harness never ran the lint, so the cross-check was quietly measuring
// something the command does not do.
func Book(path string) error {
	panic("HOLE(9): read, refresh, write back if changed, read again, tangle")
}

// read is the front half, and it is separate because Book performs
// it twice on the same path.
func read(path string) (*book.Document, error) {
	panic("HOLE(9): parse, resolve, and run the lints the fence owns so far")
}

// A book is rewritten the way a target is: beside itself, then
// renamed over. The driver writes in place, which is what Emacs does
// when it saves a buffer, and the bytes are the same either way. The
// difference only shows when a run dies in the middle of the write,
// and the file at stake here is the program itself.
func replaceBook(path string, body []byte) error {
	panic("HOLE(9): write beside the book and rename over it")
}

// The shape is the tangling shape minus one step. Comma escapes come
// off, one trailing newline comes off, and the body is dedented unless
// the block asked to keep its indentation. Noweb references stay as
// they are written, because the driver asks org-babel for the body and
// org-babel expands nothing at that point. So a twin of a block full
// of references shows a diff of the references, which is also the
// diff a reader wants to see.
func NamedBody(d *book.Document, name string) (string, error) {
	panic("HOLE(9): the first block named name, shaped as org-babel hands it over")
}

// It keeps the first error and walks the whole tree anyway. That
// looks like a bug and is the opposite: collect writes nothing, so
// finishing the walk costs nothing, and returning the error at the
// end means Run stops before its write loop starts. Every refusal
// raised here therefore precedes every write. The refusals raised
// later, while a body is being assembled, do not have that property,
// and a failure on the third target leaves the first two on disk.
type unit struct {
	target string
	lang   string
	blocks []*book.SrcBlock
}

func collect(d *book.Document) ([]unit, error) {
	var units []unit
	var firstErr error
	at := map[string]int{}
	fail := func(format string, args ...any) {
		if firstErr == nil {
			firstErr = fmt.Errorf(format, args...)
		}
	}
	d.Walk(func(n book.Node) bool {
		switch n := n.(type) {
		case *book.Headline:
			return !n.Commented && !n.Archived
		case *book.SrcBlock:
			if n.Lang == "" {
				fail("%s:%d: src block without a language", d.Path, n.Line)
				return true
			}
			if n.Params.Tangle == "no" || n.Params.Tangle == "" {
				return true
			}
			target, err := targetFor(d, n)
			if err != nil {
				fail("%s", err)
				return true
			}
			if sameFile(d.Path, target) {
				fail("%s:%d: refusing to tangle into the book itself: %s", d.Path, n.Line, target)
				return true
			}
			i, seen := at[target]
			if !seen {
				i = len(units)
				at[target] = i
				units = append(units, unit{target: target, lang: n.Lang})
			}
			if units[i].lang != n.Lang {
				fail("%s: blocks in %s and %s; one file takes one comment syntax",
					target, units[i].lang, n.Lang)
				return true
			}
			units[i].blocks = append(units[i].blocks, n)
		}
		return true
	})
	return units, firstErr
}

// A target is a path, a derived name, or nothing. The derived name
// is the book's own, with the language's extension, and the map of
// extensions is org's: two entries and the language's own name for
// everything else, which is why a python block asking for a derived
// name lands in a file named for the language.
func targetFor(d *book.Document, b *book.SrcBlock) (string, error) {
	dir := filepath.Dir(d.Path)
	switch target := b.Params.Tangle; target {
	case "yes":
		base := filepath.Base(d.Path)
		base = strings.TrimSuffix(base, filepath.Ext(base))
		return filepath.Join(dir, base+"."+langExt(b.Lang)), nil
	default:
		if filepath.IsAbs(target) {
			return target, nil
		}
		return filepath.Join(dir, target), nil
	}
}

func langExt(lang string) string {
	switch lang {
	case "emacs-lisp", "elisp":
		return "el"
	}
	return lang
}

// bodyText is the body pipeline in the contract's order; the second
// dedent and the final trim run after noweb expansion, exactly as
// the oracle orders them. A block that preserves its indentation
// skips both dedents and keeps its lead through the trim, which is
// the one branch in the pipeline and the reason the trim takes a
// flag.
func bodyText(d *book.Document, b *book.SrcBlock) (string, error) {
	body := unescapeCommas(string(d.Source[b.Body.Start:b.Body.End]))
	body = strings.TrimSuffix(body, "\n")
	keepLead := b.Params.PreserveIndent
	if !keepLead {
		body = dedent(body)
	}
	if b.Params.Noweb {
		expanded, err := expandNoweb(d, b, body)
		if err != nil {
			return "", err
		}
		body = expanded
	}
	if keepLead {
		return orgTrim(body, keepLead), nil
	}
	return orgTrim(dedent(body), keepLead), nil
}

// "\n\n    package main\n\n"  keeping the lead  ->  "    package main"
//   "\n\n    package main\n\n"  otherwise         ->  "package main"
func orgTrim(s string, keepLead bool) string {
	s = strings.TrimRight(s, orgSpace)
	if !keepLead {
		return strings.TrimLeft(s, orgSpace)
	}
	for {
		nl := strings.IndexByte(s, '\n')
		if nl < 0 || strings.Trim(s[:nl], " \t") != "" {
			return s
		}
		s = s[nl+1:]
	}
}

const orgSpace = " \t\n"

// Org's own escape: a line that would otherwise open a block or a
// heading wears a comma, and tangling takes one comma off. The run
// may be longer than one, and only the comma nearest the marker
// belongs to the escape.
func unescapeCommas(body string) string {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		j := 0
		for j < len(line) && (line[j] == ' ' || line[j] == '\t') {
			j++
		}
		k := j
		for k < len(line) && line[k] == ',' {
			k++
		}
		if k == j {
			continue
		}
		if rest := line[k:]; strings.HasPrefix(rest, "*") || strings.HasPrefix(rest, "#+") {
			lines[i] = line[:k-1] + rest
		}
	}
	return strings.Join(lines, "\n")
}

// That is org's behaviour, and it is worth knowing why, because the
// obvious guess is wrong. Org cuts with ~move-to-column n t~, and the
// forcing flag does not split a tab: it pads with spaces in front of
// the tab until point sits at column n. The delete that follows takes
// exactly the padding it just added, so the line comes out as it went
// in. A tab is only removed when the measure lands on its far edge.
func dedent(body string) string {
	lines := strings.Split(body, "\n")
	measure := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if c := indentColumns(line); measure < 0 || c < measure {
			measure = c
		}
	}
	if measure <= 0 {
		return body
	}
	for i, line := range lines {
		lines[i] = cutColumns(line, measure)
	}
	return strings.Join(lines, "\n")
}

func indentColumns(line string) int {
	col := 0
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case ' ':
			col++
		case '\t':
			col += 8 - col%8
		default:
			return col
		}
	}
	return col
}

func cutColumns(line string, n int) string {
	col, i := 0, 0
	for i < len(line) && col < n {
		switch line[i] {
		case ' ':
			col++
		case '\t':
			col += 8 - col%8
		default:
			return line[i:]
		}
		i++
	}
	if col > n {
		return line
	}
	return line[i:]
}

// commentFor and wrapComment split the doc-comment story where the
// driver splits it: what text the extent and docstring rule choose,
// then how the language's comment syntax wears it.
func commentFor(d *book.Document, b *book.SrcBlock, comments commentTable) (string, error) {
	text := docstring(string(d.Source[extentOf(d, b):b.BeginAt]))
	if text == "" {
		return "", nil
	}
	return wrapComment(b.Lang, text, comments)
}

// The extent starts at whichever anchor sits nearest above the block:
// just past a headline's stars, just past a previous block's end
// keyword, or the start of the file. A block that tangles nowhere is
// still an anchor, so every block in the tree counts, not only the
// ones being written.
func extentOf(d *book.Document, b *book.SrcBlock) int {
	at := 0
	d.Walk(func(n book.Node) bool {
		switch n := n.(type) {
		case *book.Headline:
			if n.AfterStars < b.BeginAt && n.AfterStars > at {
				at = n.AfterStars
			}
		case *book.SrcBlock:
			if n.AfterEnd < b.BeginAt && n.AfterEnd > at {
				at = n.AfterEnd
			}
		}
		return true
	})
	return at
}

// Wrapping follows comment-region, which is more particular than a
// prefix on every line. It ignores whitespace at either end of the
// text, indents every marker to the shallowest line it is about to
// comment, and keeps whatever indentation the line had beyond that.
// A blank line inside becomes the marker and a space, which the
// file-wide trim later reduces to the bare marker.
//
// The marker takes one space unless the prefix already ends in
// whitespace. The banner's rule differs: it always adds a space, so
// a project whose prefix ends in one gets two there and one here.
//
// A prefix with nothing but whitespace in it has no marker to write.
// The driver does not refuse that, it dies partway with the file
// unwritten, and matching a crash is worth less than naming it.
func wrapComment(lang, text string, comments commentTable) (string, error) {
	prefix := comments.prefix(lang)
	if strings.TrimSpace(prefix) == "" {
		return "", fmt.Errorf("%s has no comment marker, so a :comments org block cannot be written; give it one in .lit/comments.json", lang)
	}
	marker := prefix
	if !strings.HasSuffix(marker, " ") && !strings.HasSuffix(marker, "\t") {
		marker += " "
	}
	lines := strings.Split(text, "\n")
	lead := shallowest(lines)
	var out strings.Builder
	for _, line := range lines {
		rest := line
		if len(rest) >= len(lead) {
			rest = rest[len(lead):]
		} else {
			rest = strings.TrimLeft(rest, " \t")
		}
		out.WriteString(lead)
		out.WriteString(marker)
		out.WriteString(rest)
		out.WriteString("\n")
	}
	return out.String(), nil
}

// The indentation of the shallowest line that has anything on it.
// A blank line does not lower it.
func shallowest(lines []string) string {
	lead := ""
	found := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		if !found || len(indent) < len(lead) {
			lead, found = indent, true
		}
	}
	return lead
}

// The docstring is everything after the last "# doc" marker line, or
// the last paragraph when there is none. Keyword lines drop first,
// which is why a keyword between two prose lines joins them: the rule
// takes the line and its newline together, and the split that finds
// the last paragraph runs afterwards.
func docstring(text string) string {
	if i := lastDocMarker(text); i >= 0 {
		return sigil(strings.Trim(dropKeywordLines(text[i:]), " \t\n\r"))
	}
	clean := dropKeywordLines(text)
	var last string
	for _, para := range regexp.MustCompile(`\n\n+`).Split(clean, -1) {
		if para != "" {
			last = para
		}
	}
	return sigil(strings.Trim(last, " \t\n\r"))
}

// The marker is the line "# doc" and nothing else, so a trailing
// space disqualifies it, and the last one in the text wins.
func lastDocMarker(text string) int {
	at := -1
	for i := 0; ; {
		j := strings.Index(text[i:], "# doc\n")
		if j < 0 {
			break
		}
		start := i + j
		if start == 0 || text[start-1] == '\n' {
			at = start + len("# doc\n")
		}
		i = start + 1
	}
	return at
}

// Only a keyword at column zero counts, so an indented one reaches
// the comment as ordinary text.
func dropKeywordLines(text string) string {
	return regexp.MustCompile(`(?m)^#\+[^\n]*\n?`).ReplaceAllString(text, "")
}

// The sigil marks an aside for the reader and is not part of the
// comment. It needs the trailing space, and it runs after the trim,
// which is the one way a docstring can start with a newline.
func sigil(text string) string {
	for _, word := range []string{"Note: ", "Apropos: "} {
		if strings.HasPrefix(text, word) {
			return text[len(word):]
		}
	}
	return text
}

// expandNoweb resolves references against the tree: a named block
// wins, else the noweb-ref concatenation, prefixes replicate onto
// every expansion line, recursion allowed, cycles and misses loud.
// The reference pattern is org's, ported character for character.
// Hand-rolling a line scanner instead would miss two things it does:
// a name may hold spaces inside it but not at either end, and a
// one-character name whose closer is followed by another reference on
// the same line swallows both into one unresolvable name.
var nowebRe = regexp.MustCompile(`(.*?)(<<([^ \t\n](?:.*?[^ \t\n])?)>>)`)

func expandNoweb(d *book.Document, b *book.SrcBlock, body string) (string, error) {
	return expandInto(d, body, nil)
}

// The text before a reference is repeated in front of every line the
// reference expands to, except the first, which already sits after
// that text. The prefix is what lies between the previous reference
// and this one on the same line, so a second reference on a line is
// prefixed by the text between the two and not by the whole line. It
// cannot reach back past a line ending, which is why the pattern's
// dot does not match one.
func expandInto(d *book.Document, body string, open []string) (string, error) {
	var out strings.Builder
	rest := body
	for {
		m := nowebRe.FindStringSubmatchIndex(rest)
		if m == nil {
			out.WriteString(rest)
			return out.String(), nil
		}
		prefix := rest[m[2]:m[3]]
		name := rest[m[6]:m[7]]
		out.WriteString(rest[:m[0]])
		out.WriteString(prefix)
		text, err := resolve(d, name, open)
		if err != nil {
			return "", err
		}
		out.WriteString(strings.ReplaceAll(text, "\n", "\n"+prefix))
		rest = rest[m[1]:]
	}
}

// A name resolves to the block that carries it, or to the blocks that
// share it as a :noweb-ref, joined by each one's own separator with
// the last one's dropped. A referenced body is expanded further only
// when that block asks for it: org gates recursion on the block's own
// :noweb, so a plain block's references reach the file as text.
func resolve(d *book.Document, name string, open []string) (string, error) {
	for _, seen := range open {
		if seen == name {
			return "", fmt.Errorf("%s: noweb reference cycle: %s", d.Path, strings.Join(append(open, name), " -> "))
		}
	}
	if strings.ContainsAny(name, "()") {
		return "", fmt.Errorf("%s: <<%s>> is a call reference, which is evaluation and not in the subset", d.Path, name)
	}
	if err := refusedName(d, name); err != nil {
		return "", err
	}
	if named := namedBlock(d, name); named != nil {
		return bodyOf(d, named, append(open, name))
	}
	group := groupBlocks(d, name)
	if len(group) == 0 {
		return "", fmt.Errorf("%s: <<%s>> resolves to nothing%s", d.Path, name, nearMiss(d, name))
	}
	var parts []string
	for i, g := range group {
		text, err := bodyOf(d, g, append(open, name))
		if err != nil {
			return "", err
		}
		parts = append(parts, text)
		if i < len(group)-1 {
			parts = append(parts, g.Params.NowebSep)
		}
	}
	return strings.Join(parts, ""), nil
}

// Two names resolve to something in org that babble will not follow.
// A heading's ~CUSTOM_ID~ is searched before any block, and it answers
// with the whole subtree as raw text, delimiter lines included and
// nothing unescaped, which no book means by a reference. And a name
// that some block also carries as a ~:noweb-ref~ resolves to the block
// or to the group depending on what org expanded earlier in the same
// run, since the group scan replaces the cache it consults. Neither is
// worth reproducing and both are quiet, so both are refused.
func refusedName(d *book.Document, name string) error {
	var err error
	d.Walk(func(n book.Node) bool {
		switch n := n.(type) {
		case *book.Headline:
			if n.Properties["CUSTOM_ID"] == name {
				err = fmt.Errorf("%s: <<%s>> names a heading's CUSTOM_ID, which org expands to the whole subtree as raw text; rename one of them", d.Path, name)
				return false
			}
		case *book.SrcBlock:
			if n.Name == name && n.Params.NowebRef == name {
				err = fmt.Errorf("%s: <<%s>> is both a block name and a :noweb-ref value, and which one org picks depends on what it expanded before; rename one of them", d.Path, name)
				return false
			}
		}
		return true
	})
	return err
}

// The shape below exists because the obvious single walk gets this
// wrong, and the way it gets it wrong is quiet. Carrying one flag
// that turns true at a COMMENT heading and never turns off makes
// every block after that heading look commented, wherever in the
// document it actually sits, so one dead draft near the top
// unresolves references far below it and the error names the
// reference rather than the draft. The flag has no way back down,
// because a walk that reports entering a node never reports leaving
// it. Asking twice removes the need for the flag: each walk is a
// plain search under a fixed rule about COMMENT, each is right on its
// own, and the answer is whether they agree.
func namedBlock(d *book.Document, name string) *book.SrcBlock {
	first := firstNamed(d, name, false)
	if first == nil || first != firstNamed(d, name, true) {
		return nil
	}
	return first
}

// firstNamed is the search both questions need: the first block in
// document order carrying name, over the whole document or over
// only the part org would still tangle.
func firstNamed(d *book.Document, name string, skipCommented bool) *book.SrcBlock {
	var found *book.SrcBlock
	d.Walk(func(n book.Node) bool {
		switch n := n.(type) {
		case *book.Headline:
			return !(skipCommented && n.Commented)
		case *book.SrcBlock:
			if found == nil && n.Name == name {
				found = n
			}
		}
		return true
	})
	return found
}

func groupBlocks(d *book.Document, name string) []*book.SrcBlock {
	var group []*book.SrcBlock
	d.Walk(func(n book.Node) bool {
		switch n := n.(type) {
		case *book.Headline:
			return !n.Commented
		case *book.SrcBlock:
			if n.Params.NowebRef == name {
				group = append(group, n)
			}
		}
		return true
	})
	return group
}

// A referenced body arrives the way the block would have tangled it,
// minus the target: comma escapes gone, one trailing newline gone,
// and dedented unless the block preserves its indentation.
func bodyOf(d *book.Document, b *book.SrcBlock, open []string) (string, error) {
	body := unescapeCommas(string(d.Source[b.Body.Start:b.Body.End]))
	body = strings.TrimSuffix(body, "\n")
	if !b.Params.PreserveIndent {
		body = dedent(body)
	}
	if !b.Params.Noweb {
		return body, nil
	}
	return expandInto(d, body, open)
}

// The near miss is worth naming. Org folds case when it looks a name
// up and babble does not, so a reference that differs only in case
// tangles there and stops here, and the reader needs to see why.
func nearMiss(d *book.Document, name string) string {
	miss := ""
	d.Walk(func(n book.Node) bool {
		if b, ok := n.(*book.SrcBlock); ok && miss == "" {
			if strings.EqualFold(b.Name, name) && b.Name != name {
				miss = b.Name
			}
			if strings.EqualFold(b.Params.NowebRef, name) && b.Params.NowebRef != name {
				miss = b.Params.NowebRef
			}
		}
		return true
	})
	if miss == "" {
		return ""
	}
	return fmt.Sprintf("; %s differs only in case, and names here are case-sensitive", miss)
}

// A prefix that is present but empty is used as it stands. The elisp
// lookup treats the empty string as an answer; the Go habit of testing
// whether a value is empty would fall through to the hash.
type commentTable map[string]string

// The rows the toolchain's own books need. A language absent from
// the table takes the hash, which comments shells, Python, Ruby and
// make; the languages that would read a hash as code name themselves
// here or in the project's own file.
var builtinComments = commentTable{
	"go": "//", "js": "//",
	"emacs-lisp": ";;", "elisp": ";;",
	"lua": "--",
}

func loadCommentTable(bookPath string) (commentTable, error) {
	t := commentTable{}
	for lang, prefix := range builtinComments {
		t[lang] = prefix
	}
	path := filepath.Join(filepath.Dir(bookPath), ".lit", "comments.json")
	f, err := os.Open(path)
	if err != nil {
		return t, nil
	}
	defer f.Close()
	var raw map[string]any
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	for lang, v := range raw {
		prefix, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("%s: comment prefix for %s is not a string", path, lang)
		}
		t[lang] = prefix
	}
	return t, nil
}

// A prefix that is present wins even when it is empty, which is why
// this reads the second return value rather than testing for "".
func (t commentTable) prefix(lang string) string {
	if prefix, ok := t[lang]; ok {
		return prefix
	}
	return "#"
}

// A directory is created when any block in the unit asks for it, not
// when the first one does. Blocks sharing a target can disagree about
// ~:mkdirp~, and taking the first block's answer meant the file's
// directory depended on which block happened to come first.
func banner(lang string, comments commentTable) string {
	return comments.prefix(lang) + " Code generated from BOOK.org by make tangle. DO NOT EDIT.\n"
}

// The file is its blocks in order, a blank line between them unless
// one asks for none, and a newline after each. The banner and the
// blank line under it open the file, which is where the driver's
// hook puts them.
func assemble(d *book.Document, u unit, comments commentTable) ([]byte, error) {
	var out strings.Builder
	out.WriteString(banner(u.lang, comments))
	out.WriteString("\n")
	for i, b := range u.blocks {
		if i > 0 && b.Params.Padline {
			out.WriteString("\n")
		}
		if b.Params.Comments == "org" {
			comment, err := commentFor(d, b, comments)
			if err != nil {
				return nil, err
			}
			out.WriteString(comment)
		}
		body, err := bodyText(d, b)
		if err != nil {
			return nil, err
		}
		out.WriteString(body)
		out.WriteString("\n")
	}
	return []byte(trimBareMarkers(out.String(), comments.prefix(u.lang))), nil
}

// The driver trims these in a hook that runs over the whole tangled
// file once it is written, so it reaches block bodies as readily as
// doc comments. Applying it to the assembled file is what puts the
// two engines in the same place. With an empty prefix the driver's
// pattern matches a line holding one space, and emptying that line
// is reproduced rather than skipped: a rule that quietly differed
// here would be a second divergence wearing the shape of a fix.
func trimBareMarkers(content, prefix string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if line == prefix+" " {
			lines[i] = prefix
		}
	}
	return strings.Join(lines, "\n")
}

func writeTarget(bookPath string, u unit, content []byte) error {
	mkdirp := false
	for _, b := range u.blocks {
		if b.Params.Mkdirp {
			mkdirp = true
			break
		}
	}
	if mkdirp {
		if err := os.MkdirAll(filepath.Dir(u.target), 0o755); err != nil {
			return err
		}
	}
	if old, err := os.ReadFile(u.target); err == nil && bytes.Equal(old, content) {
		return nil
	}
	return os.WriteFile(u.target, content, 0o644)
}

func sameFile(a, b string) bool {
	pa, err := filepath.Abs(a)
	if err != nil {
		return a == b
	}
	pb, err := filepath.Abs(b)
	if err != nil {
		return a == b
	}
	return pa == pb
}
