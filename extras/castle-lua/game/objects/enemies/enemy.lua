Enemy = GameObject:extend()

function Enemy:new(...)
	Enemy.super.new(self, ...)

	self.enemy = true

	self.friction = false
	self.limitVel = false
	self.z = -1
	self.dying = false

	self.flashDur = 0.05
	self.flashCount = 4
	self.flash = false
	self.flashing = false

	self.hitjump = 50

	self.value = {3, 5}
end

function Enemy:hit(damage, col)
	if not self.flashing then
		camera:shake(0.1, 1)
		self.flashing = true
		self.flash = true
		self.timer:every(self.flashDur,
			function() self.flash = not self.flash end,
			self.flashCount,
			function()
				self.flashing = false
				self.flash = false
		end)
		self.hp = self.hp - damage
		if self.hp <= 0 then
			self:deadEffect(col)
		else
			self.ground = false
			self.vely = -self.hitjump
			if col.normal.x > 0 then
				self.velx = -self.hitjump
			elseif col.normal.x < 0 then
				self.velx = self.hitjump
			end
		end
	end
end

function Enemy:deadEffect(col)
	camera:shake(0.2, 2)
	self.dying = true
	self.flash = false
	self.area.world:remove(self) -- remove hitbox

	self.circle = true
	self.alpha = 255
	self.velx = math.max(-15, math.min(15, self.velx/1.5))
	self.vely = love.math.random(-10, 5)

	self.update = function(self, dt)
		self.timer:update(dt)
		self.x, self.y = self.x + self.velx * dt, self.y + self.vely * dt
	end

	local draw_sprite = self.draw
	self.draw = function()
		if self.circle then love.graphics.circle("fill", self.x + self.w/2, self.y + self.h/2, self.w) end
		love.graphics.setColor(244/255, 244/255, 244/255, self.alpha/255)
		draw_sprite(self)
	end

	-- spawn smoke
	local speed = 10
	local s = {
		img = love.graphics.newImage("res/smoke.png"),
		anim = {7, 7},
		duration = 1,
		color = {1, 1, 1, 180/255},
	}
	for i = 1, love.math.random(3, 5) do
		s.velx = love.math.random(-speed, speed)
		s.vely = love.math.random(-speed, speed)
		local x = self.x + self.w/2 + love.math.random(-5, 5)
		local y = self.y + self.h/2 + love.math.random(-5, 5)
		self.area:addGameObject("Particle", x, y, s)
	end

	-- spawn coins
	local speed = 60
	for i = 1, love.math.random(unpack(self.value)) do
		local s = {
			velx = love.math.random(-speed, speed),
			vely = -love.math.random(speed/3, speed),
		}
		local coin = self.area:addGameObject("Coin", self.x, self.y, s):addPhysics()
	end

	self.timer:after(0.05, function() self.circle = false end)
	self.timer:tween(1, self, {alpha = 0}, "linear", function() self.dead = true end)
end

