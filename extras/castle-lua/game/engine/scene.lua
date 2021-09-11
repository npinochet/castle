Scene = Object:extend()

function Scene:new()
	self.stack = {}
	self.current_scene = nil
end

function Scene:update(dt)
	if self.current_scene then self.current_scene:update(dt) end
end

function Scene:draw()
	if self.current_scene then self.current_scene:draw() end
end

function Scene:getCurrentScene()
	return self.current_scene
end

function Scene:set(scene_name, ...)
	if not _G[scene_name] then error("There's no scene called: "..scene_name) end
	self.current_scene = _G[scene_name](...)
	if self.current_scene.scene then error("A scene has the attribute 'scene' reserved") end
	self.current_scene.scene = self
	return self.current_scene
end

function Scene:goto(scene_name, ...)
	if self.current_scene and self.current_scene.destroy then self.current_scene:destroy() end
	return self:set(scene_name, ...)
end

function Scene:push(scene_name, ...)
	if self.current_scene then
		if self.current_scene.pause then self.current_scene:pause() end
		self.stack[#self.stack + 1] = self.current_scene
	end
	return self:set(scene_name, ...)
end

function Scene:pop(...)
	if self.current_scene and self.current_scene.destroy then self.current_scene:destroy() end
	self.current_scene = self.stack[#self.stack]
	self.stack[#self.stack] = nil
	if self.current_scene and self.current_scene.resume then self.current_scene:resume(...) end
	return self.current_scene
end

function Scene:destroy()
	while self.current_scene do
		Scene:pop()
	end
end
