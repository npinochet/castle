Spike = GameObject:extend()
Spike.img = love.graphics.newImage("res/spike.png")

function Spike:new(...)
	Spike.super.new(self, ...)

	self.enemy = true
	self.z = -1

	self.w, self.h = 6, 4
	self.offsetx, self.offsety = (8 - self.w)/2, (8 - self.h)
	self.x, self.y = self.x + self.offsetx, self.y + self.offsety

	self:setHitbox(self.w, self.h)
	self.damage = 1
	self.hit = function() end
end

function Spike:draw()
	love.graphics.setColor(1,1,1,1)
	love.graphics.draw(self.img, self.x - self.offsetx, self.y - self.offsety)
	Spike.super.draw(self)
end
