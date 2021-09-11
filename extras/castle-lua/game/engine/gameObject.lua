GameObject = Object:extend()
GameObject.id = 0
GameObject.physicsDefaults = {
	gravity = 300,
	maxXVelocity = 100,
	maxYVelocity = 200,
	groundFriction = 12, -- 10 - 50
	airFriction = 1,
}
GameObject.checkSolid = function(item, other)
	if item and other.isSolid then return other:isSolid(item) end
	if (other.solid
	or (other.layer and other.layer.properties.solid)
	or (other.properties and other.properties.solid)) then
		return true
	end
	return false
end

function GameObject:new(area, x, y, opts)
	local opts = opts or {}
	if opts then for k, v in pairs(opts) do self[k] = v end end
	self.area = area
	self.x, self.y = x, y
	self.timer = Timer()
	self.id = GameObject.id
	self.z = self.z or 0
	self.dead = false

	-- Set default physics
	for i, v in pairs(GameObject.physicsDefaults) do self[i] = v end

	GameObject.id = GameObject.id + 1
end

function GameObject:setState(state)
	for k, v in pairs(state) do self[k] = v end
end

function GameObject:update(dt)
	self.timer:update(dt)
end

function GameObject:draw()
	if debug then
		love.graphics.setColor(0, 1, 0, 0.6)
		if self.area.world and self.area.world:hasItem(self) then
			love.graphics.rectangle("line", self.area.world:getRect(self))
		end
	end
end

function GameObject:destroy()
	self.timer:destroy()
	if self.area.world and self.area.world:hasItem(self) then
		self.area.world:remove(self)
	end
	self.area = nil
end

function GameObject:updateHitbox(dt, customFilter)
	if self.area.world and self.area.world:hasItem(self) then

		-- apply physics
		if self.physics then
			local fric = self.groundFriction
			if not self.ground then
				fric = self.airFriction
				self.vely = self.vely + self.gravity*dt
			end

			if self.friction and not self.moving then
				self.velx = self.velx - self.velx*fric*dt
			end
		end

		if self.limitVel then
			local maxx, maxy = self.maxXVelocity, self.maxYVelocity
			self.velx = math.min(maxx, math.max(-maxx, self.velx))
			self.vely = math.min(maxy, math.max(-maxy, self.vely))
		end

		-- actual collision update
		local x, y = self.x + self.velx * dt, self.y + self.vely * dt
		local filter = function(item, other)
			if customFilter then
				local colType = customFilter(other)
				if colType ~= nil then return colType end
			end
			if not self.noclip and self.checkSolid(item, other) then return "slide" end
			return "cross"
		end
		local actualX, actualY, cols = self.area.world:move(self, x, y, filter)
		self.x, self.y = actualX, actualY

		-- check if ground quering two point (left and right) on the bottom of the hitbox
		if self.physics then
			self.ground = false
			local x, y, w, h = self.area.world:getRect(self)

			for i, qx in ipairs({x, x + w}) do
				local _, len = self.area.world:queryPoint(qx, y + h + 0.01, function(item)
					return not self.noclip and self.checkSolid(self, item)
				end)

				if len > 0 then
					self.vely = 0
					self.ground = true
					break
				end
			end
		end

		return cols, actualX - x, actualY - y
	end
end


function GameObject:setHitbox(width, height)
	if self.area.world then
		if self.area.world:hasItem(self) then
			self.area.world:update(self, self.x, self.y, width, height)
		else
			self.noclip = self.noclip == nil and false or self.noclip
			self.physics = self.physics == nil and true or self.physics
			self.friction = self.friction == nil and true or self.friction
			self.limitVel = self.limitVel == nil and true or self.limitVel

			self.velx, self.vely = self.velx or 0, self.vely or 0
			self.ground = false

			self.area.world:add(self, self.x, self.y, width, height)
		end
	end
end
