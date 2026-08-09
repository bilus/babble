-- Code generated from BOOK.org by make tangle. DO NOT EDIT.

-- Render mermaid src blocks to PDF images at weave time via mmdc.
-- MERMAID_TMP names the scratch directory for generated files; the
-- weave recipe creates and removes it. MERMAID_PUPPETEER_CONFIG, when
-- set, names a puppeteer config file passed on to the browser mmdc
-- drives; a container needs one to turn the sandbox off. When mmdc is
-- missing or a render fails, the block stays as fontified source and
-- the weave still completes.
local dir = os.getenv("MERMAID_TMP") or "."
local puppeteer = os.getenv("MERMAID_PUPPETEER_CONFIG")
local n = 0

function CodeBlock(el)
  if not el.classes:includes("mermaid") then return nil end
  n = n + 1
  local base = string.format("%s/mermaid-%d", dir, n)
  local f = assert(io.open(base .. ".mmd", "w"))
  f:write(el.text)
  f:close()
  local args = { "-q", "-i", base .. ".mmd", "-o", base .. ".pdf",
                 "-f", "-b", "transparent" }
  if puppeteer and puppeteer ~= "" then
    table.insert(args, "-p")
    table.insert(args, puppeteer)
  end
  local ok = pcall(pandoc.pipe, "mmdc", args, "")
  if not ok then
    io.stderr:write("mermaid.lua: mmdc failed; keeping source block\n")
    return nil
  end
  return pandoc.Para({ pandoc.Image({}, base .. ".pdf", "",
                                    pandoc.Attr(el.identifier)) })
end
