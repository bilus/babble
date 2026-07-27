-- Code generated from BOOK.org by make tangle. DO NOT EDIT.

-- Review-weave styling for plan markup: badge boxes, colored state
-- keywords, gray tags. Runs after lit/preprocess.py has rewritten
-- inline tasks into planned/hole special blocks.

local function box(el, color)
  local blocks = pandoc.List({ pandoc.RawBlock('latex', '\\begin{planbox}{' .. color .. '}') })
  blocks:extend(el.content)
  blocks:insert(pandoc.RawBlock('latex', '\\end{planbox}'))
  return blocks
end

function Div(el)
  if el.classes:includes('planned') then return box(el, 'badgeplan') end
  if el.classes:includes('hole') then return box(el, 'badgehole') end
end

-- pandoc's org reader leaves custom #+TODO keywords as plain text, so
-- color them here when a heading starts with one.
local NOT_DONE = { PLANNED = true, HOLE = true, FILLING = true }
local DONE = { LIVE = true, DROPPED = true }

function Header(el)
  local first = el.content[1]
  if first and first.t == 'Str' then
    local color = (NOT_DONE[first.text] and 'cmdel') or (DONE[first.text] and 'cmins')
    if color then
      el.content[1] = pandoc.Span({
        pandoc.RawInline('latex', '\\texorpdfstring{\\textcolor{' .. color .. '}{'),
        pandoc.Str(first.text),
        pandoc.RawInline('latex', '}}{' .. first.text .. '}'),
      })
      return el
    end
  end
end

-- Paragraphs carrying CriticMarkup get a small mark in the margin so
-- a scrolling eye catches them without reading.
local function has_cm(inlines)
  for _, il in ipairs(inlines) do
    if il.t == 'RawInline' and il.format:match('tex')
       and il.text:match('\\cm') then
      return true
    end
  end
  return false
end

function Para(el)
  if has_cm(el.content) then
    el.content:insert(1, pandoc.RawInline('latex', '\\planmark{cmnote}{}'))
    return el
  end
end

local function tint(el, latex)
  local out = pandoc.List({ pandoc.RawInline('latex', latex) })
  out:extend(el.content)
  out:insert(pandoc.RawInline('latex', '}'))
  return out
end

function Span(el)
  if el.classes:includes('todo') then
    return tint(el, '\\textcolor{cmdel}{\\bfseries ')
  end
  if el.classes:includes('done') then
    return tint(el, '\\textcolor{cmins}{\\bfseries ')
  end
  if el.classes:includes('tag') then
    return tint(el, '{\\small\\color{gray}')
  end
end
