
io.stdout:setvbuf("no")

--[[
	Test new game

	-- make the maps the elemental unit (in a folder: map tiles png, backgrounds, code, map-only-objects?)
	-- ^ have a separate entity tiles png?

	-- maybe make enemy freeze for a second when hit instead of jump?
	-- make the hook attack by holding the spear button

	-- a simple reset when dying or falling out of bounce
	-- maybe have controllable jump height

	-- maybe make a new hitbox for hurting enemys (bopbox?) for enemys that can only be hurt on the head (jump and spear)
	-- coins doesn't behave well with bump? I dont't know why

	-- make an enemy that shoots
	-- make the spawner enemies disapear when out of camera

	-- make enemy turn when reaches the edge of the camera (for redSnake y ghoul)
	-- make enemy turn to player when hooked (like the greensnake) (the run to the other direction making it dificult to hit)
	-- make permanent dead entities

	-- make hooked on air enemies that are speared to hit 3 for damage (to one shot snakes)

	-- make a hitable block that drops coins (its fun to jump and hit them)

	-- spawner broke

	-- fix pixel bleeding

	-- maybe fix the double damage on spear (spear hitbox is on for too long?)
]]

-- libs
Object = require("libs/classic")
Input = require("libs/boipushy")
Timer = require("libs/chrono")
humpera = require("libs/humpcamera")
bump = require("libs/bump")
sti = require("libs/sti")
anim8 = require("libs/anim8")
if debug then inspect = require("libs/inspect") end
-- others
require("utils")

function love.load()
	input = Input()
	love.graphics.setDefaultFilter("nearest", "nearest")
	love.graphics.setLineStyle("rough")
	mainCanvas = love.graphics.newCanvas(wx, wy)

	requireFolder("engine")
	requireFolder("scenes")
	requireFolder("objects")
	requireFolder("maps")

	input:bind("w","up")
	input:bind("a","left")
	input:bind("s","down")
	input:bind("d","right")

	input:bind("up","up")
	input:bind("left","left")
	input:bind("down","down")
	input:bind("right","right")

	input:bind("z","action")
	input:bind("x","back")
	input:bind("p","pause")

	scene = Scene()
	scene:goto("Game")

	if debug then
		GState = require("GState")
		input:bind("q", function() debug = not debug end)
		input:bind("escape", function() love.event.quit() end)
	end
end

function love.update(dt)
	love.window.setTitle(love.timer.getFPS())
	scene:update(dt)
end

function love.draw()
	love.graphics.setCanvas(mainCanvas)
		love.graphics.clear()
		if fov > 1 then love.graphics.translate((wx/fov)/2, (wy/fov)/2) end
		scene:draw()
	love.graphics.setCanvas()
	love.graphics.draw(mainCanvas, 0, 0, 0, sx, sy)
end
