-- Code generated from BOOK.org by make tangle. DO NOT EDIT.

-- Convert SVG figures to PDF at weave time, since LaTeX cannot
-- include SVG. MERMAID_TMP names the scratch directory the weave
-- recipe creates and removes.
local dir = os.getenv("MERMAID_TMP") or "."
local n = 0

function Image(el)
  if not el.src:match("%.svg$") then return nil end
  n = n + 1
  local out = string.format("%s/svg-%d.pdf", dir, n)
  local ok = pcall(pandoc.pipe, "rsvg-convert",
    { "-f", "pdf", "-o", out, el.src }, "")
  if not ok then
    io.stderr:write("svg.lua: rsvg-convert failed on " .. el.src .. "\n")
    return pandoc.Emph({ pandoc.Str("[figure " .. el.src .. " could not be converted]") })
  end
  el.src = out
  return el
end
