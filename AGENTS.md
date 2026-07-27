# babble

BOOK.org is the program; the sources tangle from it. The working
rules for literate work live in lit/AGENTS.md, and the plan lives
in BOOK.org's Plan chapter; a session picking up work starts from
the census there, not by reading everything. reqsync
(~/dev/bilus/reqsync) is the read-only oracle corpus; nothing in
this repository ever writes there.

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
  committed goldens are Emacs output by construction.
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
