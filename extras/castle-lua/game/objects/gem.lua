Gem = GameObject:extend()
Gem.img = love.graphics.newImage("res/gem.png")

function Gem:new(...)
	Gem.super().new(self, ...)

	self.item = true
	self.obtained = false
	self.value = 100
	self.physics = self.physics or false
	self.noclip = self.noclip or true
	self.limitVel = false
	self.z = 1

	self.pad = 2
	self.w, self.h = self.img:getWidth() + self.pad*2, self.img:getHeight() + self.pad
	self:setHitbox(self.w, self.h)

	self.x = self.x - self.pad
	--self.y = self.y - self.pad
end

function Gem:update(dt)
	Gem.super().update(self, dt)
	self:updateHitbox(dt)
end

function Gem:draw()
	Gem.super().draw(self)
	love.graphics.setColor(1,1,1,1)
	love.graphics.draw(self.img, math.floor(self.x) + self.pad, math.floor(self.y) + self.pad)
end

function Gem:obtain(player)
	if not self.obtained then
		self.obtained = true
		self.draw = function() end
		player.coins = player.coins + self.value

		-- spawn +1
		local s = {
			z = 1,
			text = "+"..tostring(self.value),
			color = {1, 1, 1, 1},
			velx = 0,
			vely = -10,
			duration = 1,
		}
		local part = self.area:addGameObject("Particle", self.x, self.y, s)
		self.timer:tween(1, part.color, {[4] = 0}, "linear", function()
			self.dead = true
		end)
	end
end
