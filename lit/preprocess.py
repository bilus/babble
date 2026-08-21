#!/usr/bin/env python3
# Code generated from BOOK.org by make tangle. DO NOT EDIT.

"""Review-weave preprocessor: org with plan markup -> pandoc-friendly org.

- inline tasks (15+ stars) -> #+begin_planned / #+begin_hole special blocks
- CriticMarkup {++new++} {--old--} {~~old~>new~~} {>>note<<} -> LaTeX
- #+begin:/#+end: dynamic-block wrapper lines dropped (inner content kept)
- LaTeX preamble injected after #+options: (geometry only if absent)
"""
import bisect
import os
import re
import sys

KEYWORDS = {"PLANNED": "planned", "FILLING": "planned", "LIVE": "planned",
            "DROPPED": "planned", "HOLE": "planned"}

PREAMBLE = [
    "#+latex_header: \\usepackage[margin=2.6cm]{geometry}",
    "#+latex_header: \\usepackage{xcolor}",
    "#+latex_header: \\usepackage[normalem]{ulem}",
    "#+latex_header: \\definecolor{cmdel}{RGB}{176,32,32}",
    "#+latex_header: \\definecolor{cmins}{RGB}{22,122,48}",
    "#+latex_header: \\definecolor{cmnote}{RGB}{28,82,160}",
    "#+latex_header: \\definecolor{badgeplan}{RGB}{176,98,20}",
    "#+latex_header: \\definecolor{badgehole}{RGB}{104,54,160}",
    "#+latex_header: \\newcommand{\\cmdel}[1]{\\textcolor{cmdel}{\\sout{#1}}}",
    "#+latex_header: \\newcommand{\\cmins}[1]{\\textcolor{cmins}{#1}}",
    "#+latex_header: \\newcommand{\\cmnote}[1]{\\textcolor{cmnote}{\\itshape [#1]}}",
    "#+latex_header: \\usepackage{framed}",
    "#+latex_header: \\usepackage{marginnote}",
    "#+latex_header: \\reversemarginpar",
    "#+latex_header: \\usepackage{bookmark}",
    # inside framed environments, color must be explicitly paired
    # (begingroup/endgroup or textcolor): a bare \color relying on the
    # environment group leaks past the box under framed's splitting
    "#+latex_header: \\newcommand{\\planmark}[2]{\\marginnote{\\textcolor{#1}{\\rule{2.5pt}{9pt}\\,{\\scriptsize #2}}}}",
    # badge variant: same margin mark plus a sticky-note annotation
    # (xdvipdfmx pdf:ann special), so PDF viewers list the badge's why
    # under Highlights and Notes
    "#+latex_header: \\newcommand{\\planbadgemark}[3]{\\marginnote{\\special{pdf:ann width 8pt height 8pt << /Type /Annot /Subtype /Text /Name /Comment /C [0.69 0.38 0.08] /Contents (#3) >>}\\hspace{14pt}\\textcolor{#1}{\\rule{2.5pt}{9pt}\\,{\\scriptsize #2}}}}",
    "#+latex_header: \\newenvironment{planbox}[1]{\\def\\FrameCommand{{\\color{#1}\\vrule width 3pt}\\hspace{8pt}}\\MakeFramed{\\advance\\hsize-\\width\\FrameRestore}\\begingroup\\color{#1}\\itshape}{\\endgroup\\endMakeFramed}",
    # block quotes, including the Note/Apropos asides notes.lua turns
    # into them, render as sidebars: a gray left rule through the same
    # framed machinery as planbox
    "#+latex_header: \\definecolor{quotebar}{RGB}{160,160,160}",
    "#+latex_header: \\renewenvironment{quote}{\\def\\FrameCommand{{\\color{quotebar}\\vrule width 2pt}\\hspace{10pt}}\\MakeFramed{\\advance\\hsize-\\width\\FrameRestore}}{\\endMakeFramed}",
    # numbered caption above every tangling block (notes.lua emits the
    # calls, adding [name] when the block has one); detokenize keeps
    # underscores in filenames and block names typesettable
    "#+latex_header: \\newcounter{codelisting}",
    "#+latex_header: \\newcommand{\\codecaption}[2][]{\\par\\medskip\\noindent{\\stepcounter{codelisting}\\small\\textbf{Listing \\thecodelisting:} \\texttt{\\detokenize{#2}}\\if\\relax\\detokenize{#1}\\relax\\else\\quad\\textcolor{gray}{\\texttt{\\detokenize{<<#1>>}}}\\fi}\\par\\nopagebreak\\vspace{2pt}}",
    # noweb references inside highlighted code: texlinks.py rewrites
    # the placeholders notes.lua planted into \NowebRef calls. Plain
    # \hyperlink, not \hyperref[label]: label lookup explodes under
    # fancyvrb's verbatim catcodes, so notes.lua plants matching
    # \hypertarget anchors at the named blocks instead. The dest name
    # is detokenized too: fancyvrb makes - active, and an active -
    # expands to hyphenation boxes inside the destination string
    "#+latex_header: \\newcommand{\\NowebRef}[1]{\\hyperlink{noweb:\\detokenize{#1}}{\\detokenize{<<#1>>}}}",
]

