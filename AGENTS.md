# babble

BOOK.org is the program; the sources tangle from it. The working
rules for literate work live in lit/AGENTS.md, and the plan lives
in BOOK.org's Plan chapter; a session picking up work starts from
the census there, not by reading everything. reqsync
(~/dev/bilus/reqsync) is the read-only oracle corpus; nothing in
this repository ever writes there.

## Changing code

The book carries a change before the sources do, every time. Per
stage, and per change to live code that comes out of a review:

1. Put the change in the book first. A badge at every site a human
   must look at, a `--planned` twin with `:tangle no` beside each
   block whose body changes, a named hole for each addition, and
   the stage entry that ties them together.
2. Commit that, weave it, halt. The human reviews intent at the
   sites, in the rendered book.
3. Only then apply: swap the twins' tangle headers, fill the holes
   with their tests, and delete what acceptance retires.

Approval of the plan is not approval of a stage's changes, and a
finding raised in review is a plan diff, not a licence to edit. An
addition needs no twin, since a hole body compiles and nothing
calls it; a modification to code the suite executes always does.

## Gates

The review gates follow lit/AGENTS.md: the human approves plans
and skeletons, and an agent never approves work it authored. When
the human states approval in conversation ("approved", "skeleton
approved"), the agent ticks the matching checkbox in the book's
plan entry on their behalf; the chat message is the approval, the
tick is clerical. The other checklist boxes are verification
items: the agent ticks each one only at the moment it has actually
verified it, output in hand.

## Every stage plans before it applies

A stage lands in two commits with a gate between them, never one.

1. The planned layer. Badges (fifteen-star PLANNED inline tasks
   tagged with the stage) at every site the stage touches, holes
   for new code, a `--planned` twin beside every live block the
   stage modifies, CriticMarkup for prose edits, and the plan
   entry's own jump list refreshed. The suite stays green because
   twins carry `:tangle no` and holes are unreachable. Commit,
   `make review`, hand over the PDF, stop.
2. Acceptance, only after the human approves: fill the holes,
   swap the twins' `:tangle` headers, delete the badges and the
   dead blocks, tick the boxes, file the Changes entry.

Filling a hole the human already approved needs no twin; the hole
is the approved scaffolding. Changing a block that currently
tangles always needs one, however small the edit, including edits
the human asked for in conversation. When in doubt, plan it: a
twin costs one commit, an unreviewed change costs the gate.

## The oracle workflow

babble must match batch Emacs byte for byte, and the test suite
enforces it in two layers:

- `go test ./...` runs the pure suite against committed goldens;
  no Emacs needed.
- `ORACLE=1 go test ./...` also runs every fixture through batch
  Emacs (lit/tangle.el) and through babble in two parallel trees,
  failing on any byte difference.
- `UPDATE=1 ORACLE=1 go test ./...` remints each fixture's
  expected outputs from the Emacs tree, never from babble's, so
  committed goldens are Emacs output by construction. UPDATE
  without ORACLE is an error: there is nothing else to mint from.
- The oracle mode adds to the pure suite rather than replacing it,
  so with ORACLE=1 babble runs each fixture twice, once against
  its goldens and once against Emacs.
- The harness arrives at stage 6 of the plan; before that these
  switches are seams in internal/oracletest, not features.

Verification for any change: `make tangle`, `go build ./...`,
`go vet ./...`, `go test ./...`, `make check`. The census is
`grep -rn "HOLE[(]" --include="*.go" .` over the tangled tree.

## Writing style

Correctness outranks style. The contract chapters say exactly what
the oracle does; no edit may trade that precision away.

Apply the styles in this order, later rules winning on conflict:

1. Structure from the asimov style reference (the
   asimov-tech-article skill, references/asimov-style.md):
   windowpane prose, a chapter opens with its goal, high level
   before detail, a mechanism behind every fact.
2. Voice from the writing-like-user skill: short declarative
   sentences mixed with longer reasoning ones, sentences that may
   start with And or But or So, an occasional question answered
   immediately, first person where it is honest, parenthetical
   asides. No contrastive constructions ("not X, it is Y").
3. avoiding-ai-tells throughout.
4. Typography, the hard rules: ASCII only, no em dashes (use
   parentheses, commas, periods), fill near 70 columns. These win
   over anything a skill says, the em-dash advice especially.

Paragraphs open with the point in one plain sentence; detail
follows; a closing consequence is optional. Vary the shape, since
a page of same-shaped paragraphs tires the reader. Every chapter
states one goal; every section holds one topic; never return to
topic A after leaving it for topic B (merge or split instead).

Terms: introduce every book-specific term before its first use, in
one clause or sentence. The reader is a software developer:
explain org, babel, and lit conventions; never explain general
programming. The Plan chapter sits first in the file by
convention, so a term's first plan-chapter use carries a short
gloss and the Introduction or a contract chapter owns the real
definition.

Tone: no quips, no metaphors, no anthropomorphism. Dry statements
and concrete numbers. At most one light aside per chapter, none
inside contract text.

Mechanics of a prose pass over the book:

- Never touch code block bodies, dynamic block interiors, property
  drawers, checklists, or heading lines with their tags.
- Keep the comma-escaping in examples exactly as it is.
- Prose must never match the census sigils: spell holes dodged
  (HOLE[(]) when naming them in running text, avoid CriticMarkup
  brace pairs, never write the two-hyphen twin suffix as one
  token.
- After any edit: make tangle, make check, then grep the book for
  non-ASCII bytes and em dashes; all four must come back clean.
