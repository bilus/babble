# Code generated from BOOK.org by make tangle. DO NOT EDIT.

LIT ?= lit

.PHONY: tangle weave review check setup

tangle:
	emacs -Q --batch -l $(LIT)/tangle.el --eval '(lit-tangle "BOOK.org")'

weave: book.pdf

book.pdf: BOOK.org $(LIT)/preprocess.py $(LIT)/plan.lua $(LIT)/notes.lua $(LIT)/mermaid.lua $(LIT)/svg.lua $(LIT)/texlinks.py
	python3 $(LIT)/preprocess.py BOOK.org book-weave.org
	t=$$(mktemp -d) && MERMAID_TMP=$$t pandoc book-weave.org -s --toc \
	  --lua-filter=$(LIT)/mermaid.lua --lua-filter=$(LIT)/svg.lua \
	  --lua-filter=$(LIT)/notes.lua \
	  --lua-filter=$(LIT)/plan.lua -V colorlinks -V linkcolor=NavyBlue \
	  -o book-weave.tex && \
	  python3 $(LIT)/texlinks.py book-weave.tex && \
	  tectonic book-weave.tex && \
	  mv book-weave.pdf book.pdf; rc=$$?; \
	  rm -rf "$$t" book-weave.org book-weave.tex book-weave.pdf; exit $$rc

review: weave

setup:
	@test ! -f BOOK.org || { echo "BOOK.org already exists"; exit 1; }
	@t="$(TARGET)"; \
	if [ -z "$$t" ]; then \
	  if [ -f go.mod ]; then t=go; \
	  elif [ -f requirements.txt ]; then t=python; fi; \
	fi; \
	test -n "$$t" || { echo "no go.mod or requirements.txt here; use TARGET=go or TARGET=python"; exit 1; }; \
	test -f "$(LIT)/templates/$$t/BOOK.org" || { echo "unknown TARGET '$$t'"; exit 1; }; \
	cp "$(LIT)/templates/$$t/BOOK.org" BOOK.org && \
	echo "BOOK.org created from the $$t template" && \
	{ test -f Makefile || printf 'include lit/lit.mk\n' > Makefile; } && \
	$(MAKE) -f $(LIT)/lit.mk tangle

check:
	@d=$$(mktemp -d) && mkdir $$d/lit && cp BOOK.org $$d && \
	  cp -R $(LIT)/* $$d/lit/ && ( \
	  cd $$d && emacs -Q --batch -l lit/tangle.el \
	    --eval '(lit-tangle "BOOK.org")' >/dev/null 2>&1 \
	) && diff -r -x BOOK.org -x lit -x example -x templates \
	  -x Makefile -x .gitignore -x go.mod -x go.sum \
	  -x requirements.txt -x README.md -x AGENTS.md -x book.pdf \
	  -x book-review.pdf -x docs -x .git -x .DS_Store \
	  $(CHECK_EXTRA) $$d . \
	  && echo ok
