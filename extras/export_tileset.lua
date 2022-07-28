if TilesetMode == nil then return app.alert "Use Aseprite v1.3" end

local lay = app.activeLayer
if not lay.isTilemap then return app.alert "No active tilemap layer" end

local tileset = lay.tileset

local dlg = Dialog("Export Tileset")
dlg:file{ id="file", label="Export to File:", save=true, focus=true,
          filename=app.fs.joinPath(app.fs.filePath(lay.sprite.filename), lay.name .. ".png") }
   :number{ id = "width", label = "Width:", text = '16' }
   :label{ label="# Tiles", text=tostring(#tileset) }
   :separator()
   :button{ text="&Export", focus=true, id="ok" }
   :button{ text="&Cancel" }
   :show()

local data = dlg.data
if data.ok then
  local spec = lay.sprite.spec
  local grid = tileset.grid
  local size = grid.tileSize
  spec.width = size.width * data.width
  spec.height = size.height * math.ceil(#tileset / data.width)
  local image = Image(spec)
  image:clear()
  for i = 0, #tileset-1 do
    local tile = tileset:getTile(i)
    local w, h = i % data.width, math.floor(i / data.width)
    image:drawImage(tile, w*size.width, h*size.height)
  end
  image:saveAs(data.file)

  -- Open exported file?
  --app.open(data.file)
end
