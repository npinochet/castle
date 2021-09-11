
-- Extend hump.camera funcionality

camera = humpera.new()
camera.set = true
camera.following = nil
camera.borders = nil
camera.smoother = humpera.smooth.damped(9)
camera.box = {5, 5}

camera.shaking = false
camera.timer, camera.delay = nil, nil
camera.magnitud = nil

function camera:follow(obj)
	if not obj then self.following = nil return end
	self.following = obj
end

function camera:update(dt)

	if self.set then
		if self.following then
			local fol = self.following
			local x, y = fol.x + (fol.w and fol.w/2 or 0), fol.y + (fol.h and fol.h/2 or 0)
			local midx, midy = (wx*sx)/2, (wy*sy)/2
			local box = camera.box or {0, 0}
			self:lockWindow(x, y, midx - box[1], midx + box[1], midy - box[2], midy + box[2], self.smoother)
		end

		local bor = self.borders
		if bor then
			local x = math.max(math.min(self.x, bor[2]), bor[1])
			local y = math.max(math.min(self.y, bor[4]), bor[3])
			self:lookAt(x, y)
		end
	end

	if self.shaking then
		self.timer = self.timer + dt
		if self.timer >= self.delay then
			self.shaking = false
			return
		end

		local prog = 1 - (self.timer/self.delay)
		local shakex = love.math.random(-1, 1)*self.magnitud*prog
		local shakey = love.math.random(-1, 1)*self.magnitud*prog
		self:lookAt(self.x + shakex, self.y + shakey)
	end
end

function camera:setBorders(x, y, w, h)
	if not x then self.borders = nil return end
	self.borders = {x + wx/2, x+w - wx/2, y + wy/2, y+h - wy/2}
end

function camera:shake(time, magnitud)
	self.shaking = true
	self.delay = time
	self.timer = 0
	self.magnitud = magnitud
end
