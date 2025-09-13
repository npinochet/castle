package entity

import (
	"game/comps/body"
	"game/comps/hitbox"
	"game/comps/render"
	"game/core"
	"game/ext"
	"game/libs/bump"
	"game/vars"
	"math/rand/v2"
	"sync"
	"time"
)

const fakeWallOpenNeighborDelay = 300 * time.Millisecond

var mutex sync.Mutex

type FakeWall struct {
	*core.BaseEntity
	body                 *body.Comp
	hitbox               *hitbox.Comp
	render, renderNormal *render.Comp
	open                 bool
}

// TODO: add to opened in savefile (?) this can not be done as is is destroyed when open
func NewFakeWall(x, y, _, _ float64, _ *core.Properties) *FakeWall {
	tiles, err := vars.World.Map.TilesFromPosition(x, y, true, vars.World.Space)
	if err != nil {
		panic("fake wall: Failed to get tiles from position")
	}

	wall := &FakeWall{
		BaseEntity: &core.BaseEntity{X: x, Y: y, W: tileSize, H: tileSize},
		body:       &body.Comp{Solid: true, Tags: []bump.Tag{"solid", "fakeWall"}},
		hitbox:     &hitbox.Comp{},
		render: &render.Comp{
			Image: tiles[vars.PipelineScreenTag].Image,
			Layer: core.LayerIndex,
		},
		renderNormal: &render.Comp{
			Image:  tiles[vars.PipelineNormalMapTag].Image,
			Layer:  core.LayerIndex,
			Normal: true,
		},
	}
	wall.Add(wall.body, wall.hitbox, wall.render, wall.renderNormal)

	return wall
}

func (fw *FakeWall) Init() {
	fw.hitbox.HitFunc = fw.hurt
	fw.hitbox.PushHitbox(bump.Rect{W: tileSize, H: tileSize}, hitbox.Hit, nil)
}

func (fw *FakeWall) Update(_ float64) {}

func (fw *FakeWall) Opened() bool { return fw.open }

func (fw *FakeWall) OpenInChain() {
	mutex.Lock()
	defer mutex.Unlock()

	if fw.open {
		return
	}
	fw.Open()
	vars.World.Camera.Shake(0.1, 0.1)
	for range 5 + rand.IntN(5) {
		vars.World.Add(NewSmoke(fw))
	}
	time.AfterFunc(fakeWallOpenNeighborDelay, func() {
		horizonal := ext.QueryItems(fw, bump.Rect{X: fw.X - tileSize/2, Y: fw.Y, W: tileSize * 2, H: tileSize}, "fakeWall")
		vertical := ext.QueryItems(fw, bump.Rect{X: fw.X, Y: fw.Y - tileSize/2, W: tileSize, H: tileSize * 2}, "fakeWall")
		for _, neighbor := range append(horizonal, vertical...) {
			neighbor.OpenInChain()
		}
	})
}

func (fw *FakeWall) Open() {
	if fw.open {
		return
	}
	fw.open = true
	vars.World.Remove(fw)
}

func (fw *FakeWall) hurt(other core.Entity, _ *bump.Collision, _ float64, _ hitbox.ContactType) {
	if fw.open || !core.GetFlag(other, vars.PlayerTeamFlag) {
		return
	}
	fw.OpenInChain()
}
