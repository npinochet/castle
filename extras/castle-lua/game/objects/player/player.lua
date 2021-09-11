Player = GameObject:extend()
Player.spritesheet = love.graphics.newImage("res/player.png")

function Player:new(...)
	Player.super.new(self, ...)

	self.player = true
	self.hp = 3
	self.coins = 0
	self.z = 1
	--self.solid = true
	self.maxXVelocity = 45

	self.speed = 350
	self.w, self.h = 6, 9
	self:setHitbox(self.w, self.h)
	self.canControl = true
	self.canTurn = not self.canControl -- can turn when can't control
	self.stillFrame = false
	self.moving = false
	self.attacking = false
	self.weapon = false

	self.jumpSpeed = 110

	self.dir = "right"
	self.idleState = "idle"
	self.animState = self.idleState
	self.grid = anim8.newGrid(15, 15, self.spritesheet:getWidth(), self.spritesheet:getHeight())
	self.offsetx = (15/2-self.w/2)
	self.offsety = (15-self.h)

	local duration = 0.2
	self.anims = {
		idle = anim8.newAnimation(self.grid(1, 1), duration),
		run = anim8.newAnimation(self.grid("2-5", 1), duration),
		preSpearAttack = anim8.newAnimation(self.grid(1, 2), duration),
		spearAttack = anim8.newAnimation(self.grid("2-4", 2), {0.1, 0.2, 0.1}),
		hookAttack = anim8.newAnimation(self.grid(1, 3), duration),
		afterHookAttack = anim8.newAnimation(self.grid(2, 3), duration),
	}
	self.currentAnim = self.anims[self.animState]
	self.canvas = love.graphics.newCanvas(100, 100)
	self.canx = self.canvas:getWidth()/2 - self.w/2
	self.cany = self.canvas:getHeight()/2

	self.flashDur = 0.1
	self.flashCount = 10
	self.flash = false
	self.flashing = false
	self.hitjump = 50

	self.extraDraw = nil
	self.endAttack = nil
end

function Player:getState()
	local s = {}
	s.dir = self.dir
	s.sword = self.sword
	s.animState = self.animState
	s.inventory = self.inventory
	s.equipped = self.equipped
	return s
end

function Player:update(dt)
	Player.super.update(self, dt)

	self.moving = false
	if self.canControl then
		self:control(dt)
	else
		if self.canTurn then
			self.dir = input:down("right") and "right" or self.dir
			self.dir = input:down("left") and "left" or self.dir
		end
	end

	self.currentAnim = self.anims[self.animState]
	self.currentAnim:update(dt)

	local cols, diffx, diffy = self:updateHitbox(dt)

	-- actually STOP when hitting a wall
	if (diffx < 0 and self.velx > 0) or (diffx > 0 and self.velx < 0) then self.velx = 0 end
	if (diffy > 0 and self.vely < 0) then self.vely = 0 end

	if cols then self:handleCollisions(cols) end
end

function Player:canvasDraw()
	love.graphics.push()
	love.graphics.clear()
	love.graphics.origin()
	love.graphics.setColor(1, 1, 1, 1)
	if self.flash then love.graphics.setColor(220/255, 40/255, 40/255) end
	self.currentAnim:draw(self.spritesheet, self.canx, self.cany, 0, 1, 1, self.offsetx, self.offsety)

	-- draw weapons and smoke
	if self.extraDraw then self:extraDraw() end
	love.graphics.pop()
end

function Player:draw()
	-- draw on canvas
	love.graphics.setCanvas(self.canvas)
	self:canvasDraw()
	love.graphics.setCanvas(mainCanvas)

	-- inverse drawing
	love.graphics.setColor(1, 1, 1, 1)
	local flip, offset = 1, 0
	if self.dir == "left" then flip, offset = -1, self.canvas:getWidth() end
	love.graphics.draw(self.canvas, self.x - self.canx, self.y - self.cany, 0, flip, 1, offset)

	Player.super.draw(self)
