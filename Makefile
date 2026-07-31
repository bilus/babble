CHECK_EXTRA := -x PROMPT.md -x babble -x testdata -x .github \
  -x devbox.json -x devbox.lock -x .devbox -x Dockerfile -x .dockerignore
include lit/lit.mk
