# Literate programming

How to work on a literate program. The rules are generic; the examples
and the toolchain section are this repository's. This file is for
agents; lit/README.md is the human companion.

## Order serves the reader

This is the first rule and the one the rest of the file serves.

The book's order is the order that teaches. A file's order is the
tangler's problem. Put each piece of code where the concept it belongs
to is explained, and let named chunks assemble the file.

That is what literate programming is for, rather than a preference
about layout. Knuth's claim in 1984 was that a program "is best
thought of as a web instead of a tree", and that a programmer states
its parts "in whatever order is best for human comprehension, not in
some rigidly determined order like top-down or bottom-up". The
tangler exists so that choosing the reader's order costs nothing.

So the question is never "which file does this belong to". It is
"where will a reader need this". Someone meeting a rule needs three
things together: the rule, the reason for it, and the thing that
proves it. When those sit in three chapters because they tangle to
three files, the book has been arranged for the machine.

The deviations show it concretely. A deviation is a place the tangler
refuses what org accepts, and its test is a small book that should
trip it plus the error it should produce. All three belong in one
place. Gathering them into a fixtures chapter, away from the rules
they pin, buys nothing except that the chunks sit next to the other
chunks of their file, which is the tangler's convenience and not the
reader's. The file is assembled from chunks named at each site:

    #+begin_src txtar :tangle testdata/script/refusals.txtar :noweb yes
    <<refusal-scripts>>
    <<refusal-books>>
    #+end_src

Each deviation adds to both groups where it is explained, and the
tangler produces one file.

Grouping by output file is the commonest way to lose the point, and
the easiest to catch: if the only thing a chapter's contents have in
common is their destination, it should not be a chapter.

One limit. A file the machine owns cannot live in the book, because
the book is a document a human owns and a machine would rewrite it.
A golden minted from an oracle is such a file. A fixture written by
hand is not, and the difference is who edits it, not whether it is a
test.

## The contract

- The book is the program. Code changes are edits to the book's code
  blocks; sources are regenerated with `make tangle` and must come out
  exactly as a direct edit would have produced them. Never edit
  generated files directly.
- Verification is always: tangle, then build and test the tangled
  tree, then `make check` (tangle into a scratch directory, diff).
  Byte-identity is the fact-checker behind every claim that the book
  is the program.
- The weave (`make weave`) must stay clean too; the rendered book is a
  deliverable, not a byproduct.

## Structure of the book

- Top-down: the entry point first (`main`), then the engine it drives
  (`sync.Run`), then the packages the engine stands on. Named chunks
  let prose forward-reference code that is defined where it is easiest
  to explain.
- Every code block gets one to three sentences of why directly before
  it: context or decision, never a paraphrase of the code. When
  narrative and doc comment start repeating each other, the doc
  comment wins and the narrative shrinks.
- Long functions become skeletons of named regions
  (`<<sync-create-loop>>`, `<<doc-link-helpers>>`). This is
  extract-block, not extract-function: the decomposition a refactoring
  would give you, without moving any code. Cut regions at blank-line
  boundaries; give each region prose that answers the questions a
  first-time reader would actually ask.
- To learn those questions, spawn a context-free subagent on the plain
  sources: "which functions are hardest to follow, and what would you
  ask?" Write the region prose against its report, and let it tell you
  which long functions are fine as they are. A second, adversarial
  reviewer subagent feeds a "Known issues" chapter near the end of the
  book: a registry of real, verified bugs and limitations, each with
  its failure scenario.
- Tests live in a back-matter part, one section per chapter, linked in
  both directions. Chapters stay about the design; test sections say
  what is pinned and why. Within a section, each test is its own
  named, described block, and the chapter's claims footnote to them
  (see Prose).
- The build glue (lit/lit.mk) is tangled from the toolchain's own
  book, lit/BOOK.org, never from a consuming project's; the one-line
  Makefile that includes it, .gitignore, and the dependency
  manifests are hand-managed outside the book.

