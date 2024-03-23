package core

import (
	"game/libs/bump"
	"log"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
	"github.com/lafriks/go-tiled/render"
)

const (
	viewPropName = "view"
)

const defaultCollisionPriority = -2

type Properties struct {
	FlipX, FlipY bool
	View         *tiled.Object
	Custom       map[string]string
}

type EntityContructor func(x, y, w, h float64, props *Properties) Entity

type Map struct {
	data                             *tiled.Map
	foregroundImage, backgroundImage *ebiten.Image
}

func NewMap(mapPath string, foregroundLayerName, backgroundLayerName string) *Map {
	data, err := tiled.LoadFile(mapPath)
	if err != nil {
		log.Println("Error parsing Tiled map:", err)
	}

	renderer, err := render.NewRenderer(data)
	if err != nil {
		log.Println("Tiled map unsupported for rendering:", err)
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
		log.Println("map: tiled layer unsupported for rendering:", err)
	}

	foreImage := ebiten.NewImageFromImage(renderer.Result)
	renderer.Clear()

	if err := renderer.RenderLayer(backLayerIndex); err != nil {
		log.Println("map: tiled layer unsupported for rendering:", err)
	}

	backImage := ebiten.NewImageFromImage(renderer.Result)

	return &Map{data, foreImage, backImage}
}

func (m *Map) Update(_ float64) {
	// TODO: Update animation system.
}

func (m *Map) FindObjectID(id int) (*tiled.Object, error) {
	for _, group := range m.data.ObjectGroups {
		for _, obj := range group.Objects {
			if obj.ID == uint32(id) {
				return obj, nil
			}
		}
	}

	return nil, tiled.ErrInvalidObjectPoint
}

func (m *Map) FindObjectFromTileID(id uint32, objectGroupName string) (*tiled.Object, error) {
	var objects []*tiled.Object
	for _, group := range m.data.ObjectGroups {
		if objectGroupName == "" {
			objects = append(objects, group.Objects...)

			continue
		}
		if objectGroupName == group.Name {
			objects = group.Objects

			break
		}
	}

	for _, obj := range objects {
		tile, err := m.data.TileGIDToTile(obj.GID)
		if err != nil || tile.IsNil() {
			continue
		}
		if tile.ID == id {
			return obj, nil
		}
	}

	return nil, tiled.ErrInvalidTileGID
}

func (m *Map) LoadBumpObjects(space *bump.Space, objectGroupName string) {
	var objects []*tiled.Object
	for _, group := range m.data.ObjectGroups {
		if objectGroupName == group.Name {
			objects = group.Objects

			break
		}
	}

	for _, obj := range objects {
		if obj.Polygons != nil {
			left, right, up, down := 0.0, 0.0, 0.0, 0.0
			for _, p := range *obj.Polygons[0].Points {
				left, right, up, down = math.Min(left, p.X), math.Max(right, p.X), math.Min(up, p.Y), math.Max(down, p.Y)
			}
			slope := bump.Slope{L: 1, R: 1}
			for _, p := range *obj.Polygons[0].Points {
				if p.Y == up {
					if p.X == left {
						slope.L = 0
					} else {
						slope.R = 0
					}
				}
			}
			rect := bump.Rect{
				X: obj.X + left, Y: obj.Y + up,
				W: right - left, H: down - up,
				Priority: defaultCollisionPriority + 1,
				Slope:    slope,
			}
			space.Set(obj, rect, "map", "slope")

			continue
		}
		space.Set(obj, bump.Rect{X: obj.X, Y: obj.Y, W: obj.Width, H: obj.Height, Priority: defaultCollisionPriority}, "map")
	}
}

func (m *Map) LoadEntityObjects(world *World, objectGroupName string, entityBindMap map[uint32]EntityContructor) {
	var objects []*tiled.Object
	for _, group := range m.data.ObjectGroups {
		if objectGroupName == group.Name {
			objects = group.Objects

			break
		}
	}

	for _, obj := range objects {
		tile, err := m.data.TileGIDToTile(obj.GID)
		if err != nil || tile.IsNil() {
			continue
		}
		if construct, ok := entityBindMap[tile.ID]; ok {
			props := &Properties{
				FlipX:  tile.HorizontalFlip,
				FlipY:  tile.VerticalFlip,
				Custom: map[string]string{},
			}
			for _, prop := range obj.Properties {
				switch prop.Name {
				case viewPropName:
					id, _ := strconv.Atoi(prop.Value)
					obj, err := m.FindObjectID(id)
					if err != nil {
						panic("tiled: cannot find object with id " + prop.Value)
					}
					props.View = obj
				default:
					props.Custom[prop.Name] = prop.Value
				}
			}
			world.Add(construct(obj.X, obj.Y-obj.Height, obj.Width, obj.Height, props))
		}
	}
}

func (m *Map) GetObjectsRects(objectGroupName string) ([]bump.Rect, bool) {
	var objects []*tiled.Object
	for _, group := range m.data.ObjectGroups {
		if objectGroupName == group.Name {
			objects = group.Objects

			break
		}
	}

	if objects == nil {
		return nil, false
	}

	rects := make([]bump.Rect, len(objects))
	for i, obj := range objects {
		rects[i] = bump.Rect{X: obj.X, Y: obj.Y, W: obj.Width, H: obj.Height}
	}

	return rects, true
}
