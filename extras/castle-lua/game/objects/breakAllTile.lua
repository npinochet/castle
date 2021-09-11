require("objects/breakTile")
BreakAllTile = BreakTile:extend()

function BreakAllTile:hit(damage, col)
	if not self.dying then
		self:deadEffect(col)

		local map = self.area.scene.map

		 -- remove tiles
		for _, t in pairs(self.tiles) do if t[2] then map:swapTile(t[2], map.tiles[1]) end end

		-- break adjasent breaktiles
		local adj = {}
		table.insert(adj, self:checkPoint(self.x + self.w + 1, self.y + self.h/2)) -- right
		table.insert(adj, self:checkPoint(self.x - 1, self.y + self.h/2)) -- left
		table.insert(adj, self:checkPoint(self.x + self.w/2, self.y - 1)) -- up
		table.insert(adj, self:checkPoint(self.x + self.w/2, self.y + self.h + 1)) -- down

		self.timer:after(0.1, function()
			for i, c in pairs(adj) do if c then c:hit(damage, col) end end
		end)
	end
end

function BreakAllTile:deadEffect(col)
	camera:shake(0.1, 1)
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
	s.velx = love.math.random(-speed, speed)
	s.vely = love.math.random(-speed, speed)
	local x = self.x + self.w/2 + love.math.random(-self.w/2, self.w/2)
	local y = self.y + self.h/2 + love.math.random(-self.h/2, self.h/2)
	self.area:addGameObject("Particle", x, y, s)

	self.timer:after(0.05, function() self.circle = false end)
end

function BreakAllTile:checkPoint(x, y)
	local w = self.area.world
	local cols, len = w:queryPoint(x, y, self.filterSolid)

	for _, c in pairs(cols) do if c.is and c:is(BreakAllTile) then return c end end
	return false
end

function BreakAllTile.filterSolid(obj)
	return GameObject.checkSolid(self, obj)
end
