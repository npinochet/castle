Particle = GameObject:extend()

function Particle:new(...)
	Particle.super.new(self, ...)
	self.z = self.z or 1

	-- has animation (self.anim = {w, h})
	if self.anim then
		local w, h = self.anim[1], self.anim[2], self.anim[3]
		local frames = math.floor((self.img:getWidth()/w))

		self.grid = anim8.newGrid(w, h, self.img:getWidth(), self.img:getHeight())
		self.anim = anim8.newAnimation(self.grid("1-"..frames, 1), self.duration/frames)
	end

	if self.duration and not self.loop then
		self.timer:after(self.duration, function() self.dead = true end)
	end
end

function Particle:update(dt)
	Particle.super.update(self, dt)
	if self.velx then
		self.x, self.y = self.x + self.velx * dt, self.y + self.vely * dt
	end
	if self.anim then self.anim:update(dt) end
end

function Particle:draw()
	love.graphics.setColor(self.color or {1,1,1,1})
	if self.anim then self.anim:draw(self.img, self.x, self.y) elseif
	self.text then love.graphics.print(self.text, self.x, self.y)  else
	love.graphics.draw(self.img, self.x, self.y)
	end
end
