Spawner = GameObject:extend()

function Spawner:new(...)
	Spawner.super.new(self, ...)

	self.player = self.area.scene.player
	self.entity = self.tiledObj.properties.entity
	local everyStr = self.tiledObj.properties.every or "1.8-2.5"
	self.every = {}
	for s in string.gmatch(everyStr, "([^-]+)") do table.insert(self.every, tonumber(s)) end
	self.w, self.h = self.tiledObj.width, self.tiledObj.height

	self.timerTag = false
	self.y = self.tiledObj.y

	self.groundY = tonumber(self.tiledObj.properties.groundY or self.y)
end

function Spawner:update(dt)
	Spawner.super.update(self, dt)
	if shortaabb(self.player, self) then
		if not self.timerTag then
			self.timerTag = self.timer:every(self.every, function() self:spawn() end)
		end
	else
		if self.timerTag then self.timer:cancel(self.timerTag) end
		self.timerTag = nil
	end
end

function Spawner:spawn()
	local dir = love.math.random() >= 0.5 and "right" or "left"
	local ent = self.area:addGameObject(self.entity, self.x, self.y, {dir = dir})

	ent.x = dir == "right" and camera.x - wx/2 - ent.w + 1 or camera.x + wx/2 - 1
	ent.y = self.groundY - ent.h
end

function Spawner:destroy()
	self.player = nil
	Spawner.super.destroy(self)
end