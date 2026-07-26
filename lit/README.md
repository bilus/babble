# lit/ - the literate toolchain

BOOK.org is the program. Everything in this directory turns that one
file into sources, PDFs, and a reviewable plan: lit.mk carries the
make targets (the top-level Makefile just includes it), tangle.el
drives org-babel in batch and defines the dynamic blocks, notes.lua
and plan.lua style the weaves, and preprocess.py rewrites plan markup
into org that pandoc understands. Agents should read AGENTS.md in
this directory; this file is for people. example/ is a complete toy
book exercising the whole convention, and doubles as the toolchain's
smoke test: `make -C example tangle check`, or the same targets from
inside the directory.

## The flow

Edit BOOK.org, never the generated files. Then:

    make tangle    # regenerate sources, refresh dynamic blocks
    go test ./...  # the tangled tree must stay green
    make check     # tangle into a scratch dir and diff: book == program
    make weave     # book.pdf, with any in-flight plan rendered

A book with no plan in flight weaves as a plain book; make review is
an alias of weave. Everything here is local and safe to run at any
time.

## Starting a new Go project

    mkdir app && cd app && git init
    go mod init example.com/app
    cp -R /path/to/lit .        # vendor the toolchain
    make -f lit/lit.mk setup    # go.mod detected: copies the go template
    make tangle && make check

setup refuses to overwrite an existing BOOK.org, copies the
template, writes the one-line Makefile when none exists, and runs
the first tangle; from then on plain make works. go.mod, go.sum,
.gitignore, and the Makefile itself stay outside the book: their
tools and your hands manage them, and the check diff leaves them
alone. Retitle BOOK.org and grow the chapters from there.

## Starting a new Python project

Same shape with a requirements.txt in place of go.mod; setup then
picks the python template, and requirements.txt stays outside the
book the same way. When the directory has neither manifest, or both,
force the choice: make -f lit/lit.mk setup TARGET=python (or
TARGET=go).

## Reading a review

While work is in flight, the book carries its own plan, and the
woven book.pdf gives you five ways at it:

- The Plan chapter, first in the book: the stage claim, the boundary
  checklist, the Changes draft (the journal entry the stage will
  leave behind), and a jump list of every site. The blue line numbers
  link to the badges.
- Bookmarks: the "Plan sites" branch of the PDF outline lists every
  PLANNED badge.
- Highlights and Notes: each badge doubles as a sticky note carrying
  its why.
- The margins, while flipping: a colored bar and a stage mark on
  every badge box, a small mark on every paragraph with red and green
  edits.
- Red and green in the text itself: deletions struck through,
  insertions colored, with the rule that the green text must read as
  if the change had always been there.

In Emacs the same jump list opens on the exact line, `C-c / T
PLANNED` builds a sparse tree of the stage entries, and `C-c C-c` on
the census block prints every hole in the sources and every
plan-markup line in the book.

## What accept and reject mean

Acceptance is a verbatim unwrap, never a rewording: swap the
`:tangle` headers between a live block and its `--planned` twin and
retangle; keep the green text, delete the red and the marks; date the
Changes draft and file it in the Changes chapter; delete the plan
entry and its badges. Rejection deletes the twin and the marks, and
nothing else moves. The full rules with examples are in AGENTS.md
here.

## Working with AI

Point the agent at lit/AGENTS.md and it has the whole contract; the
short version for the human supervising one:

- Ask for changes to the program, never to the sources. The agent
  edits BOOK.org and retangles; every tangled file says "Code
  generated ... DO NOT EDIT." in its first line, and a hand edit
  there is a review red flag.
- Plans arrive inside the book, never as separate documents: a stage
  entry in the Plan chapter with a claim, a checklist, and a draft of
  the journal entry it will leave behind, plus a PLANNED badge at
  every site the change touches.
- Review by weaving: make weave renders the plan in place, and the
  jump list, margin marks, Bookmarks branch, and sticky notes take
  you to each site. Accepting a stage is mechanical (swap two :tangle
  headers, file the journal entry, delete the plan entry); rejecting
  deletes the twin and the marks, and nothing else moves.
- Trust the greps: the census block indexes every hole and every
  plan-markup line, so neither you nor the agent hunts by reading.

## One-time Emacs setup

    (require 'org-inlinetask)   ;; badges are inline tasks
    (load "lit/tangle.el")      ;; dynamic-block writers, C-c C-c refresh

Batch runs need neither: make tangle loads the driver itself.
