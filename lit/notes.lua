-- Code generated from BOOK.org by make tangle. DO NOT EDIT.

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

-- Pandoc's org reader keeps :tangle and :noweb as block attributes;
-- blocks that land in a file get a paper-style numbered caption
-- naming it, plus the block's <<name>> when it has one. A named
-- region block (no target of its own, included through a reference)
-- is captioned with the file its includer tangles to, resolved
-- transitively over the whole document. Twins and illustrations are
-- referenced by nothing and tangle nowhere, so they stay
-- uncaptioned.
--
-- Syntax highlighting shreds a <<ref>> into operator and identifier
-- tokens, so no literal <<ref>> survives into the .tex for a filter
-- to linkify. Instead, in :noweb yes blocks each ref is swapped for
-- an alphanumeric placeholder (hex-encoded name), which every
-- lexer keeps whole; texlinks.py decodes it back into a \NowebRef
-- hyperlink after pandoc writes the .tex.
local function hex(s)
  return (s:gsub('.', function(c) return string.format('%02x', c:byte()) end))
end

local REF = '<<([%w_][%w_.%-]*)>>'

function Pandoc(doc)
  -- first pass: each block's own target, and who references whom
  local blocks = {}
  doc:walk({ CodeBlock = function(el)
    local t = el.attributes['tangle']
    local entry = { id = el.identifier, names = {} }
    if t and t ~= 'no' then
      entry.target = t
    end
    if el.attributes['noweb'] == 'yes' then
      for name in el.text:gmatch(REF) do
        table.insert(entry.names, name)
      end
    end
    table.insert(blocks, entry)
  end })
  -- resolve region names to their includer's file, transitively
  local resolved = {}
  for _, e in ipairs(blocks) do
    if e.target and e.id ~= '' then
      resolved[e.id] = e.target
    end
  end
  for _ = 1, #blocks do
    local changed = false
    for _, e in ipairs(blocks) do
      local t = e.target or (e.id ~= '' and resolved[e.id])
      if t then
        for _, name in ipairs(e.names) do
          if not resolved[name] then
            resolved[name] = t
            changed = true
          end
        end
      end
    end
    if not changed then
      break
    end
  end
  -- second pass: captions, anchors, and reference placeholders
  return doc:walk({ CodeBlock = function(el)
    local out = el
    if el.attributes['noweb'] == 'yes' then
      local text = el.text:gsub(REF, function(name)
        return 'NOWEBREF' .. hex(name) .. 'FERBEWON'
      end)
      if text ~= el.text then
        out = el:clone()
        out.text = text
      end
    end
    local anchor = ''
    if el.identifier ~= '' then
      anchor = '\\hypertarget{noweb:' .. el.identifier .. '}{}'
    end
    local t = el.attributes['tangle']
    local file = (t and t ~= 'no') and t or nil
    if not file and el.identifier ~= '' then
      file = resolved[el.identifier]
    end
    if file then
      local caption
      if el.identifier ~= '' then
        caption = anchor .. '\\codecaption[' .. el.identifier .. ']{' .. file .. '}'
      else
        caption = '\\codecaption{' .. file .. '}'
      end
      return { pandoc.RawBlock('latex', caption), out }
    end
    if anchor ~= '' then
      return { pandoc.RawBlock('latex', anchor), out }
    end
    if out ~= el then
      return out
    end
  end })
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
