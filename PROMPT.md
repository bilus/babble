# Babble: a self-contained tangler for lit/ books

You are building Babble, a Go reimplementation of the narrow
org-babel tangling subset used by the lit/ literate-programming
toolchain in ~/dev/bilus/reqsync. The authoritative oracle is batch
Emacs running org + babel + lit/tangle.el; `babble tangle` must
produce byte-identical results. This is a port of a documented
subset with an executable oracle, not a port of org-mode. When the
oracle and this prompt disagree, the oracle wins; when the oracle
and lit/AGENTS.md disagree, stop and ask.

Babble is itself a literate program built on the same lit/
toolchain: BOOK.org is the source of truth, all Go code (tests
included) tangles from it, and until babble reaches parity, batch
Emacs tangles babble's own book. Self-hosting is the finish line:
v1 is done when babble tangles its own book byte-identically.

## Ground rules

- Repo: this directory, module github.com/bilus/babble, latest
  stable Go, plain git on main until told otherwise.
- Bootstrap: copy lit/ from ~/dev/bilus/reqsync/lit (copy, not
  symlink; reqsync stays upstream of lit/ until it is extracted),
  then `go mod init github.com/bilus/babble` and
  `make -f lit/lit.mk setup` (go.mod makes it pick the go template),
  then grow BOOK.org from there.
- Treat ~/dev/bilus/reqsync as a read-only oracle corpus. Never
  modify it.
- The plan lives in the book, never in an external file, per
  lit/AGENTS.md: the Plan chapter with stage entries, badges, and
  holes is hole-driven-delivery's plan, the LOGBOOK trail plus the
  Changes journal are its ledger. PROMPT.md is a bootstrap brief,
  not a plan; do not create docs/plans/.
- Skills: invoke the hole-driven-delivery skill for the process
  (plan, typechecked skeleton of named holes, fill one hole at a
  time with its tests in the same commit, green after every fill,
  halt at stage boundaries) and the avoiding-ai-tells skill before
  writing anything a human reads. Stage review happens in the book:
  `make weave` renders the plan, the human ticks the checklist.
- Commits: imperative subject, one or two sentence body in first or
  third person. Never add Co-Authored-By or any trailer naming an AI.
  Commit only when the work is verified green.
- Prose style everywhere (book, README, comments, commit messages,
  test names): no em dashes; use parentheses, commas, periods.
  ASCII only, as lit/AGENTS.md requires for books.
- Shells: run `source ~/.zshrc` first; devbox provides go, emacs,
  make, diff.
