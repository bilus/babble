CHECK_EXTRA := -x PROMPT.md -x babble -x testdata -x .github \
  -x devbox.json -x devbox.lock -x .devbox -x Dockerfile -x .dockerignore -x puppeteer.json -x .venv -x 'book-*.pdf'
include lit/lit.mk

# Emacs refreshes a dynamic block by searching raw text, wherever the
# line sits, so an unescaped opener quoted inside a code body is one
# it rewrites in place. Without this the tangler would damage the
# book rather than refuse it. The check hangs off the recipe that
# runs Emacs rather than off the tangle target, because a
# prerequisite of the target is ordered only when make runs serially.
# It needs a babble built from the book it is checking, so on a fresh
# checkout there is none and it stands aside.
tangle_el: guard_book

guard_book:
	@if [ -x ./babble ]; then ./babble parse BOOK.org >/dev/null; \
	else echo "no ./babble yet, tangling unguarded"; fi

.PHONY: guard_book
