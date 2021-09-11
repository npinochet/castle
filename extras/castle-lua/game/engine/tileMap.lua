TileMap = Object:extend()

-- tileset local id
TileMap.entityBind = {
	[100] = "Player",
	[101] = "Torch",
	[102] = "Sign",
	[103] = "Spike",
	[104] = "Anchor",
	[105] = "Chest",
	[106] = "Coin",
	[107] = "Gem",

	[150] = "Snake",
	[151] = "RedSnake",
	[152] = "Ghoul",
	--[153] = "Fred",
	[154] = "Worm",

	[200] = "BreakTile",
	[201] = "JumpThrough",
	[202] = "BreakAllTile",
}

function TileMap:new(map_file_path)
	self.area = Area(self)
	self.area:addPhysicsWorld()

	self.map = sti(map_file_path, {"bump"})
	self.map:bump_init(self.area.world)

	self.timer = Timer()
	self.parallax = {}


	-- set the correct entity depending on object tile gid and entityBind

	-- get firstgid to change from local id to gid on object tiles
	local firstgid = 1
	for i, t in pairs(self.map.tilesets) do
		if t.name == "entities" then firstgid = t.firstgid break end
	end

	-- also correct the way tiled save the y position (y - h) on objects for some reason
	for i, o in pairs(self.map.layers["entity"].objects) do
		if o.gid then o.type = self.entityBind[o.gid - firstgid] end
		o.y = o.y - o.height
	end
end

function TileMap:update(dt)
	self.timer:update(dt)
	self.area:update(dt)
	self.map:update(dt)
end

function TileMap:draw(background)
	love.graphics.setColor(1, 1, 1, 1)
	for i, v in pairs(self.parallax) do
		love.graphics.draw(v.img, v.x - v.w, v.y)
		love.graphics.draw(v.img, v.x, v.y)
	end
	camera:attach(nil, nil, wx, wy)
		self.map.layers["background"].draw()
		if background then background() end
		self.area:draw()
		love.graphics.setColor(1, 1, 1, 1)
		self.map.layers["foreground"].draw()
		if debug then
			love.graphics.setColor(1,0,0,0.6)
			self.map:bump_draw(self.area.world)
			love.graphics.setColor(1,1,1, 1) -- ???
		end
	camera:detach()
end

function TileMap:destroy()
	for index, _ in ipairs(self.map.layers) do self.map:bump_removeLayer(index, self.area.world) end
	for i, _ in pairs(self.parallax) do self.parallax[i] = nil end
	self.area:destroy()
	self.timer:destroy()
end

function TileMap:addParallaxBackground(img, speed, vertical)
	local p = {}
	p.img = img
	p.vertical = vertical or false
	p.x, p.y = 0, 0
	p.w, p.h = img:getWidth(), img:getHeight()
	p.speed = speed or 0.5
	table.insert(self.parallax, p)
end

function TileMap:updateParallax(dt)
	for i, p in pairs(self.parallax) do
		local camx, camy = camera.x - wx/2, camera.y - wy/2
		local offsetx, offsety = (-camx * p.speed) % p.w, (-camy * p.speed) % p.h
		p.x = p.vertical and 0 or offsetx
		p.y = p.vertical and offsety or 0
	end
end

function TileMap:getCollidingObjects(layer, obj)
	local layer = self.map.layers[layer]
	local cols = {}

	if layer.colliding then
		if shortaabb(layer.colliding, obj) then table.insert(cols, layer.colliding) end
		for i, r in pairs(layer.objects) do
			if r ~= layer.colliding and shortaabb(obj, r) then
				table.insert(cols, r)
				break
			end
		end
	else
		for i, r in pairs(layer.objects) do
			if shortaabb(obj, r) then
				table.insert(cols, r)
				break
			end
		end
	end
	if #cols == 1 then layer.colliding = cols[1] end
	return cols
end

function TileMap:loadRoom(room)
	local room = room or self:getCollidingObjects("room", self.player)[1]
	self:spawnEntities("entity", room)
end

function TileMap:spawnEntities(layer, box)
	local layer = self.map.layers[layer]
	for _, obj in pairs(layer.objects) do
		if (not box) or shortaabb(box, obj) then
			if obj.type ~= "" then
				local ent = self.area:addGameObject(obj.type, obj.x, obj.y, {tiledObj = obj})

				-- correct entity size with tiled position
				-- (ground is the correct way to put entites, even if they are very tall)
				if self.area.world:hasItem(ent) then
					local _, _, _, h = self.area.world:getRect(ent)
					ent.y = ent.y + obj.height - h
					ent:updateHitbox(0, function() return false end)
				end
			end
		end
	end
end

function TileMap:getEntityObject(ent, layer)
	local layer = self.map.layers[layer]
	for i, obj in pairs(layer.objects) do
		if obj.type == ent then return obj, i end
	end
	return false
end
