Anchor = GameObject:extend()
Anchor.img = love.graphics.newImage("res/anchor.png")

function Anchor:new(...)
	Anchor.super.new(self, ...)

	self.z = -1

	self.w, self.h = 6, 6
	self.offsetx, self.offsety = (8 - self.w)/2, (8 - self.h)

	self:setHitbox(self.w, self.h)
	self.damage = 1
	self.hit = function() end
end

function Anchor:draw()
	love.graphics.setColor(1,1,1,1)
	love.graphics.draw(self.img, self.x - self.offsetx, self.y - self.offsety)
	Anchor.super.draw(self)
end

function Anchor:hookPull(player, col, dist)
	local dir = col.normal.x * -1

	local power = (10 + dist) * 2

	player.velx = dir * power
	player.vely =  -dist * 2
	player.ground = false

	local oldMaxX, oldMaxY = player.maxXVelocity, player.maxYVelocity

	player.maxXVelocity = math.max(power, player.maxXVelocity)
	player.maxYVelocity = math.max(power, player.maxYVelocity)

	-- limit back to normal on land after 0.1 sec (to avoid inmidiate negation)
	player.timer:after(0.1, function()
		local overwrite = player.update
		player.update = function(self, dt)
			overwrite(self, dt)
			if self.ground then
				player.maxXVelocity, player.maxYVelocity = oldMaxX, oldMaxY
				player.update = overwrite
			end
		end
	end)
end
