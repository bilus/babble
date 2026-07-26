# Code generated from BOOK.org by make tangle. DO NOT EDIT.

# slugify keeps the alphanumeric runs and hyphenates the gaps; a run
# of junk is one gap, not several, and the ends come out clean.
def slugify(title):
    kept = "".join(c if c.isalnum() else " " for c in title.lower())
    return "-".join(kept.split())

# truncate cuts at the last hyphen inside the limit, never mid-word;
# a single run longer than the limit takes the hard cut.
def truncate(slug, limit):
    if len(slug) <= limit:
        return slug
    cut = slug.rfind("-", 0, limit + 1)
    return slug[:cut] if cut > 0 else slug[:limit]

# slug is the whole pipeline: slugified, then cut at the limit.
def slug(title, limit=64):
    """
    >>> slug("Quiet Brown Fox, part 2", limit=15)
    'quiet-brown-fox'
    """
    return truncate(slugify(title), limit)

# deduplicate returns the first of base, base-2, base-3, ... absent
# from taken; the bare slug always wins when free.
def deduplicate(base, taken):
    raise NotImplementedError("HOLE(2): first of base, base-2, ... not in taken")