## Writing style

How to write and edit book prose, in working order:

1. Load the writing-like-user skill before drafting a sentence, and
   keep its voice throughout: direct, conversational, varied sentence
   length, dashes for asides. The skill is the voice; the rules below
   are structure.
2. Every chapter opens by stating its goal, in one or two sentences
   the reader can test the rest against. Every section has one
   overarching topic. Two topics may share a section when they
   genuinely pair, but never as a sandwich: A, then B, then back to
   A means the section wants splitting or reordering.
3. Lead from high level to detail, in the book, in each chapter, and
   in each section. Nothing dips into a mechanism the surrounding
   text has not framed first.
4. The book's Introduction presents the idealized system: inputs,
   outputs, and the main principles, in common IT shapes (a
   pipeline, views of one source) a developer can picture at once.
   No edge cases, no hard features. Each hard problem opens the
   chapter that solves it, as that chapter's motivation---the
   mutual-reference deadlock opens the sync engine, not the book.
5. Introduce every book-specific term once, before its first
   load-bearing use. Keep a term table while editing: term, where
   introduced, where first used. Terms a working developer already
   owns (AST, fixture, regex) get no introduction; wasting a
   paragraph on one is as bad as an unexplained coinage.
6. Open a paragraph with its summary sentence and let the rest
   unpack it; close with the consequence or moral when one exists.
   Vary the shape (an occasional question opener, a two-sentence
   paragraph with no moral), because the pattern applied ritually
   numbs the reader.
7. No flowery language: no personification of the program, no
   metaphor that outlives its sentence, no applause lines. When a
   phrase reads like it wants admiration, delete it and state the
   fact it was decorating.
8. Never justify a design by its history. A chapter argues from the
   problem, not from the bug that prompted the work or the version
   that came before. "X, because the old Y drifted" is a journal
   entry wearing a chapter's clothes: cut the clause, keep the
   reason that survives without it. The test is whether a reader who
   has never seen an earlier version notices anything missing; if
   the sentence only lands for someone who watched the change, it
   belongs in `* Changes`.

## Prose

- Voice and structure follow "Writing style" above. The habits that
  survive from the older asimov reference: open on the real puzzle,
  state the key early, put a mechanism behind every fact, at most
  one rhetorical hinge per section.
- Apply avoiding-ai-tells to everything a human reads: prose, code
  comments, commit messages. ASCII only.
- Doc comments are prose. The paragraph directly above a code block is
  that block's doc comment and tangles back into the source, line
  breaks preserved. An aside carries a "Note:" or "Apropos:" sigil:
  styled distinctly in the weave, stripped on tangle, landing in the
  source as an ordinary comment.
- Document decisions where they bite, in the chapter that owns the
  code. Sometimes a Note aside is the right form, precisely because it
  reaches the source too.
- A behavioral claim links to the test that pins it with an org
  footnote: the reference rides in a narrative paragraph, never in a
  doc-comment paragraph (those tangle into source), and never inside
  CriticMarkup braces (the weave turns those into raw LaTeX) - after
  the closing brace is fine. Definitions collect at the end of the
  file, so test names never sit in running text.
- Every test lives in its own block, named after it
  (`#+name: TestLoadRejectsCRLF`) and carrying `:comments org`, with
  a short paragraph directly above it: why the test exists, not how
  it works, opening with the test's name in Go doc style. That
  paragraph tangles as the function's doc comment, so it stays
  footnote-free. A footnote definition links straight to the block
  with `[[TestLoadRejectsCRLF]]`: the link jumps there in the
  buffer, and the weave turns it into a PDF link (notes.lua repairs
  what pandoc leaves spurious). Helper functions ride in unnamed,
  uncommented blocks beside their tests.

## Process

