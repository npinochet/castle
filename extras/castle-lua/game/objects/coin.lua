Coin = GameObject:extend()
Coin.img = love.graphics.newImage("res/coin.png")

function Coin:new(...)
	Coin.super().new(self, ...)

	self.item = true
	self.obtained = false
	self.value = 1
	self.physics = self.physics or false
	self.noclip = self.noclip or true
	self.limitVel = false

	self.size = self.img:getHeight() -- 3
	self.grid = anim8.newGrid(self.size, self.size, self.img:getWidth(), self.img:getHeight())
	self.anim = anim8.newAnimation(self.grid("1-6", 1), 0.1, "pauseAtStart")
	self.timer:every(1.8, function() self.anim:resume() end)

	self.z = 1

	self.hitPad = 2
	self.w, self.h = self.size + self.hitPad*2, self.size + self.hitPad
	self:setHitbox(self.w, self.h)

	self.x = self.x - self.hitPad
	self.y = self.y - self.hitPad
end

function Coin:update(dt)
	Coin.super().update(self, dt)
	self.anim:update(dt)
	self:updateHitbox(dt, function(other) if other.is and other:is(Enemy) then return false end end)
end

function Coin:draw()
	Coin.super().draw(self)
	love.graphics.setColor(1,1,1,1)
	self.anim:draw(self.img, math.floor(self.x) + self.hitPad, math.floor(self.y) + self.hitPad)
end

function Coin:obtain(player)
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

function Coin:addPhysics()
	self.physics = true
	self.noclip = false

	-- add a little no obtainable time
	self.obtained = true
	self.timer:after(0.4, function() self.obtained = false end)
end
