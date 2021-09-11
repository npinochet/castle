JumpThrough = GameObject:extend()

function JumpThrough:new(...)
	JumpThrough.super.new(self, ...)

	self.solid = true
	self.w, self.h = self.tiledObj.width, self.tiledObj.height
	self:setHitbox(self.w, self.h)
end

function JumpThrough:isSolid(item)
	if not item then return self.solid end
	return item.y + item.h <= self.y
end