# The content between the delimiters stays org: pandoc renders its
# markup (verbatim, links, braces) and only the \cm..{ } shells are
# raw. Flattening the content into the raw span instead breaks on
# inner braces and turns org markup into literal TeX.
def cm(macro, s):
    return "@@latex:\\%s{@@%s@@latex:}@@" % (macro, s)

def criticmarkup(text):
    text = re.sub(r"\{~~(.+?)~>(.+?)~~\}",
                  lambda m: cm("cmdel", m.group(1)) + " " + cm("cmins", m.group(2)),
                  text, flags=re.S)
    text = re.sub(r"\{\+\+(.+?)\+\+\}",
                  lambda m: cm("cmins", m.group(1)),
                  text, flags=re.S)
    text = re.sub(r"\{--(.+?)--\}",
                  lambda m: cm("cmdel", m.group(1)),
                  text, flags=re.S)
    text = re.sub(r"\{>>(.+?)<<\}",
                  lambda m: cm("cmnote", m.group(1)),
                  text, flags=re.S)
    return text

def strip_org(s):
    s = re.sub(r"\[\[[^]]*\]\[([^]]*)\]\]", r"\1", s)
    s = re.sub(r"\[\[([^]]*)\]\]", r"\1", s)
    return s.strip()

def badge_titles(text):
    """Badge display titles, in the same order badges() counts them."""
    star = re.compile(r"^\*{15,}\s+(.*\S)\s*$")
    out = []
    for l in text.split("\n"):
        m = star.match(l)
        if not m or m.group(1).strip() == "END":
            continue
        t = m.group(1)
        tm = re.match(r"^(.*?)\s+(:[A-Za-z0-9_@:]+:)$", t)
        if tm:
            t = tm.group(1)
        out.append(strip_org(t))
    return out

def pdf_note(title, body):
    """Sticky-note contents: PDF-string- and macro-arg-safe ASCII.

    Balanced parens are legal unescaped in a PDF literal string, and
    backslash escapes would be undefined control sequences inside the
    special, so parens stay only when balanced in order."""
    s = title
    if body:
        s += ": " + " ".join(l.strip() for l in body if l.strip())
    s = strip_org(s)
    s = re.sub(r"[{}~%$&^_\\#]", " ", s)
    s = " ".join(s.split())
    depth = 0
    for c in s:
        if c == "(":
            depth += 1
        elif c == ")":
            depth -= 1
            if depth < 0:
                break
    if depth != 0:
        s = " ".join(s.replace("(", " ").replace(")", " ").split())
    return s[:237] + "..." if len(s) > 240 else s