end

function Player:control(dt)
	if not self.stillFrame then self.animState = self.idleState end

	if self.stillFrame == "control" and input:pressed() then self.stillFrame = false end

	if input:down("right") or input:down("left") then
		self.animState = "run"
		self.moving = true
	end

	if input:down("right") then
		self.velx = self.velx + self.speed*dt
		self.dir = "right"
	end

	if input:down("left") then
		self.velx = self.velx - self.speed*dt
		self.dir = "left"
	end

	if input:down("up") or input:down("action") then
		if self.ground then
			self.ground = false
			self.vely = -self.jumpSpeed
		end
	end

	if not self.attacking then
		if input:down("down") and input:pressed("back") then
			self:hookAttack()
		elseif --[[input:down("back") or]] input:pressed("back") then -- maybe when down too?
			self:spearAttack()
		end
	end
end

function Player:spearAttack()
	local delay = 0.25 -- based on castlevania 16/60 (16 frames)

	-- faster delay if last attack was a hook
	if self.weapon and self.weapon:is(Hook) then delay = delay/2 end
	
	self.attacking = true
	self.canControl = false
	self.canTurn = true
	self.animState = "preSpearAttack"
	local s = {
		player = self,
		dir = self.dir,
		firstDelay = delay,
	}
	self.weapon = self.area:addGameObject("Spear", self.x, self.y, s)

	self.endAttack = function()
		self.endAttack = nil
		self.anims["spearAttack"].onLoop = function() end
		self.anims["spearAttack"]:pauseAtStart()
		self.currentAnim = self.anims[self.idleState]
		
		self.weapon.dead = true
		self.extraDraw = nil
		self.canControl = true
		self.attacking = false
	end

	self.timer:after(delay, function()
		self.canTurn = false
		self.animState = "spearAttack"
		self.anims[self.animState]:resume()
		self.anims[self.animState].onLoop = self.endAttack
		if not self.weapon.dead then self.weapon:nextState() end
	end)
end

function Player:hookAttack()
	self.attacking = true
	self.canControl = false
	self.animState = "hookAttack"
	local s = {
		player = self,
		dir = self.dir,
	}
	self.weapon = self.area:addGameObject("Hook", self.x, self.y, s)

	self.endAttack = function()
		self.endAttack = nil
		self.weapon:ending()
		self.animState = "afterHookAttack"
		self.stillFrame = "control"
		self.canControl = true
		self.attacking = false
		self.extraDraw = nil

		-- 0.1 seconds passes or an input is made to continue animation
		self.timer:after(0.1, function()
			if self.stillFrame == "control" then
				self.stillFrame = false
			end
		end)

		-- prevent smaller delay for spear if hook is actioned
		self.timer:after(0.5, function()
			if self.weapon and self.weapon.dead then self.weapon = nil end
		end)
	end
end

function Player:handleCollisions(cols)
	for _, c in pairs(cols) do
		local o = c.other
		if o.enemy then
			if c.normal.x == 0 then
				c.normal.x = o.dir == "right" and 1 or -1
			end
			self:hit(o.damage, c)
		end
		if o.item then o:obtain(self) end
	end
end

function Player:hit(damage, col)
	if not self.flashing then
		camera:shake(0.4, 2)
		self.flashing = true
		self.flash = true
		self.timer:every(self.flashDur,
		function() self.flash = not self.flash end,
		self.flashCount,
		function()
			self.flash = false
			self.flashing = false
		end)
		self.hp = self.hp - damage

		self.ground = false
		self.vely = -self.hitjump
		if col.normal.x < 0 then
			self.velx = -self.hitjump
		elseif col.normal.x > 0 then
			self.velx = self.hitjump
		end
		if self.endAttack then self.endAttack() end
		self.canControl = false
		self.timer:after(0.2, function()
			self.canControl = true
		end)
	end
end
