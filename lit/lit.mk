# Code generated from BOOK.org by make tangle. DO NOT EDIT.

.PHONY: tangle weave review check setup

tangle:
	emacs -Q --batch -l lit/tangle.el --eval '(lit-tangle "BOOK.org")'

weave: book.pdf

book.pdf: BOOK.org lit/preprocess.py lit/plan.lua lit/notes.lua
	python3 lit/preprocess.py BOOK.org book-weave.org
	pandoc book-weave.org -s --toc --lua-filter=lit/notes.lua \
	  --lua-filter=lit/plan.lua -V colorlinks -V linkcolor=NavyBlue \
	  --pdf-engine=tectonic -o book.pdf
	rm -f book-weave.org

review: weave

setup:
	@test ! -f BOOK.org || { echo "BOOK.org already exists"; exit 1; }
	@t="$(TARGET)"; \
	if [ -z "$$t" ]; then \
	  if [ -f go.mod ]; then t=go; \
	  elif [ -f requirements.txt ]; then t=python; fi; \
	fi; \
	test -n "$$t" || { echo "no go.mod or requirements.txt here; use TARGET=go or TARGET=python"; exit 1; }; \
	test -f "lit/templates/$$t/BOOK.org" || { echo "unknown TARGET '$$t'"; exit 1; }; \
	cp "lit/templates/$$t/BOOK.org" BOOK.org && \
	echo "BOOK.org created from the $$t template" && \
	{ test -f Makefile || printf 'include lit/lit.mk\n' > Makefile; } && \
	$(MAKE) -f lit/lit.mk tangle

check:
	@d=$$(mktemp -d) && mkdir $$d/lit && cp BOOK.org $$d && \
	  cp -R lit/* $$d/lit/ && ( \
	  cd $$d && emacs -Q --batch -l lit/tangle.el \
	    --eval '(lit-tangle "BOOK.org")' >/dev/null 2>&1 \
	) && diff -r -x BOOK.org -x example -x Makefile -x .gitignore \
	  -x go.mod -x go.sum -x requirements.txt -x README.md \
	  -x AGENTS.md -x book.pdf -x book-review.pdf -x docs -x .git \
	  -x .DS_Store $$d . \
	  && echo ok