def badges(lines):
    star = re.compile(r"^\*{15,}\s+(.*\S)\s*$")
    head = re.compile(r"^\*+\s")
    out, i, k = [], 0, 0
    while i < len(lines):
        m = star.match(lines[i])
        if not m:
            out.append(lines[i])
            i += 1
            continue
        title = m.group(1)
        body = None
        j = i + 1
        while j < len(lines):
            if head.match(lines[j]):
                mm = star.match(lines[j])
                if mm and mm.group(1).strip() == "END":
                    body = lines[i + 1:j]
                    i = j
                break
            j += 1
        tags = ""
        tm = re.match(r"^(.*?)\s+(:[A-Za-z0-9_@:]+:)$", title)
        if tm:
            title, tags = tm.group(1), tm.group(2)
        kw = None
        parts = title.split(None, 1)
        if parts and parts[0] in KEYWORDS:
            kw = parts[0]
            title = parts[1] if len(parts) > 1 else ""
        block = KEYWORDS.get(kw, "planned")
        label = ("*%s* \u2014 " % kw if kw else "") + title
        if tags:
            label += "  (" + tags.strip(":").replace(":", ", ") + ")"
        stage = re.search(r"stage(\d+)", tags)
        # inline in the label line: the margin mark aligns with the
        # box's first line, the hypertarget (ordinal k, matching the
        # badge order the plan-toc rewrite counts) receives the index
        # links, and the sticky note carries the badge's why
        k += 1
        mark = ("@@latex:\\hypertarget{planbadge-%d}{}"
                "\\planbadgemark{badgeplan}{%s}{%s}@@ ") % (
            k, "s" + stage.group(1) if stage else "",
            pdf_note(strip_org(title), body))
        out.append("#+begin_" + block)
        out.append(mark + label)
        if body:
            out.append("")
            out.extend(body)
        out.append("#+end_" + block)
        i += 1
    return out

INCLUDE = re.compile(r'^[ \t]*#\+INCLUDE:(.*)$', re.IGNORECASE)
BARE = re.compile(r'^"([^"]+)"$')
WRAPPED = re.compile(r'^"([^"]+)"\s+example$')
CUSTOM_ID = re.compile(r'^[ \t]*:CUSTOM_ID:[ \t]*(\S+)', re.IGNORECASE)


def read_book(src, seen=None, ids=None, base=None):
    """The master and every chapter it includes, as one text.

    Only the bare quoted form is understood, the same one the tangler
    accepts, so both halves of the toolchain read the same book. A
    custom id defined twice is refused here rather than left to pandoc:
    this function still knows which file and line the second definition
    sits on, and after expansion nobody does.
    """
    seen = seen if seen is not None else set()
    ids = ids if ids is not None else {}
    path = os.path.abspath(src)
    base = base if base is not None else os.path.dirname(path)
    # positions are reported the way the includes are written, relative
    # to the master, since that is how the reader will go looking
    here = os.path.relpath(path, base)
    if path in seen:
        raise SystemExit("%s: includes itself" % here)
    seen.add(path)
    out = []
    for n, line in enumerate(open(path, encoding="utf-8").read().split("\n"), 1):
        m = INCLUDE.match(line)
        if m:
            rest = m.group(1).strip()
            if WRAPPED.match(rest):
                out.append(line)      # verbatim display; pandoc expands it
                continue
            arg = BARE.match(rest)
            if not arg:
                raise SystemExit(
                    '%s:%d: only #+INCLUDE: "path" and #+INCLUDE: "path" '
                    "example are understood" % (here, n))
            out.append(read_book(
                os.path.join(os.path.dirname(path), arg.group(1)),
                seen, ids, base))
            continue
        cid = CUSTOM_ID.match(line)
        if cid:
            key = cid.group(1)
            if key in ids:
                raise SystemExit(
                    "%s:%d: custom id %s is already defined at %s"
                    % (here, n, key, ids[key]))
            ids[key] = "%s:%d" % (here, n)
        out.append(line)
    return "\n".join(out)


