Worm = Enemy:extend()
Worm.spritesheet = love.graphics.newImage("res/worm.png")
Worm.quads = {
	head = love.graphics.newQuad(0, 0, 10, 9, Worm.spritesheet:getWidth(), Worm.spritesheet:getHeight()),
	body1 = love.graphics.newQuad(0, 9, 9, 9, Worm.spritesheet:getWidth(), Worm.spritesheet:getHeight()),
	body2 = love.graphics.newQuad(0, 18, 9, 9, Worm.spritesheet:getWidth(), Worm.spritesheet:getHeight()),
}
local part = love.image.newImageData(1, 1)
part:setPixel(0, 0, 158/255, 69/255, 57/255)
Worm.particle = love.graphics.newImage(part)

function Worm:new(...)
	Worm.super.new(self, ...)

	self.noclip = true
	self.damage = 1
	self.hp = 5
	self.long = 5

	local padx, pady = 2, 2
	self.w, self.h = 10 - padx*2, 9 - pady*2
	self:setHitbox(self.w, self.h)

	self.offsetx = padx
	self.offsety = pady

	self.dir = self.dir or "left"
	self.r = 0

	self.jumped = false
	self.gravity = 150--100
	self.jumpForce = {30, -100}--{25, -85}
	self.physics = false

	-- hide in the floor
	self.y = self.y + 9
	self.groundLevel = self.y + 3 -- compensate tilemap spawn fix

	-- action jumps
	self.timer:after({0, 1}, function()
		self:jump()
		self.timer:every(4, function() if not self.dying then self:jump() end end)
	end)

	-- body and it's hitbox
	self.body = {}

	local padx, pady = 2, 2
	for i = 1, self.long do
		local b = {}
		b.damage, b.enemy = self.damage, true
		b.x, b.y = self.x, self.y
		b.w, b.h = 9 - padx*2, 9 - pady*2
		b.offsetx, b.offsety = padx, pady
		b.hit = function(_, d, c) self:hit(d, c) end
		if debug then b.area = self.area end

		self.area.world:add(b, b.x, b.y, b.w, b.h)
		table.insert(self.body, b)
	end

	-- head trace
	self.maxTrace = 30
	self.trace = {}
	for i = 1, self.maxTrace do table.insert(self.trace, {self.x, self.y}) end
end

function Worm:update(dt)
	Worm.super.update(self, dt)

	if self.jumped and self.y > self.groundLevel then
		self.jumped = false
		self.physics = false
		self.y = self.groundLevel
		self.velx, self.vely = 0, 0
	end

	local cols = self:updateHitbox(dt, function(other)
		if other.is and other:is(Enemy) then return false end
	end)
	--if cols then self:handleCollisions(cols) end

	-- update trace
	table.insert(self.trace, 1, {self.x, self.y})
	table.remove(self.trace)

	-- update body
	for i, b in pairs(self.body) do self:updateBody(i, b, dt) end

	-- rotate head
	local hip = self.velx == 0 and self.r or self.vely/self.velx
	self.r = math.atan(hip) - math.pi/2 + (self.velx < 0 and 0 or math.pi)
end

function Worm:updateBody(i, b, dt)
	local headOffset = 1.2/self.long
	local norm = (i-1)/(self.long-1) -- 0..1
	norm = (norm + headOffset)/(1+headOffset) -- head..1
	local pos = self.trace[math.floor(norm*(self.maxTrace-1)) + 1]
	b.x, b.y = pos[1], pos[2]
	self.area.world:update(b, b.x, b.y)
end

function Worm:draw()
	if not self.dying then love.graphics.setColor(1, 1, 1, 1) end
	if self.flash then love.graphics.setColor(220/255, 40/255, 40/255) end

	love.graphics.setScissor(0, 0, wx, self.groundLevel - camera.y + wy/2 - 3)

	local midw, midh = self.w/2, self.h/2
	love.graphics.draw(self.spritesheet, self.quads["head"], self.x + midw, self.y + midh, self.r, 1, 1, self.offsetx + midw, self.offsety + midh)

	-- draw body
	for i, b in pairs(self.body) do
		local c = i % 2 == 1 and "1" or "2"
		love.graphics.draw(self.spritesheet, self.quads["body"..c], b.x, b.y, 0, 1, 1, b.offsetx, b.offsety)
	end

	love.graphics.setScissor()

	Worm.super.draw(self)
	for i, b in pairs(self.body) do Worm.super.draw(b) end
end

function Worm:jump()
	-- going to jump animation particles
	local speed = 20
	local s = {
		img = self.particle,
		duration = {0.5, 1},
	}

	local animationTag = self.timer:every(0.05, function()
		s.velx = love.math.random(-speed, speed)
		s.vely = love.math.random(-speed)
		local x = self.x + love.math.random(0, self.w)
		local y = self.groundLevel - 3
		local p = self.area:addGameObject("Particle", x, y, s)
		self.timer:tween(s.duration[2], p, {vely = -speed/10}, "linear")
	end)

	self.timer:after(1.5, function()
		self.timer:cancel(animationTag)

		-- get turn target
		local player = self.area.scene.player
		local turnTarget = self.x + self.w/2 > player.x + player.w/2 and -1 or 1

		self.jumped = true
		self.physics = true
		self.velx, self.vely = self.jumpForce[1]*turnTarget, self.jumpForce[2]
	end)
end

function Worm:handleCollisions(cols)
	
end

function Worm:hit(damage, c)
	local velx, vely = self.velx, self.vely
	local x, y = self.x, self.y
	Worm.super.hit(self, damage, c)
	self.velx, self.vely = velx, vely
	if self.dying then
		self.velx, self.vely = 0, 0
	end
end

function Worm:deadEffect(col)
	for i, b in pairs(self.body) do
		if self.area.world:hasItem(b) then self.area.world:remove(b) end
	end
	if not self.dying then
		self.w = self.w*self.long
		self.h = self.h*self.long
		Worm.super.deadEffect(self, col)
	end
end

function Worm:destroy()
	for i, b in pairs(self.body) do
		if self.area.world:hasItem(b) then self.area.world:remove(b) end
		b.area = nil
		self.body[i] = nil
	end
	for i, v in pairs(self.trace) do self.trace[i] = nil end
	Worm.super.destroy(self)
end
