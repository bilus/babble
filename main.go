// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// babble tangles lit books without Emacs. This file is tangled from
// BOOK.org.
package main

import (
	"io"
	"os"
)

const usage = `usage: babble <command> [flags] BOOK.org

commands:
  tangle  refresh dynamic blocks and write every eligible code block
  parse   print the book's logical tree (--dump for JSON)
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run dispatches the subcommands and owns the exit codes.
func run(args []string, stdout, stderr io.Writer) int {
	panic("HOLE(2): dispatch parse and tangle; parse --dump works after this fill")
}

// tangleCmd is the whole tangle flow in one place: parse the book,
// refresh its dynamic blocks, reparse the refreshed bytes, then
// write the targets. The reparse matters because refresh moves line
// numbers and the tangle must see the book it just wrote.
func tangleCmd(bookPath string, stderr io.Writer) int {
	panic("HOLE(5): parse, refresh, reparse, tangle; errors to stderr")
}