- At the end of stage 1, write this repo's root AGENTS.md the way
  reqsync does it: a short stub pointing at lit/AGENTS.md, plus a
  babble-specific section for the oracle workflow (how to run
  ORACLE=1, how goldens are minted, reqsync's read-only status), so
  later sessions do not need PROMPT.md.

## Read these before writing anything

1. ~/dev/bilus/reqsync/lit/AGENTS.md, all of it. The toolchain
   section documents the org quirks the books rely on; the planning
   section defines the markup the dynamic blocks index.
2. ~/dev/bilus/reqsync/lit/tangle.el, the driver you are
   reimplementing: go-mode stub, docstring rule, comment-line
   trimming, generated-file banner, backup suppression, and the two
   dynamic-block writers (src-block-diff, plan-toc).
3. The two real corpora: ~/dev/bilus/reqsync/BOOK.org and
   ~/dev/bilus/reqsync/lit/example/BOOK.org.
4. The org sources, for exact semantics. Locate them with
   `emacs -Q --batch --eval '(princ (file-name-directory (locate-library "ob-tangle")))'`
   and read ob-tangle.el fully, plus these ob-core.el functions:
   org-babel-tangle, org-babel-tangle-collect-blocks,
   org-babel-tangle-single-block, org-babel-spec-to-string,
   org-babel-get-src-block-info, org-babel-parse-header-arguments,
   org-babel-balanced-split, org-babel-read,
   org-babel-expand-noweb-references, org-babel-find-named-block,
   and the org-map-dblocks / org-prepare-dblock / org-update-dblock
   trio in org.el.
5. Upstream org tests as coverage inspiration only:
   https://github.com/bzg/org-mode/blob/main/testing/lisp/test-ob-tangle.el
   They are GPL: do not copy fixtures, strings, or code. Write
   original fixtures covering the same behavior areas.

## What babble must do

`babble tangle BOOK.org` performs, in order, what
`emacs -Q --batch -l lit/tangle.el --eval '(lit-tangle "BOOK.org")'`
does:

1. Refresh every dynamic block in the book (plan-toc and
   src-block-diff), rewriting the book file only when its content
   changed.
2. Tangle every eligible src block into its target files.

Byte-parity scope: the tangled output files and the book file after
refresh, including the do-not-rewrite-identical-files behavior
(unchanged targets keep their mtime). stdout and stderr text need
not match the oracle.

## Architecture requirement: parse first, into a logical tree

The first implementation phase is a parser from .org into a logical
document tree: headlines (level, COMMENT keyword, tags), prose
runs, src blocks (name, language, switches, header arguments raw
and parsed, body), dynamic blocks (name, arguments, interior),
inline tasks, and keyword lines. The tangler and the dynamic-block
writers make every decision against this tree, never by re-scanning
org text. The tree is the seam for future input formats (markdown
is the likely second frontend), so keep org spellings out of the
core types: org-specific recognition lives in the frontend, generic
semantics (file assembly, noweb, comment extent, banners, dblock
refresh) live in the engine.

AST-first is not a departure from the oracle, it is how the oracle
already works, and this section records the investigation that
established it. org-babel's scan loop (org-babel-map-src-blocks)
uses org-babel-src-block-regexp only to find candidates; every hit
must be confirmed by org-element-context
(org-babel-active-location-p), and org-babel-get-src-block-info
then reads language, name, switches, header arguments, body, and
position from the org-element node's properties, not from the regex
captures. The element grammar is the authority for recognition and
attributes. The scanning regex has exactly three observable effects
of its own, and each has a home in babble:

1. A block only tangles if the scan regex saw it, and the regex
   demands a language token after #+begin_src. Recognition
   contract: a tangle candidate is a src-block node that has a
   language. Babble makes a languageless block an error instead of
   a silent skip (see deviations). Keyword matching is
   case-insensitive (the scan binds case-fold-search).
2. The :comments org extent is anchored on raw match positions, so
   parser nodes must record them: for each headline, the byte
   offset just past the stars and the single following space; for
   each src block, the offset where its #+begin_src line starts
   (post-affiliated, so a #+name: line above sits inside the
   extent) and the offset just past the literal #+end_src text on
   its closing line. The extent for a block is the raw slice from
   max(nearest headline above: after-stars offset; previous src
   block: after-end_src offset; neither: offset 0) to the block's
   begin-line start. The heading title itself falls inside the
   extent; only the docstring rule's last-paragraph selection keeps
   it out of tangled output.
3. The oracle's backward anchor search does not element-confirm,
   so block delimiters quoted verbatim inside example or export
   bodies would anchor extents in emacs but not in any AST. Lint
   them out of the subset (see deviations) so the divergence can
   never bite.

The document model is therefore the tree plus the retained source
bytes plus a line index. Every node carries its line number
(plan-toc links) and its byte span into the retained source
(extents, dblock interior splicing). The plan census that plan-toc
renders is defined by the lit convention on lines, not on
structure: the engine runs the sigil patterns over retained source
lines (translated from the Emacs regex dialect in tangle.el to Go
regexp), with the tree supplying splice points. Everything else in
the port is string algorithms that never needed a buffer: the
balanced header splitter, body normalization, noweb expansion over
a name index built from the tree, comment wrapping, banners,
assembly.

One more oracle fact keeps the port small: emacs -Q loads no
language packs, so org-babel-expand-body:LANG is never defined and
body expansion always takes the generic path, which is the identity
when there are no :var headers. Expansion in babble is the
identity, and :var is rejected anyway.

Give the parser its own fixtures early: a `babble parse --dump`
subcommand emitting a stable rendering of the tree (JSON or an
S-expression), exercised by testscript, so parser regressions
surface without the tangler.

## Semantics checklist (orientation, not a substitute for reading)

- Collection: walk src-block nodes in document order; skip blocks
  under COMMENT headings (keyword inheritance applies, an ancestor
  walk on the tree) and archived headings; group by resolved target
  file; a block with :tangle no is skipped, a block without a
  language is an error (see deviations).
- Targets: :tangle no | yes | path. Path resolves relative to the
  book's directory; yes derives the name from the book name plus the
  language extension map. :mkdirp creates parent directories.
- Header arguments: merge order is defaults, then language defaults,
  then property-inherited header-args, then the block's own header.
  Parsing uses the balanced splitter (respects (), [], and double
  quotes). The books use literal values only.
- Body pipeline, in order: the raw lines between the header line
  and the end line; drop exactly one trailing newline if present;
  remove common indentation unless preserve-indentation applies;
  noweb expansion when enabled; then a final trim (org-trim after
  another indentation removal), so trailing blank lines never
  survive into output.
- Assembly: blocks land in document order, exactly one blank line
  between blocks in the same file (padline), final newline at EOF.
- :comments org, the subtlest part: the comment text for a block is
  the raw slice defined by the extent anchors in the architecture
  section (after the stars of the nearest heading above, after the
  literal #+end_src of the previous block, whichever is later;
  offset 0 when neither exists; ending where the block's begin line
  starts), run through the driver's docstring rule: take everything
  after the last `# doc` marker line when present, else the last
  paragraph; drop `#+` keyword lines; strip a leading "Note: " or
  "Apropos: " sigil. The result is comment-wrapped in the block
  language's line-comment syntax; blank lines inside become a bare
  comment prefix with no trailing space. Reproduce the net effect
  of comment-region plus the driver's trim hook, byte for byte.
- Banner: every tangled file opens with
  `<prefix> Code generated from BOOK.org by make tangle. DO NOT EDIT.`
  followed by one blank line. Prefixes: .go gets //, .el gets ;;,
  .lua gets --, everything else #.
- Noweb (:noweb yes): <<name>> resolves to the named block, or to
  the concatenation of all blocks sharing :noweb-ref joined with
  :noweb-sep (default newline), in document order. A reference is
  single-line and its name neither starts nor ends with whitespace
  (org-babel-noweb-wrap). Text preceding <<name>> on its line is
  replicated as a prefix on every expansion line. Expansion is
  recursive; a cycle is an error. Blocks under COMMENT headings do
  not participate.
- Dynamic blocks: `#+begin: NAME args` dispatches on NAME; refresh
  deletes the interior wholesale and rewrites it. src-block-diff
  renders `diff -u` of two named block bodies with the two temp-file
  header lines dropped, wrapped in a src diff block, or the words
  "no differences". plan-toc emits the jump list: one item per plan
  site found by the census regexes, labels spelled without the
  sigils, line-number links, with the writer compensating for the
  lines it is about to insert (sites below the block shift by the
  item count). Refresh must be idempotent: a second run produces
  byte-identical output.

## Deliberate deviations (error loudly instead of imitating)

- Lisp-valued header arguments (values starting with (, ', `, or [)
  are rejected with a clear message; org evaluates them.
- An unresolvable noweb reference is an error; org silently expands
  to empty for most languages. A noweb reference cycle is an error;
  org recurses without bound.
- A src block without a language is an error; the oracle's scan
  regex cannot even see such a block, so emacs silently ignores it,
  which is worse than refusing.
- #+header: affiliated keywords are rejected; the oracle merges
  them into header arguments with precedence the books never rely
  on.
- Block or dynamic-block delimiters quoted verbatim inside example,
  src, or export block bodies are a lint error: the oracle's raw
  backward searches would anchor :comments extents on them, and no
  structural parser can agree with that.
- Unsupported constructs are errors that name the construct and the
  supported set: :comments link/both/noweb, :var, :shebang,
  :tangle-mode, :no-expand, coderef switches (-r/-l), :noweb
  strip-tangle, inline src blocks, #+INCLUDE.
- None of these errors may trigger on the two real corpora. Add a
  corpus test asserting exactly that.

## Diff engine

The oracle's src-block-diff writer shells out to `diff -u`. For
parity, v1 of babble invokes the system `diff -u` the same way and
applies identical post-processing. A pure-Go diff is a possible
later stage, and lands only if byte-identical across the whole
corpus or together with an explicit one-time regeneration story.

## Testing

- Framework: testscript from github.com/rogpeppe/go-internal, with
  txtar fixtures in testdata/script/. Each .txtar carries the input
  book plus any pre-existing files, and the expected outputs; the
  script runs `babble tangle` and cmp's every expected file,
  including the book itself after dynamic-block refresh. Use
  testscript's setup so `babble` runs as a command inside scripts.
- Book versus data: Go code, including the test harness, tangles
  from BOOK.org. The txtar fixtures under testdata/ are data files
  outside the book, like the dependency manifests: hand-started,
  machine-updated by the golden update mode, never tangled.
- Oracle harness: an oracle test mode, enabled by ORACLE=1 (skipped
  otherwise), that for every txtar extracts the inputs twice, runs
  the Emacs oracle command in one tree and babble in the other, and
  fails on any byte difference between the trees. Golden update
  mode, UPDATE=1, rewrites each txtar's expected outputs from the
  ORACLE tree, never from babble's, so committed goldens are
  emacs-produced by construction. Include a small runner script or
  make targets: test (pure Go), oracle, update-goldens.
- Real-book sweep: the oracle harness also copies the reqsync book,
  the lit/example book, the lit/BOOK.org toolchain book, and
  babble's own BOOK.org (with the lit/ driver for the oracle side)
  into temp trees and cross-checks both engines over them.
- Fixture coverage to reach, with original fixtures inspired by the
  upstream areas: block ordering across multiple target files;
  :tangle yes/no/path; COMMENT skip with inheritance; :mkdirp;
  padline and trailing-newline behavior; skip-write-when-identical
  (assert mtime unchanged on a second run); :comments org extent
  including a #+name line between prose and block, multi-paragraph
  docstrings with `# doc`, Note:/Apropos: sigils, blank lines
  inside comments, a block directly under a heading with no prose
  (the title lands in the extent), and a block before any heading
  or prior block (extent from offset 0); trailing blank lines in a
  body dropped by the final trim; case-variant #+BEGIN_SRC and an
  indented block with an indented #+end_src; banners per language;
  noweb named refs,
  :noweb-ref concatenation with custom :noweb-sep, multiple refs on
  one line, prefix interposition, recursive expansion; tangle-into-
  self guard; dynamic-block refresh idempotence including plan-toc
  line-number shift; a fixture book carrying full plan markup
  (badges, CriticMarkup, twin with :tangle no, holes) proving the
  twin never tangles and the census sigil dodges confuse nothing;
  every deviation error path.
- CI shape (GitHub Actions, devbox per the user's usual setup): job
  one runs the pure Go suite with committed goldens and no Emacs;
  job two provisions Emacs and runs ORACLE=1 over fixtures plus the
  real-book sweep.

## Acceptance for v1

- All testscript fixtures green without Emacs installed.
- ORACLE=1 reports zero byte differences across all fixtures and
  all three real books.
- Self-hosting: `babble tangle BOOK.org` on babble's own book
  reproduces the committed tangled sources and the refreshed book
  byte for byte, and `make check` agrees.
- `CGO_ENABLED=0 go build` produces a single static binary; go vet
  and go test -race are clean.
- The book documents usage, the subset contract, the deviation
  list, and the oracle workflow (including how to regenerate
  goldens); the weave is clean.

## Out of scope for v1

The weave side (preprocess.py, pandoc, tectonic), watching or
editing org files, a Windows diff story, and swapping lit/lit.mk
over to babble (that is a follow-up in reqsync once parity holds).
