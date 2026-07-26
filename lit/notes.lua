-- CriticMarkup pluses in a plain weave parse as org strikeout, and
-- soul's \st cannot typeset what lands inside ("Reconstruction
-- failed"); render the span raw instead, pluses restored. The review
-- weave never gets here: preprocess.py rewrites the marks first.
function Strikeout(el)
  local out = pandoc.List({ pandoc.Str('+') })
  out:extend(el.content)
  out:insert(pandoc.Str('+'))
  return out
end

-- Org links to #+name:'d blocks work in the buffer, but pandoc's org
-- reader leaves them as "spurious-link" spans; the named block itself
-- does get a \label. Repair the spans into real internal links.
function Span(el)
  if el.classes:includes('spurious-link') and el.attributes['target'] then
    local name = pandoc.utils.stringify(el.content)
    return pandoc.Link({ pandoc.Code(name) }, '#' .. el.attributes['target'])
  end
end

-- Style "Note:" / "Apropos:" paragraphs as indented asides with a bold
-- lead, so promoted source comments stand apart from the narrative.
function Para(p)
  local first = p.content[1]
  if first and first.t == 'Str' and (first.text == 'Note:' or first.text == 'Apropos:') then
    local sigil = first.text:sub(1, -2)
    table.remove(p.content, 1)
    if p.content[1] and p.content[1].t == 'Space' then
      table.remove(p.content, 1)
    end
    table.insert(p.content, 1, pandoc.Space())
    table.insert(p.content, 1, pandoc.Strong{pandoc.Str(sigil .. '.')})
    return pandoc.BlockQuote{pandoc.Para(p.content)}
  end
end
