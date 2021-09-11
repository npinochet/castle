BreakTile = GameObject:extend()

function BreakTile:new(...)
	BreakTile.super.new(self, ...)

	self.solid = true
	self.dying = false
	self.circle = false
	self.w, self.h = self.tiledObj.width, self.tiledObj.width
	self:setHitbox(self.w, self.h)

	local map = self.area.scene.map

	-- get tiles and instance
	self.tiles = {}
	for y = 1, math.floor(self.w/map.tilewidth) do
		for x = 1, math.floor(self.h/map.tileheight) do
			local entry = {}
			local px, py = self.x + (x-1)*map.tilewidth, self.y + (y-1)*map.tileheight
			local tx, ty = map:convertPixelToTile(px, py)
			local tile = map.layers["foreground"].data[ty+1][tx+1]
			tile = tile or map.layers["background"].data[ty+1][tx+1]
			table.insert(entry, tile)

			for _, t in pairs(map.tileInstances[tile.gid]) do
				if t.x == px and t.y == py then
					table.insert(entry, t)
					break
				end
			end
			table.insert(self.tiles, entry)
		end
	end
end

function BreakTile:update(dt)
	BreakTile.super.update(self, dt)
	if self.circle then love.graphics.circle("fill", self.x + self.w/2, self.y + self.h/2, self.w) end
end

function BreakTile:hit(damage, col)
	if not self.dying then
		self:deadEffect(col)

		local map = self.area.scene.map

		 -- remove tiles
		for _, t in pairs(self.tiles) do if t[2] then map:swapTile(t[2], map.tiles[1]) end end
	end
end

function BreakTile:deadEffect(col)
	camera:shake(0.2, 2)
	self.dying = true
	self.circle = true
	self.area.world:remove(self) -- remove hitbox

	-- spawn smoke
	local speed = 10
	local s = {
		img = love.graphics.newImage("res/smoke.png"),
		anim = {7, 7},
		duration = 1,
		color = {1, 1, 1, 180/255},
	}
	for i = 1, love.math.random(3, 5)*(#self.tiles) do
		s.velx = love.math.random(-speed, speed)
		s.vely = love.math.random(-speed, speed)
		local x = self.x + self.w/2 + love.math.random(-self.w/2, self.w/2)
		local y = self.y + self.h/2 + love.math.random(-self.h/2, self.h/2)
		self.area:addGameObject("Particle", x, y, s)
	end

	self.timer:after(0.05, function() self.circle = false end)
end


function BreakTile:destroy()
	local map = self.area.scene.map

	-- restore tiles
	for _, t in pairs(self.tiles) do if t[2] then map:swapTile(t[2], t[1]) end end
	for _, t in pairs(self.tiles) do t = nil end
	self.tiles = nil
	BreakTile.super.destroy(self)
end