- Changes follow hole-driven-delivery: a plan with numbered decisions,
  a skeleton of the new seams as holes in the book, reproducing tests
  written first and seen to fail, one hole at a time, the suite green
  at every boundary, a halt for review at each one. The plan lives in
  the book, never in an external file: the Plan chapter is
  hole-driven-delivery's plan, and the LOGBOOK trail plus the Changes
  journal are its ledger. Holes and chunks compose naturally: a hole
  is a block whose body panics. The book-side format for plans, twins,
  badges, and the journal is "Planning changes in the book" below.
- The skeleton commit is a gate, and the gate's command is
  `make review`. After committing a skeleton, weave it (`make review`
  renders the book with the plan styled: badges boxed, twins beside
  their originals, generated diffs, CriticMarkup in color), put
  book.pdf in front of the human, and stop. Before committing, check
  that every twin has a src-block-diff under it: a skeleton whose
  twins show only before-and-after blocks is not reviewable and is
  not ready for the gate; acceptance work starts
  only after the human ticks "skeleton approved". An agent never
  ticks that box for a plan it wrote. The one exception: a skeleton
  the human authored and committed themselves is pre-approved, and
  being told to execute it is the tick.
- A test that has never failed proves nothing. When the code a test
  guards already works, do the "seen to fail" step by breaking the
  guard on purpose: make the production block violate the guarantee,
  watch the test that names it fail, then revert. One break at a time,
  every test the stage adds or widens, and the stage entry says it was
  done. Expect some tests not to survive. A test written against
  working code passes on the day it is written whether it guards
  anything or not, and the two usual causes are a fixture that trips
  an earlier check, so the assertion reads an error the test did not
  cause, and a guarantee the code cannot report on, such as a
  swallowed error. The second is a bug in the code, not a reason to
  skip the check.
- Commit bodies are one or two sentences, first person or third
  person. ("I added X so that Y." / "This commit fixes Z.")

## Planning changes in the book

hole-driven-delivery says what happens in what order; this section
says where each piece of a planned change lives in the book and what
it looks like, so a reviewer sees intent exactly where it will land.
Asked to write a plan for a literate program, write a stage entry
here: a markdown file under docs/plans/ or anywhere else outside the
book is never the plan. The rules are generic; the examples are
reqsync's stage 4, crash-safe write-back.

A book under change has three layers, each with a lifetime and a
voice:

- Steady-state text: the program and its narrative. Permanent, and it
  speaks from no date, as if the change had always been there. It
  also speaks from no history: a chapter explains what the code does
  and why that is the right shape, never what it replaced, what used
  to be broken, or which bug prompted it. The reader arrives at a
  finished program and needs no archaeology to read it. Motivation
  comes from the problem the code solves, not from the project's
  past, however good the war story. The war story goes in the
  journal, where its date makes it true.
- Transition scaffolding: the Plan chapter, badges, CriticMarkup,
  twin blocks, hole bodies. Acceptance deletes it. This is the only
  place for "now", "no longer", "still", or any other comparison with
  the past.
- The journal: a permanent `* Changes` chapter near the back. Entries
  speak from their date, which is what makes change-talk permanent
  there.

Voice must match lifetime, and the test is mechanical: acceptance is
a verbatim unwrap, never a reword. Delete the red, keep the green,
swap two tangle headers, move the dated draft to the journal, delete
the plan entry. If any of that would require editing the surviving
text, the text sits in the wrong layer.

Setup, once per book: a todo line in the header,

    #+todo: PLANNED(p!) HOLE(h!) FILLING(f!) | LIVE(l!) DROPPED(d@)

`(require 'org-inlinetask)` wherever the book is edited, and the
permanent `* Changes` chapter. The `!` markers timestamp every state
change into LOGBOOK drawers: the in-flight ledger. It dies with its
plan entry, on purpose; the journal entry is the distilled survivor.