def main(src, dst):
    text = read_book(src)
    text = criticmarkup(text)
    text = text.replace("\u2192", "$\\rightarrow$")
    has_geometry = re.search(r"^#\+latex_header:.*geometry", text,
                             re.IGNORECASE | re.MULTILINE)
    preamble = [p for p in PREAMBLE
                if "geometry" not in p or not has_geometry]
    # badge ordinals by original line number: the jump list's targets
    badge_lines = [i + 1 for i, l in enumerate(text.split("\n"))
                   if re.match(r"^\*{15,}\s+(?!END\b)", l)]
    # a "Plan sites" branch in the PDF outline, one entry per badge;
    # level 1 keeps the book's chapters as siblings, not children
    marks = []
    titles = badge_titles(text)
    if titles:
        marks.append("#+latex: \\bookmark[level=1,dest=planbadge-1]"
                     "{Plan sites}")
        for j, t in enumerate(titles, 1):
            marks.append("#+latex: \\bookmark[level=2,dest=planbadge-%d]"
                         "{%s}" % (j, re.sub(r"[{}\\%$&#^_~]", " ", t)))
    lines = []
    in_toc = False
    injected = False
    for line in text.split("\n"):
        if re.match(r"^#\+begin: plan-toc", line):
            in_toc = True
            continue
        if re.match(r"^#\+(begin|end):", line):
            if line.startswith("#+end:"):
                in_toc = False
            continue
        if in_toc:
            # file links are external in a PDF; link each item to its
            # governing badge instead, the nearest one above the site
            m = re.match(r"^- \[\[file:[^]]*::(\d+)\]\[(\d+)\]\] (.*)$", line)
            if m and bisect.bisect_right(badge_lines, int(m.group(1))):
                k = bisect.bisect_right(badge_lines, int(m.group(1)))
                line = "- @@latex:\\hyperlink{planbadge-%d}{%s}@@ %s" % (
                    k, m.group(2), m.group(3))
            else:
                line = re.sub(r"\[\[file:[^]]*\]\[([^]]*)\]\]", r"\1", line)
        # inject after #+options: when the book has one, else before
        # the first content line, so headerless books weave too
        if (not injected and line.strip()
                and not line.startswith("#+")):
            lines.extend(preamble)
            lines.extend(marks)
            injected = True
        lines.append(line)
        if not injected and line.startswith("#+options:"):
            lines.extend(preamble)
            lines.extend(marks)
            injected = True
    open(dst, "w", encoding="utf-8", newline="\n").write(
        "\n".join(badges(lines)) + "\n")

def deps(src, seen=None, base=None):
    """Every file the book is made of, the master included.

    Reported for make, which needs to know when the PDF is stale.
    Both include forms count: a chapter changes the book, and a file
    shown as an example changes the page it appears on.
    """
    seen = seen if seen is not None else []
    path = os.path.abspath(src)
    base = base if base is not None else os.path.dirname(path)
    rel = os.path.relpath(path, base)
    if rel in seen:
        return seen
    seen.append(rel)
    try:
        text = open(path, encoding="utf-8").read()
    except OSError:
        return seen
    for line in text.split("\n"):
        m = INCLUDE.match(line)
        if not m:
            continue
        arg = BARE.match(m.group(1).strip()) or WRAPPED.match(m.group(1).strip())
        if arg:
            deps(os.path.join(os.path.dirname(path), arg.group(1)), seen, base)
    return seen


if __name__ == "__main__":
    # --deps runs while make is still reading the Makefile, before a
    # target exists, so it reports what it can and never fails. A bad
    # include is refused later, by the tangler and by the weave, where
    # there is somewhere sensible to report it.
    if sys.argv[1] == "--deps":
        print("\n".join(deps(sys.argv[2])))
    else:
        main(sys.argv[1], sys.argv[2])
