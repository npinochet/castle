Hook = GameObject:extend()
Hook.spritesheet = love.graphics.newImage("res/hook.png")
Hook.quads = {
	ball = love.graphics.newQuad(18, 2, 3, 3, Hook.spritesheet:getWidth(), Hook.spritesheet:getHeight()),
	rope = love.graphics.newQuad(18, 0, 4, 2, Hook.spritesheet:getWidth(), Hook.spritesheet:getHeight()),
	handle = {
		love.graphics.newQuad(0, 0, 9, 9, Hook.spritesheet:getWidth(), Hook.spritesheet:getHeight()),
		love.graphics.newQuad(9, 0, 9, 9, Hook.spritesheet:getWidth(), Hook.spritesheet:getHeight()),
	}
}

function Hook:new(...)
	Hook.super.new(self, ...)

	self.weapon = true
	self.damage = 0
	self.z = -1
	self.range = 50

	self.w, self.h = 3, 3
	self:setHitbox(self.w, self.h)
	self.physics = false
	self.limitVel = false
	self.noclip = true

	self.hooked = false
	self.dying = false

	self.velx = 90 * (self.dir == "right" and 1 or -1)

	self.balloffset = 4
	self.x = self.player.x + self.player.w/2 - self.w/2
	self.y = self.player.y + self.balloffset

	self.rope = {
		x = 3 * (self.dir == "right" and -1 or 1),
		y = 0,
	}

	self.player.extraDraw = function(player)
		-- Draw handle over player
		love.graphics.setColor(1, 1, 1, 1)
		local state = self.dying and 2 or 1
		local x, y = player.canx - 2, player.cany + 0
		love.graphics.draw(self.spritesheet, self.quads.handle[state], x, y)
	end
end

function Hook:update(dt)
	Hook.super.update(self, dt)

	self.y = self.player.y + self.balloffset

	if not self.dying and not self.hooked then
		if self:getDist() > self.range then
			self:ending()
		end

		local cols = self:updateHitbox(dt)
		if cols then self:handleCollisions(cols) end
	end
end

function Hook:draw()
	love.graphics.setColor(1, 1, 1, 1)

	--draw rope
	local ropeN = math.floor((self.x - self.player.x + self.player.w)/4) - 2
	if self.dir == "left" then
		ropeN = math.floor((self.player.x - self.x + self.w)/4) - 1
	end
	for i = 1, ropeN do
		local shift = 4*(i-1) * (self.dir == "left" and -1 or 1)
		love.graphics.draw(self.spritesheet, self.quads.rope, self.x + self.rope.x - shift, self.y + self.rope.y)
	end

	-- draw ball
	love.graphics.draw(self.spritesheet, self.quads.ball, self.x, self.y)

	Hook.super.draw(self)
end

function Hook:ending()
	if not self.dying then
		self.dying = true
		self.balloffset = self.balloffset - 2
		local tag = self.timer:tween(0.1, self, {x = self.player.x + self.player.w/2 - self.w/2}, "linear")

		-- correct tween in the middle
		self.timer:after(0.05, function()
			self.timer:cancel(tag)
			self.timer:tween(0.05, self, {x = self.player.x + self.player.w/2 - self.w/2}, "linear", function()
				if self.player.endAttack then self.player.endAttack() end
				self.dead = true
			end)
		end)
	end
end

function Hook:getDist()
	return math.abs((self.x + self.w/2) - (self.player.x + self.player.w/2))
end

function Hook:defaultHookPull(o, c)
	o.gravity = 100
	self.player.timer:after(0.2, function()
		o.gravity = GameObject.physicsDefaults.gravity
	end)
	o.ground = false
	o.vely = -32
	o.velx = 50 * (self.dir == "right" and -1 or 1)
	--o:hit(self.damage, c)
end

function Hook:handleCollisions(cols)
	for _, c in pairs(cols) do
		local o = c.other

		if o.hookPull or o.enemy then
			self.hooked = true
			c.normal.x = self.dir == "right" and -1 or 1
			self.velx = 0
			o.velx = 0
			self.player.timer:after(0.1, function()
				self:ending()
				if o.hookPull then o:hookPull(self.player, c, self:getDist())
				else self:defaultHookPull(o, c) end
			end)
			break
		end
	end
end
