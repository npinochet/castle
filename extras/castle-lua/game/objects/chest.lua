Chest = GameObject:extend()
Chest.img = love.graphics.newImage("res/chest.png")

function Chest:new(...)
	Chest.super().new(self, ...)

	self.value = {5, 10}
	self.noclip = true
	self.physics = false

	self.grid = anim8.newGrid(11, 7, self.img:getWidth(), self.img:getHeight())
	self.quads = self.grid("1-2", 1)

	self.hitPad = 2
	self.w, self.h = 11 + self.hitPad*2, 7 + self.hitPad
	self:setHitbox(self.w, self.h)

	self.x = self.x - self.hitPad
	--self.y = self.y - self.hitPad

	self.open = false
	self.alpha = 1
end

function Chest:update(dt)
	Chest.super().update(self, dt)
	self:updateHitbox(dt)
end

function Chest:draw()
	Chest.super().draw(self)
	love.graphics.setColor(1,1,1,self.alpha)
	local quad = self.quads[self.open and 2 or 1]
	love.graphics.draw(self.img, quad, self.x + self.hitPad, self.y + self.hitPad)
end

function Chest:hit(player)
	if not self.open then
		self.open = true

		-- spawn coins
		local speed = 60
		for i = 1, love.math.random(unpack(self.value)) do
			local s = {
				velx = love.math.random(-speed, speed),
				vely = -love.math.random(speed/3, speed),
			}
			self.area:addGameObject("Coin", self.x + self.w/2 - 2, self.y, s):addPhysics()
		end

		self.timer:tween(1.5, self, {alpha = 0}, "linear", function()
			self.dead = true
		end)
	end
end
