#!/usr/bin/env python3
# Code generated from BOOK.org by make tangle. DO NOT EDIT.

"""Rewrite noweb placeholders in the woven .tex into hyperlinks.

notes.lua swaps each <<ref>> in a :noweb yes block for an
alphanumeric placeholder, the only spelling that survives syntax
highlighting in one piece. Here the placeholder turns back into
\\NowebRef{name} where LaTeX commands work (pandoc's Highlighting
environment) and into the literal <<name>> everywhere else, since
verbatim fallbacks take no commands.
"""
import re
import sys

PLACEHOLDER = re.compile(r"NOWEBREF([0-9a-f]+)FERBEWON")


def decode(match):
    return bytes.fromhex(match.group(1)).decode("utf-8")


def rewrite(lines):
    out = []
    highlighted = False
    for line in lines:
        if "\\begin{Highlighting}" in line:
            highlighted = True
        if "\\end{Highlighting}" in line:
            highlighted = False
        if highlighted:
            line = PLACEHOLDER.sub(lambda m: "\\NowebRef{%s}" % decode(m), line)
        else:
            line = PLACEHOLDER.sub(lambda m: "<<%s>>" % decode(m), line)
        out.append(line)
    return out


def main():
    path = sys.argv[1]
    with open(path, encoding="utf-8") as f:
        lines = f.readlines()
    with open(path, "w", encoding="utf-8", newline="\n") as f:
        f.writelines(rewrite(lines))


if __name__ == "__main__":
    main()
