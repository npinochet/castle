Castle = TileMap:extend()
Castle.entityBind = {
	[114] = "Player",
	[115] = "Torch",
	[116] = "Sign",
	[117] = "Spike",
	[118] = "Anchor",
	[119] = "Chest",
	[120] = "Coin",
	[121] = "Gem",
	[164] = "Snake",
	[165] = "RedSnake",
	[166] = "Ghoul",
	--[167] = "Fred",
	[214] = "BreakTile",
	[215] = "JumpThrough",
	[216] = "BreakAllTile",
}

function Castle:new(x, y, state)
	self.super.new(self, "maps/first/castle/export.lua")

	local playerObj, index = self:getEntityObject("Player", "entity")
	self.map.layers["entity"].objects[index] = nil
	self.player = self.area:addGameObject("Player", playerObj.x, playerObj.y - playerObj.height)

	camera:follow(self.player)
	self:loadRoom()

	self.transitioning = false
end

function Castle:update(dt)
	if not self.transitioning then
		self.super.update(self, dt)

		-- room detection
		local rooms = self:getCollidingObjects("room", self.player)
		local r = rooms[1]
		if not r then
			camera:setBorders(0, 0, self.map.width*self.map.tilewidth, self.map.height*self.map.tileheight)
		elseif #rooms == 1 then
			camera:setBorders(r.x, r.y, r.width, r.height)
		else -- new room
			self:transition(r, rooms[2])
		end

	else self.timer:update(dt) end
end

function Castle:transition(room, newRoom)
	local r, nr = room, newRoom

	self.transitioning = true
	camera.set = false

	-- remove entities in current room and load new
	for i, obj in pairs(self.area.game_objects) do
		if obj ~= self.player then obj.dead = true end
	end
	self:loadRoom(nr)

	-- calculate room directions
	local dir = nr.y > r.y and "down" or "up"
	dir = nr.x == r.x and dir or (nr.x > r.x and "right" or "left")

	local pad = 3
	local camx = dir == "right" and wx/2 or nr.width - wx/2
	local x = dir == "right" and pad or nr.width - self.player.w - pad
	local camy = dir == "down" and wy/2 or nr.height - wy/2
	local y = dir == "down" and pad or nr.height - self.player.h - pad

	if dir == "down" or dir == "up" then
		camx, x = camera.x - nr.x, self.player.x - nr.x
	else
		camy, y = camera.y - nr.y, self.player.y - nr.y
	end

	-- transition
	local delay, ease = 0.8, "out-cubic"
	self.timer:tween(delay, camera, {x = nr.x + camx, y = nr.y + camy}, ease)
	self.timer:tween(delay, self.player, {x = nr.x + x, y = nr.y + y}, ease, function()
		self.transitioning = false
		camera.set = true
		camera:setBorders(nr.x, nr.y, nr.width, nr.height)
		self.player.velx, self.player.vely = 0, 0
	end)
end
