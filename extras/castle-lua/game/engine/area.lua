Area = Object:extend()

function Area:new(scene)
	self.scene = scene
	self.game_objects = {}
end

function Area:update(dt)
	for i = #self.game_objects, 1, -1 do
		local game_object = self.game_objects[i]
		if game_object.update then game_object:update(dt) end
		if game_object.dead then
			game_object:destroy()
			table.remove(self.game_objects, i)
		end
	end
end

function Area:draw()
	-- sort draw by "z" attribute, then by "y" value over the "z" classes
	local z = {}
	local zmin, zmax = 0, 0
	for i = #self.game_objects, 1, -1 do
		local game_object = self.game_objects[i]
		local game_objectZ = game_object.z
		if game_objectZ > zmax then zmax = game_objectZ end
		if game_objectZ < zmin then zmin = game_objectZ end
		if not z[game_objectZ] then z[game_objectZ] = {} end
		table.insert(z[game_object.z], game_object)
	end
	for zi = zmin, zmax do
		if z[zi] then
			table.sort(z[zi], function(a,b) return a.y < b.y end)
			for i = 1, #z[zi] do
				if z[zi][i].draw then z[zi][i]:draw() end
			end
		end
	end
end

function Area:addGameObject(game_object_name, x, y, opts)
	if _G[game_object_name] == nil then error("Game object not found: "..game_object_name) end
	local opts = opts or {}
	local game_object = _G[game_object_name](self, x or 0, y or 0, opts)
	table.insert(self.game_objects, game_object)
	return game_object
end

function Area:addPhysicsWorld()
	self.world = bump.newWorld()
end

function Area:destroy()
    for i = #self.game_objects, 1, -1 do
        local game_object = self.game_objects[i]
        game_object:destroy()
        table.remove(self.game_objects, i)
    end
end
