Snake = Enemy:extend()
Snake.spritesheet = love.graphics.newImage("res/snake.png")

function Snake:new(...)
	Snake.super.new(self, ...)

	self.speed = 20
	self.damage = 1
	self.hp = 3

	local padx, pady = 0, 0
	self.w, self.h = 12 - padx*2, 7 - pady
	self:setHitbox(self.w, self.h)

	self.grid = anim8.newGrid(12, 7, self.spritesheet:getWidth(), self.spritesheet:getHeight())
	self.offsetx = padx
	self.offsety = pady

	self.dir = self.dir or "left"

	local duration = 0.5
	self.anims = {
		left = anim8.newAnimation(self.grid("1-2", 1), duration),
		right = anim8.newAnimation(self.grid("1-2", 1), duration):flipH(),
	}
	self.currentAnim = self.anims[self.dir]

	self:updateMove(self.dir)

	self.groundUpdate = true
	self.turnDelay = 1
	self.turnTimer = 0
end

function Snake:update(dt)
	Snake.super.update(self, dt)

	self.currentAnim = self.anims[self.dir]
	self.currentAnim:update(dt)

	-- get turn target
	local player = self.area.scene.player
	local turnTarget = player.x > self.x and "right" or "left"

	-- turn after a delay
	if self.ground then
		if self.dir == turnTarget then
			self.turnTimer = 0
		else
			self.turnTimer = self.turnTimer + dt
			if self.turnTimer >= self.turnDelay then
				self:updateMove(turnTarget)
			end
		end
	else
		if self.velx > 0 then
			self.dir = "right"
		else
			self.dir = "left"
		end
	end

	-- update movement when hitting ground
	if self.groundUpdate then
		if not self.ground then self.groundUpdate = false end
	else
		if self.ground then
			self.groundUpdate = true
			self:updateMove(turnTarget)
		end
	end

	local cols = self:updateHitbox(dt, function(other)
		if other.is and other:is(Enemy) then return false end
	end)
	--if cols then self:handleCollisions(cols) end
end

function Snake:draw()
	if not self.dying then love.graphics.setColor(1, 1, 1, 1) end
	if self.flash then
		love.graphics.setColor(220/255, 40/255, 40/255)
	end
	self.currentAnim:draw(self.spritesheet, self.x, self.y, 0, 1, 1, self.offsetx, self.offsety)
	Snake.super.draw(self)
end

function Snake:handleCollisions(cols)
	
end

function Snake:updateMove(target)
	local dir = target == "right" and 1 or -1
	self.dir = target
	self.velx = self.speed*dir
end
