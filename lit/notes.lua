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

-- A link whose target names a figure (#+name: fig:...) becomes a
-- numbered cross-reference. Pandoc's org reader turns [[fig:x]] into
-- an autolink, which the LaTeX writer would set as a \url; the label
-- comes from the #+name: on the figure itself.
function Link(el)
  if el.target:match('^fig:') then
    return pandoc.RawInline('latex', 'Figure~\\ref{' .. el.target .. '}')
  end
  -- A stored link to a heading in another chapter names the file and
  -- the custom id. Includes make the chapters one document, so by the
  -- time this runs the file names nothing; the id resolves.
  local id = el.target:match('::#(.+)$')
  if id then
    el.target = '#' .. id
    return el
  end
end

-- Pandoc's org reader wraps a captioned block in a captioned-content
-- division rather than a figure, so a mermaid diagram with a
-- #+caption: would typeset as a bare image with its caption above.
-- Promote the division to a figure, which floats and takes a number.
-- The label comes from the #+name:, which mermaid.lua leaves on the
-- image it returns.
function Div(el)
  if not el.classes:includes('captioned-content') then
    return nil
  end
  local caption, content = nil, pandoc.List({})
  for _, b in ipairs(el.content) do
    if b.t == 'Div' and b.classes:includes('caption') then
      caption = b.content
    else
      content:insert(b)
    end
  end
  if not caption or #content ~= 1 then
    return nil
  end
  local body = content[1]
  if body.t ~= 'Para' and body.t ~= 'Plain' then
    return nil
  end
  if #body.content ~= 1 or body.content[1].t ~= 'Image' then
    return nil
  end
  local img = body.content[1]
  local id = img.identifier ~= '' and img.identifier or el.identifier
  img.identifier = ''
  return pandoc.Figure({ pandoc.Plain({ img }) },
                       pandoc.Caption(caption), pandoc.Attr(id))
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
  local bodies = {}
  -- first pass: each block's own target, and who references whom
  local blocks = {}
  doc:walk({ CodeBlock = function(el)
    local t = el.attributes['tangle']
    local entry = { id = el.identifier, names = {} }
    if t and t ~= 'no' then
      entry.target = t
    end
    if expandsNoweb(el) then
      for name in el.text:gmatch(REF) do
        table.insert(entry.names, name)
      end
    end
    if el.identifier ~= '' then
      bodies[el.identifier] = el
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
    if nowebWord(el, 'inline') then
      local text = inlineRefs(el.text, bodies, {})
      if text ~= el.text then
        out = el:clone()
        out.text = text
      end
    elseif expandsNoweb(el) then
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

-- The noweb header is a word list. Org splits it and looks for a
-- word it knows, so "yes inline" tangles as "yes" and carries the
-- extra word past every reader but this one.
function nowebWord(el, word)
  for w in (el.attributes['noweb'] or ''):gmatch('%S+') do
    if w == word then return true end
  end
  return false
end

function expandsNoweb(el)
  return nowebWord(el, 'yes')
end

-- What tangling would produce: the named body in place of the
-- reference, with the text before it repeated in front of every line
-- after the first. The body's own trailing newline goes first, the
-- way tangling drops it, or the prefix would be repeated once more
-- onto a line with nothing after it. A name nothing carries is left
-- as it stands, which is how a group reference keeps its link.
function inlineRefs(text, bodies, open)
  return (text:gsub('([^\n]-)(<<([%w_][%w_.%-]*)>>)', function(prefix, whole, name)
    local block = bodies[name]
    if not block or open[name] then return prefix .. whole end
    open[name] = true
    local body = block.text:gsub('\n$', '')
    if expandsNoweb(block) then
      body = inlineRefs(body, bodies, open)
    end
    open[name] = nil
    return prefix .. body:gsub('\n', '\n' .. prefix)
  end))
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
