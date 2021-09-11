Spear = GameObject:extend()
Spear.spritesheet = love.graphics.newImage("res/spear.png")

function Spear:new(...)
	Spear.super.new(self, ...)

	self.weapon = true
	self.damage = 1
	self.canDamage = false

	self.extraReach = 1

	self.w, self.h = 17 + self.extraReach, 5
	self.physics = false
	self.noclip = true
	self.offsetx = -17/2
	self.offsety = 0

	self.grid = anim8.newGrid(25, 9, self.spritesheet:getWidth(), self.spritesheet:getHeight())
	self.anim = anim8.newAnimation(self.grid("1-4", 1), {self.firstDelay or 0.3, 0.1, 0.2, 0.1})

	if self.dir == "right" then
		self.hitpadx = 0
		self.hitpady = 3
	else
		self.hitpadx = -11 - self.extraReach
		self.hitpady = 3
	end

	self.player.extraDraw = function(player)
		love.graphics.setColor(1, 1, 1, 1)
		self.anim:draw(self.spritesheet, player.canx + self.offsetx, player.cany + self.offsety)
	end
end

function Spear:update(dt)
	Spear.super.update(self, dt)
	self.anim:update(dt)
	self.x, self.y = self.player.x + self.hitpadx, self.player.y + self.hitpady

	local cols = self:updateHitbox(dt)
	if self.canDamage then
		if cols then self:handleCollisions(cols) end
	end
end

function Spear:nextState()
	self.canDamage = true
	-- update direction if changed
	if self.player.dir ~= self.dir then
		self.dir = self.player.dir
		if self.dir == "right" then
			self.hitpadx = 0
			self.hitpady = 3
		else
			self.hitpadx = -11 - self.extraReach
			self.hitpady = 3
		end
	end
	self:update(0)
	self:setHitbox(self.w, self.h)
end

function Spear:handleCollisions(cols)
	for _, c in pairs(cols) do
		local o = c.other
		if o ~= self.player then 
			if o.enemy or o.hit then
				c.normal.x = self.dir == "right" and -1 or 1
				o:hit(self.damage, c)
			end
		end
	end
end
