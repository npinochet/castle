Ghoul = Enemy:extend()
Ghoul.spritesheet = love.graphics.newImage("res/ghoul.png")

function Ghoul:new(...)
	Ghoul.super.new(self, ...)

	self.speed = 35
	self.damage = 1
	self.hp = 1
	self.value = {0, 2}

	local padx, pady = 1, 3
	self.w, self.h = 10 - padx*2, 16 - pady
	self:setHitbox(self.w, self.h)

	self.grid = anim8.newGrid(10, 16, self.spritesheet:getWidth(), self.spritesheet:getHeight())
	self.offsetx = padx
	self.offsety = pady

	self.dir = self.dir or "left"

	local duration = 0.5
	self.anims = {
		left = anim8.newAnimation(self.grid("1-2", 1), duration),
		right = anim8.newAnimation(self.grid("1-2", 1), duration):flipH(),
	}
	self.currentAnim = self.anims[self.dir]

	self.lastTurn = true
end

function Ghoul:update(dt)
	Ghoul.super.update(self, dt)

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

function Ghoul:draw()
	if not self.dying then love.graphics.setColor(1, 1, 1, 1) end
	if self.flash then love.graphics.setColor(220/255, 40/255, 40/255) end
	self.currentAnim:draw(self.spritesheet, self.x, self.y, 0, 1, 1, self.offsetx, self.offsety)
	Ghoul.super.draw(self)
end

function Ghoul:handleCollisions(cols)
	for _, c in pairs(cols) do
		if self:checkSolid(c.other) and c.normal.x ~= 0 then
			self.dir = self.dir == "right" and "left" or "right"
			if not self.ground then self.velx = 0 end
			break
		end
	end
end

function Ghoul.filterSolid(obj)
	return GameObject.checkSolid(self, obj)
end
