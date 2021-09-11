package core

import (
	"fmt"
	"game/libs/bump"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
	"github.com/lafriks/go-tiled/render"
)

type EntityContructor func(x, y float64, props map[string]interface{}) *Entity

type TiledMap struct {
	data            *tiled.Map
	foregroundImage *ebiten.Image
	backgroundImage *ebiten.Image
}

func NewTiledMap(mapPath string, foregroundLayerName, backgroundLayerName string) *TiledMap {
	data, err := tiled.LoadFromFile(mapPath)
	if err != nil {
		fmt.Println("Error parsing Tiled map:", err.Error())
	}

	renderer, err := render.NewRenderer(data)
	if err != nil {
		fmt.Println("Tiled map unsupported for rendering:", err.Error())
	}

	var foreLayerIndex, backLayerIndex = -1, -1
	for i := range data.Layers {
		if foregroundLayerName == data.Layers[i].Name {
			foreLayerIndex = i
		}
		if backgroundLayerName == data.Layers[i].Name {
			backLayerIndex = i
		}
	}

	if err := renderer.RenderLayer(foreLayerIndex); err != nil {
		fmt.Println("Tiled layer unsupported for rendering:", err.Error())
	}

	foreImage := ebiten.NewImageFromImage(renderer.Result)

	renderer.Clear()
	if err := renderer.RenderLayer(backLayerIndex); err != nil {
		fmt.Println("Tiled layer unsupported for rendering:", err.Error())
	}

	backImage := ebiten.NewImageFromImage(renderer.Result)

	return &TiledMap{data, foreImage, backImage}
}

func (t *TiledMap) Update(dt float64) {
	// Update animation system
}

func (t *TiledMap) FindObjectPosition(objectGroupName string, id uint32) (float64, float64, error) {
	var objects []*tiled.Object
	for _, group := range t.data.ObjectGroups {
		if objectGroupName == group.Name {
			objects = group.Objects
			break
		}
	}

	for _, obj := range objects {
		tile, err := t.data.TileGIDToTile(obj.GID)
		if err != nil || tile.IsNil() {
			continue
		}
		if tile.ID == id {
			return obj.X, obj.Y - float64(t.data.TileHeight), nil
		}
	}
	return 0.0, 0.0, tiled.ErrInvalidTileGID
}

func (t *TiledMap) LoadBumpObjects(space *bump.Space, objectGroupName string, solid bool) {
	var objects []*tiled.Object
	for _, group := range t.data.ObjectGroups {
		if objectGroupName == group.Name {
			objects = group.Objects
			break
		}
	}

	for _, obj := range objects {
		space.Set(obj, bump.Rect{X: obj.X, Y: obj.Y, W: obj.Width, H: obj.Height})
	}
}

func (t *TiledMap) LoadEntityObjects(world *World, objectGroupName string, entityBindMap map[uint32]EntityContructor) {
	var objects []*tiled.Object
	for _, group := range t.data.ObjectGroups {
		if objectGroupName == group.Name {
			objects = group.Objects
			break
		}
	}

	for _, obj := range objects {
		tile, err := t.data.TileGIDToTile(obj.GID)
		if err != nil || tile.IsNil() {
			continue
		}
		if contruct, ok := entityBindMap[tile.ID]; ok {
			props := map[string]interface{}{}
			props["w"], props["h"] = obj.Width, obj.Height
			for _, prop := range obj.Properties {
				props[prop.Name] = prop.Value
			}
			world.AddEntity(contruct(obj.X, obj.Y, props))
		}
	}
}

func (t *TiledMap) GetObjectsRects(objectGroupName string) (objs []bump.Rect, ok bool) {
	var objects []*tiled.Object
	var rects []bump.Rect
	for _, group := range t.data.ObjectGroups {
		if objectGroupName == group.Name {
			objects = group.Objects
			break
		}
	}

	if objects == nil {
		return nil, false
	}

	for _, obj := range objects {
		rects = append(rects, bump.Rect{X: obj.X, Y: obj.Y, W: obj.Width, H: obj.Height})
	}

	return rects, true
}
