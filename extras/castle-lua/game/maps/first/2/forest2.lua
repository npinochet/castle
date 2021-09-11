Forest2 = TileMap:extend()

function Forest2:new(x, y, state)
	self.super.new(self, "maps/forest/2/map.lua")
	self:loadEvents()
	self.player = self.area:addGameObject("Player", x, y, {events=self.events})
	self.rival = self.area:addGameObject("Rival", 240, 182, {eventName = "rival", eventType="actionHitbox"})
	if state then
		if state.player then self.player:setState(state.player) end
	end
	camera:reset(x, y)
	camera:setBounds(0, 0, self.map.width*self.map.tilewidth, self.map.height*self.map.tileheight)

	self:addTalkStopAnim("rival", function(anim, npc)
		if anim == "hold" or anim == "holdUp" then return "holdUpTalk" end
		return "talk"
	end)
end

function Forest2:resume(x, y, state)
	if x then
		self.player.x, self.player.y = x, y
	end
	if state and state.player then self.player:setState(state.player) end
	camera:reset(x, y)
	camera:setBounds(0, 0, self.map.width*self.map.tilewidth, self.map.height*self.map.tileheight)
	self.player:updateHitbox(0)
end

function Forest2:update(dt)
	self.super.update(self, dt)
	camera:follow(self.player.x+self.player.w/2, self.player.y+self.player.h/2)
end

function Forest2:loadEvents(specify)
	self.events = self.events or {}

	if specify and not self.events[specify] then
		self.events[specify] = self._events[specify]
		return
	end

	local e = "goto_forest1"
	self.events[e] = function(other)
		self.events[e] = nil
		self.player.canControl = false
		self.player.vely = 0
		self:transition(0.5, function()
			local s = self.scene:goto("Forest1", 365, 205, {player = self.player:getState()})
			s.player.dir = "left"
		end)
	end

	local e = "goto_forest3"
	self.events[e] = function(other)
		self.events[e] = nil
		self.player.canControl = false
		self.player.velx = 0
		self:transition(0.5, function()
			local s = self.scene:goto("Forest3", 117, 300, {player = self.player:getState()})
			s.player.dir = "up"
		end)
	end

	local e = "rival"
	self.events[e] = function(other)
		if not GState.rivalFirstEncounter then return end
		self.events[e] = nil
		self.player.animState = "idle"
		self.player.velx, self.player.vely = 0, 0
		self.player.canControl = false
		textBox:write(GState.rivalName, "Go find my owner, I can't wait to get rid of this chain and be free!", function()
			self.player.canControl = true
			self:loadEvents(e)
		end)
	end

	local e = "rival_fight"
	self.events[e] = function(other)
		self.events[e] = nil
		if GState.rivalFirstEncounter then return end
		self.player.canControl = false
		self.player.animState = "idle"
		self.player.velx, self.player.vely = 0, 0

		-- script
		async(function(wait, cont)

			---
			self:transition(0.8, cont, nil, "battle") wait()
			self.scene:push("ForestBattle", "Swarm", {
				n = 20,
				player = self.player:getState(),
			}, cont) wait()
			self.scene:pop(self.player.x, self.player.y)
			---

			camera:shake(2, 0.4)
			textBox:write("[HEY YOU!](big:2)", cont) wait()
			camera:shake(2, 0.4)
			textBox:write("[OVER HERE!](big:2)", cont) wait()
			self.player.animState = "walk"

			local target = {
				x = self.rival.x - self.player.w - 9,
				y = self.rival.y + 3,
			}
			self.timer:tween(2.5, self.player, target, "linear", cont) wait()
			self.player.x, self.player.y = target.x, target.y
			self.player.animState = "idle"

			textBox:write(GState.rivalName, "Hey, where are you heading?")
			textBox:write(GState.rivalName, "So "..GState.heroName.." leaved you alone like he always do, eh?")
			textBox:write(GState.rivalName, "[He he...](revolver:8,1) It must really stink to have such a bad owner")
			textBox:write(GState.rivalName, "Not like mine who always ")
			textBox:write(GState.rivalName, "What? you didn't liked my little joke?", cont) wait()
			self.rival.animState = "hold"
			self.timer:after(0.5, cont) wait()
			textBox:write(GState.rivalName, "You think you're better than me, don't you?")
			textBox:write(GState.rivalName, "I bet you're acting all tough just cause I'm leashed, is that right?")
			textBox:write(GState.rivalName, "You think you can come here and just challenge me at my lowest point?")
			textBox:write(GState.rivalName, "You'll see what I can do, I'll show ya!")
			textBox:write(GState.rivalName, "woof! rough!! raaf!! GRRRRR")
			textBox:write(GState.rivalName, "I FIGHT YOU NOW SFVSDSVDS", cont) wait()

			self:transition(0.8, cont, nil, "battle") wait()
			self.scene:push("ForestBattle", "RivalBattle", {
				player = self.player:getState(),
				x = 160,
				y = 130,
			}, cont)
			local battle = wait()
			self.scene:pop(self.player.x, self.player.y, {player = battle.player:getState()})
			self.player.animState = "idle"
			self.rival.animState = "idle"
			self.player.dir = "right"
			self.timer:after(1, cont) wait()

			textBox:write(GState.rivalName, "Heh... you have potential neighbor, I'll give you that")
			textBox:write(GState.rivalName, "You almost made me break a sweat")
			textBox:write(GState.rivalName, "I have a little proposal for you...")
			textBox:write(GState.rivalName, "Have you consider ditching your owner and just be free?")
			textBox:write(GState.rivalName, "Once I'm unleashed we could join forces and roam the land")
			textBox:write(GState.rivalName, "Just imagine it... we could be [UNSTOPABLE](shake:0.1)")
			textBox:write(GState.rivalName, "Would you like that? what I'm saying... of course you do")
			textBox:write(GState.rivalName, "Just find my owner, he's the only one who can release me from this obnoxious leash")
			textBox:write(GState.rivalName, "I'll be eagerly waiting for that moment. See ya!", cont) wait()

			self.player.canControl = true
			GState.rivalFirstEncounter = true
		end)
	end

	if not self._events then
		self._events = {}
		for i,v in pairs(self.events) do self._events[i] = v end
	end

end
