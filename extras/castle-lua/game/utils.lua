
function requireFolder(folder)
	local items = love.filesystem.getDirectoryItems(folder)
	for _, item in ipairs(items) do
		local file = folder .. '/' .. item
		local file_type = love.filesystem.getInfo(file).type
		if file_type == "file" and file:sub(-4, -1) == ".lua" then
			require(file:sub(1, -5))
		elseif file_type == "directory" then
			requireFolder(file)
		end
	end
end

function aabb(ax1,ay1,aw,ah, bx1,by1,bw,bh)
	local ax2,ay2,bx2,by2 = ax1 + aw, ay1 + ah, bx1 + bw, by1 + bh
	return ax1 < bx2 and ax2 > bx1 and ay1 < by2 and ay2 > by1
end

function shortaabb(a, b)
	if a.w then
		a.width = a.w
		a.height = a.h
	end
	if b.w then
		b.width = b.w
		b.height = b.h
	end
	return aabb(a.x, a.y, a.width, a.height, b.x, b.y, b.width, b.height)
end

function getAngle(a, b) -- angle from a.x.y -> b.x.y
	local tan = math.atan2(a.x - b.x, a.y - b.y) + math.pi/2
	return math.cos(tan), math.sin(-tan), tan
end

-- floor coordinates to avoid tile bleeding
local old_draw = love.graphics.draw
local floor = math.floor
function love.graphics.draw(...)
	local arg = {...}
	for i, v in ipairs(arg) do if type(v) == "number" then arg[i] = floor(v + 0.5) end end
	old_draw(unpack(arg))
end

-- keep changes to alpha on setColor
local old_setColor = love.graphics.setColor
function love.graphics.setColor(...)
	local _,_,_,a = love.graphics.getColor()
	local arg = {...}
	if type(arg[1]) == "table" then arg = arg[1] end
	if not arg[4] then arg[4] = a end
	old_setColor(arg)
end

--[[
async(function(wait, cont)
	after(5, cont) wait()
	print("wait 5")
end)
async(function(wait, cont)
	battle(5, cont)
	local args = wait()
	print(args.result)
end)
]]
function async(f) -- async(function(wait, cont) after(5, "after text", cont) local arg = wait() print(arg) end)
	local co = coroutine.wrap(f)
	co(coroutine.yield, co)
end