The Plan chapter sits first in the file and is temporary: it exists
only while work is in flight. One entry per stage, holding the
stage's claim (one sentence, behavior-shaped), a properties drawer
(CUSTOM_ID for badge links, SPEC, ACCEPTANCE), the boundary
checklist, a Changes draft written in the journal voice, and a census
block. Collect entries with `C-c / T PLANNED` or the agenda.
reqsync's stage 4, condensed:

    ** PLANNED Stage 4: crash-safe write-back                  :stage4:
    :PROPERTIES:
    :CUSTOM_ID: stage-4
    :SPEC:     Known issues, "The multi-file crash window is wider ..."
    :ACCEPTANCE: TestAtomicWrite, TestWriteBackDefiningFileFirst; go test ./...
    :END:

    Claim: after this stage, a crash at any moment of write-back
    leaves every requirements file either whole-old or whole-new, and
    the defining file always lands before any referencing file.

    - [ ] skeleton approved (twin at [[sync-write-back--planned]])
    - [ ] inserted prose reads steady-state: no "now", "no longer",
      "still", and no account of what the change replaced or fixed
    - [ ] Changes draft below reads right in the journal
    - [ ] HOLE(4) atomicWrite filled, tests in the same commit
    - [ ] twin accepted: :tangle headers swapped, old block DROPPED
    - [ ] census clean, suite green
    - [ ] Changes entry dated and filed, plan entry deleted

The first box is the human's: after the skeleton commit,
`make review` (see Process) is how they see the plan before anything
is accepted.

Code changes come in two cases. An addition tangles immediately with
a hole body: it compiles, fails loudly if reached, and nothing
reaches it until something planned calls it, so the suite stays
green. The panic string is the census unit, `HOLE(<stage>):
<contract sentence>`:

    func orderDefiningFirst(files []*req.File, defining string) []*req.File {
        panic("HOLE(4): defining file first, remaining files in set order")
    }

A modification to code the suite executes cannot be holed in place; a
reachable panic breaks the green build. It gets a twin: a sibling
block named `<name>--planned` with `:tangle no`, directly below the
live block, carrying its own doc-comment paragraph (already
steady-state). The live block keeps tangling; the twin waits.
Acceptance swaps the two blocks' `:tangle` headers and retangles;
rejection deletes the twin. Nothing else moves either way.

