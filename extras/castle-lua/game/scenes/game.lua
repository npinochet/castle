Game = Object:extend()

function Game:new()
	self.mapScene = Scene()
	local s = self.mapScene:goto("Desert")
end

function Game:update(dt)
	camera:update(dt)
	self.mapScene:update(dt)

	local s = self.mapScene:getCurrentScene()
end

function Game:draw()
	self.mapScene:draw()
end

function Game:destroy()
	self.mapScene:destroy()
end
