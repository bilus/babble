// Code generated from BOOK.org by make tangle. DO NOT EDIT.

// babble tangles lit books without Emacs. This file is tangled from
// BOOK.org.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/bilus/babble/internal/book"
	"github.com/bilus/babble/internal/org"
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
	if len(args) < 1 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[0] {
	case "parse":
		fs := flag.NewFlagSet("parse", flag.ContinueOnError)
		fs.SetOutput(stderr)
		dump := fs.Bool("dump", false, "print the tree as JSON")
		if err := fs.Parse(args[1:]); err != nil {
			return 2
		}
		if fs.NArg() != 1 {
			fmt.Fprint(stderr, usage)
			return 2
		}
		d, err := org.Parse(fs.Arg(0))
		if err != nil {
			fmt.Fprintln(stderr, "babble:", err)
			return 1
		}
		if *dump {
			out, err := book.Dump(d)
			if err != nil {
				fmt.Fprintln(stderr, "babble:", err)
				return 1
			}
			if _, err := stdout.Write(out); err != nil {
				return 1
			}
		}
		return 0
	case "tangle":
		if len(args) != 2 {
			fmt.Fprint(stderr, usage)
			return 2
		}
		return tangleCmd(args[1], stderr)
	}
	fmt.Fprint(stderr, usage)
	return 2
}

// tangleCmd is the whole tangle flow in one place: parse the book,
// refresh its dynamic blocks, reparse the refreshed bytes, then
// write the targets. The reparse matters because refresh moves line
// numbers and the tangle must see the book it just wrote.
func tangleCmd(bookPath string, stderr io.Writer) int {
	panic("HOLE(5): parse, refresh, reparse, tangle; errors to stderr")
}