A badge marks every place a human must look: a PLANNED inline task,
15 stars, directly before the paragraph that carries the change,
saying why the change is made, linking the plan entry, carrying the
stage tag. Before the paragraph, never between a paragraph and its
block, because the paragraph directly above a block is its doc
comment. Every change cluster starts with one, including code whose
only change is added holes and CriticMarkup-bearing paragraphs;
there is no separate HOLE badge, since the hole strings and the
census already carry the mechanics:

    *************** PLANNED stage 4 replaces the block above ([[#stage-4][plan]]) :stage4:
    The twin below is the skeleton. It must not tangle yet: the
    end-to-end test drives writeBack, and a body that reaches
    atomicWrite's panic would turn the suite red.
    *************** END

Every twin carries a generated view directly under it, without
exception: a dynamic block naming the pair,

    #+begin: src-block-diff :old sync-write-back :new sync-write-back--planned
    #+end:

This is not decoration and not optional. A twin without one asks the
reviewer to diff two blocks by eye, which is the one job a machine
does perfectly and a human does badly; on a stage with a dozen twins
it makes the change unreviewable, and a one-word edit buried in forty
identical lines is exactly what slips through. The reviewer should be
able to read the stage by reading its diffs alone. When a twin's
delta is genuinely the whole block (a new function, a rewritten
region), the diff says so and costs nothing.

Org dispatches on the name alone: a refresh deletes the interior
wholesale, then calls `org-dblock-write:src-block-diff`
(lit/tangle.el here) with point inside, and whatever it inserts is
the new content.
This writer resolves the two `#+name:`-ed blocks, runs `diff -u` on
their bodies, drops the `---`/`+++` lines (they only name temp
files), and wraps the rest as a `src diff` block, so both Emacs and
the weave fontify it. `C-c C-c` on the begin line refreshes one;
tangling refreshes all of them first, so a committed diff cannot
disagree with the blocks it describes. Hand edits are futile, not
just forbidden: the next refresh clobbers them. If you write another
writer, note the begin-line params are read as lisp, so
`:old sync-write-back` arrives as a symbol, not a string.

Prose changes use CriticMarkup, in narrative paragraphs only:
`{++new++}`, `{--old--}`, `{~~old~>new~~}`, `{>>comment<<}`. Never
inside a doc-comment paragraph; those tangle into source, and a twin
carries its own replacement doc comment anyway. Inserted text is
future document text; rationale rides in a `{>>...<<}` note or the
plan entry. Stage 4's insertion:

    {++Every write stages through a temp file and the defining file
    lands first, so a crash between two files leaves whole files, and
    no reference is substituted before its number is on disk.++}
    File modes survive the rewrite.

The census: `grep -rn "HOLE(" --include="*.go" .` over the tangled
tree lists exactly the holes that ship; spell the pattern `HOLE[(]`
anywhere it could match itself. Org-side, `C-c / T HOLE` walks the
heading and badge keywords.

Finding the plan is a grep, never a read-through: every kind of
transition markup carries a sigil that steady-state text cannot
contain, so one pattern indexes the whole layer. The kinds and their
sigils, as extended-regexp fragments:

- stage entries and state headings: `^\*+ (PLANNED|FILLING|HOLE|LIVE|DROPPED) `
- badges, including their END lines: `^\*{15,} `
- hole bodies, and checklist lines naming them: `HOLE[(]`
- prose edits: `\{[+][+]`, `\{[-][-]`, `\{[~][~]`, `\{[>][>]`. Dodge
  with bracket classes, not backslashes: characters like `-` need no
  escaping, so `\{--` still contains the raw `{--` sigil and matches
  itself.
- twins and every reference to them: `[-]-planned` (the dodged
  spelling of the `--planned` name suffix)
- generated views: `^#\+begin: src-block-diff`
- anything tagged for the plan: `:(plan|stage[0-9]+):`

The combined pattern lives in the Plan chapter's census block, which
is the authoritative copy: run it there with `C-c C-c`, or paste it
into a shell. Scope it to the book file; example blocks in this file
would pollute a repo-wide sweep. To narrow to one stage, filter the
result by that stage's discriminators: its tag (`stage4`), its
CUSTOM_ID (`stage-4`), and its hole ids (`HOLE(4`).

Humans get finders the same machinery keeps honest. A
`#+begin: plan-toc` dynamic block in the stage entry renders a
generated jump list, one item per site: in the buffer the links jump
to the line (`file:BOOK.org::794` opens on line 794), the labels are
spelled without the sigils so the census never matches its own
index, and every tangle refreshes it like all dynamic blocks. The
weave replaces the file links, which are external and dead in
a PDF, with internal ones: every badge carries a hypertarget, and
each index item links to its governing badge, the nearest badge
above the site. Badge order is the one thing both sides can count
identically, which is what makes those links safe to generate. The
margins still flag every site: badge boxes carry a colored bar down
their left edge plus a margin mark naming the stage, and every
paragraph holding CriticMarkup gets a small margin mark of its own.
The badges also surface in the PDF chrome: a "Plan sites" branch in
the document outline (the Bookmarks sidebar) lists every badge, and
each badge plants a sticky-note annotation carrying its why, so
viewers list the plan under Highlights and Notes too. Whole-line
background highlighting belongs to the editor, not the weave:
font-lock with :extend does it in Emacs, while LaTeX backgrounds
fight line breaking for little gain.

Completion, in order: check every box, date the Changes draft with an
inactive timestamp, file it newest-first in `* Changes`, fold any
retired Known-issues entry into its narrative, delete the plan entry.
Journal entries are written once and never reworded; a correction is
a new entry.

The tooling needs none of this rendered: tangle, build, tests,
census, and sparse trees all work on the plain org. The weave
renders it: pandoc alone ignores custom todo keywords, inline tasks,
and CriticMarkup, so `make weave` first runs lit/preprocess.py,
which rewrites badges and CriticMarkup into org pandoc understands,
and lit/plan.lua styles them (badges as ruled boxes, CriticMarkup in
red and green, state keywords colored). A book with no plan in
flight weaves as a plain book, and `make review` is the same weave
under the gate's name: run it at the skeleton commit and put
book.pdf in front of the human. The strikeout guard in lit/notes.lua
stays for pandoc run by hand without the preprocessor: org parses a
short CriticMarkup span as strikethrough and hands it to soul's
\st, which cannot typeset it. Dynamic blocks need no help: pandoc
drops the wrapper lines and renders the inner diff highlighted. A
release weave, stripping the transition layer instead of rendering
it, remains unwritten.

## This repository's toolchain (org + babel)

- `make tangle`, `make weave`, `make check`, and `make review`, the
  weave run at the review gate; the targets live in lit/lit.mk. The
  one-line Makefile including it is hand-written, outside the book
  (it may also set `CHECK_EXTRA` with extra `-x` exclusions for the
  check diff). The toolchain is `lit/`; the driver is lit/tangle.el.
- lit/ is itself a literate program: lit/BOOK.org tangles lit.mk,
  tangle.el, preprocess.py, notes.lua, plan.lua, mermaid.lua, and
  texlinks.py. To change the toolchain, edit lit/BOOK.org and run
  `make -C lit tangle`; never edit those scripts directly, and never
  tangle lit.mk from a consuming project's book.
- `make setup` starts a fresh project: it copies
  lit/templates/<target>/BOOK.org into the working directory (go when
  go.mod exists, python when requirements.txt does, TARGET= forces)
  writes the one-line Makefile when none exists, and runs the first
  tangle. Dependency manifests (go.mod, go.sum, requirements.txt),
  the Makefile, and .gitignore live outside the book, and the check
  diff excludes them all.
- Every tangled file opens with a "Code generated from BOOK.org ...
  DO NOT EDIT." banner in its own comment syntax, stamped by the
  driver; edits belong in the book, always.
- Files assemble from sequential same-target blocks in document order,
  which org separates with exactly one blank line (Go's declaration
  spacing). In-function regions use noweb references, one level deep.
- The docstring convention is a custom
  `org-babel-process-comment-text`: the last paragraph above the
  block, or everything after an invisible `# doc` marker when the
  docstring has several paragraphs.
- Batch-Emacs traps tangle.el already handles: no go-mode present (a
  stub supplies `//` comment syntax), `:mkdirp` off by default,
  `#+name:` lines leaking into the grabbed comment text, and
  comment-region padding empty comment lines to `// `.
- Org trims a block body's leading blank lines, and noweb indents
  content lines but not blank lines. Consequences: cut regions at
  blank lines, and a standalone comment separated from its code by a
  blank line cannot be promoted to prose; leave it in the block.
- The weave is preprocess.py, then pandoc with mermaid.lua,
  notes.lua (Note/Apropos styling, listing captions, noweb-reference
  placeholders), and plan.lua, then texlinks.py over the LaTeX, then
  tectonic. Blocks that tangle get numbered Listing captions naming
  the target file and the block's name; noweb references inside
  listings weave as internal links. Name delivered PDFs uniquely,
  ideally commit-stamped; publish an HTML artifact when someone
  needs the book on a phone.
- lit/example is a complete toy book, in Python, exercising the whole
  convention, and doubles as the toolchain's smoke test:
  `make -C lit/example tangle check`, or the same targets from inside
  the directory.
