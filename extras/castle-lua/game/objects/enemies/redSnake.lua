RedSnake = Enemy:extend()
RedSnake.spritesheet = love.graphics.newImage("res/redSnake.png")

function RedSnake:new(...)
	--RedSnake.super.new(self, ...)
	self.updateMove = function() end
	Snake.new(self, ...)

	self.lastTurn = true
end

function RedSnake:update(dt)
	RedSnake.super.update(self, dt)

	self.currentAnim = self.anims[self.dir]
	self.currentAnim:update(dt)

	-- if abyss, then turn
	if self.ground then
		self.velx = self.speed * (self.dir == "right" and 1 or -1)

		local off = self.dir == "right" and self.w - 4 or 4
		local _, len = self.area.world:queryPoint(self.x + off, self.y + self.h + 1, self.filterSolid)

		if len <= 0 then
			if not self.lastTurn then
				self.lastTurn = true
				self.dir = self.dir == "right" and "left" or "right"
			end
		else self.lastTurn = false end
	end

	local cols = self:updateHitbox(dt, function(other)
		if other.is and other:is(Enemy) then return false end
	end)
	self:handleCollisions(cols)
end

function RedSnake:draw() Snake.draw(self) end

function RedSnake:handleCollisions(cols)
	for _, c in pairs(cols) do
		if self:checkSolid(c.other) and c.normal.x ~= 0 then
			self.dir = self.dir == "right" and "left" or "right"
			if not self.ground then self.velx = 0 end
			break
		end
	end
end

function RedSnake.filterSolid(obj)
	return GameObject.checkSolid(self, obj)
end
