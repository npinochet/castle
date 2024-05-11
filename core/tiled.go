package core

import (
	"game/libs/bump"
	"game/libs/camera"
	"image"
	"io/fs"
	"log"
	"math"
	"strconv"

	"github.com/disintegration/imaging"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
	"github.com/lafriks/go-tiled/render"
)

const (
	viewPropName = "view"
	idPropName   = "id"
)

const (
	defaultCollisionPriority = -2
	secondToMillisecond      = 1000
)

type Properties struct {
	FlipX, FlipY bool
	View         *tiled.Object
	CustomID     string
	Custom       map[string]string
}

type EntityContructor func(x, y, w, h float64, props *Properties) Entity

type AnimationFrame struct {
	image    *ebiten.Image
	duration float64
}

type Animation struct {
	frames    []*AnimationFrame
	positions [][2]float64
	timer     float64
	current   int
}

type Map struct {
	data                                       *tiled.Map
	foregroundImage, backgroundImage           *ebiten.Image
	foregroundAnimations, backgroundAnimations []*Animation
}

func NewMap(mapPath string, foregroundLayerName, backgroundLayerName string, fs fs.FS) *Map {
	data, err := tiled.LoadFile(mapPath, tiled.WithFileSystem(fs))
	if err != nil {
		log.Println("Error parsing Tiled map:", err)
	}

	renderer, err := render.NewRendererWithFileSystem(data, fs)
	if err != nil {
		log.Println("Tiled map unsupported for rendering:", err)
	}

	var foreLayerIndex, backLayerIndex = -1, -1
	for i, layer := range data.Layers {
		if foregroundLayerName == layer.Name {
			foreLayerIndex = i
		}
		if backgroundLayerName == layer.Name {
			backLayerIndex = i
		}
	}

	foregroundAnim, err := extractLayerAnimations(data, fs, foreLayerIndex)
	if err != nil {
		log.Println("map: error extracting foreground animations:", err)
	}
	backgroundAnim, err := extractLayerAnimations(data, fs, backLayerIndex)
	if err != nil {
		log.Println("map: error extracting background animations:", err)
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

	return &Map{data, foreImage, backImage, foregroundAnim, backgroundAnim}
}

func (m *Map) Update(dt float64) {
	for _, anim := range m.foregroundAnimations {
		if anim.timer += dt; anim.timer >= anim.frames[anim.current].duration {
			anim.current = (anim.current + 1) % len(anim.frames)
			anim.timer = 0
		}
	}
	for _, anim := range m.backgroundAnimations {
		if anim.timer += dt; anim.timer >= anim.frames[anim.current].duration {
			anim.current = (anim.current + 1) % len(anim.frames)
			anim.timer = 0
		}
	}
}

func (m *Map) Draw(screen *ebiten.Image, camera *camera.Camera, betweenDraw func()) {
	if m.backgroundImage != nil {
		background, _ := m.backgroundImage.SubImage(camera.Bounds()).(*ebiten.Image)
		screen.DrawImage(background, nil)
		cx, cy := camera.Position()
		for _, anim := range m.backgroundAnimations {
			for _, pos := range anim.positions {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(pos[0]-cx, pos[1]-cy)
				screen.DrawImage(anim.frames[anim.current].image, op)
			}
		}
	}
	if betweenDraw != nil {
		betweenDraw()
	}
	if m.foregroundImage != nil {
		foreground, _ := m.foregroundImage.SubImage(camera.Bounds()).(*ebiten.Image)
		screen.DrawImage(foreground, nil)
		cx, cy := camera.Position()
		for _, anim := range m.foregroundAnimations {
			frame := anim.frames[anim.current]
			for _, pos := range anim.positions {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(pos[0]-cx, pos[1]-cy)
				screen.DrawImage(frame.image, op)
			}
		}
	}
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
		tags := []bump.Tag{"map"}
		if obj.Class == "ladder" || obj.Type == "ladder" {
			tags = append(tags, "passthrough", "ladder")
		}
		if obj.Class == "passthrough" || obj.Type == "passthrough" {
			tags = append(tags, "passthrough")
		}
		space.Set(obj, bump.Rect{X: obj.X, Y: obj.Y, W: obj.Width, H: obj.Height, Priority: defaultCollisionPriority}, tags...)
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
						panic("tiled: cannot find view object with id " + prop.Value)
					}
					props.View = obj
				case idPropName:
					props.CustomID = prop.Value
				default:
					props.Custom[prop.Name] = prop.Value
				}
			}
			world.Add(construct(obj.X, obj.Y, obj.Width, obj.Height, props))
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

func extractLayerAnimations(data *tiled.Map, fs fs.FS, layerIndex int) ([]*Animation, error) {
	animationFrames := map[uint32]*Animation{}
	for _, tileset := range data.Tilesets {
		sf, err := fs.Open(tileset.GetFileFullPath(tileset.Image.Source))
		if err != nil {
			return nil, err
		}
		defer sf.Close()

		img, _, err := image.Decode(sf)
		if err != nil {
			return nil, err
		}

		for _, tile := range tileset.Tiles {
			if len(tile.Animation) == 0 {
				continue
			}
			frames := make([]*AnimationFrame, len(tile.Animation))
			for i, frame := range tile.Animation {
				duration := float64(frame.Duration) / secondToMillisecond
				fameImg := ebiten.NewImageFromImage(imaging.Crop(img, tileset.GetTileRect(frame.TileID)))
				frames[i] = &AnimationFrame{image: fameImg, duration: duration}
			}
			animationFrames[tileset.FirstGID+tile.ID] = &Animation{frames: frames}
		}
	}

	layer := data.Layers[layerIndex]
	tileID := 0
	for y := 0; y < data.Height; y++ {
		for x := 0; x < data.Width; x++ {
			if layer.Tiles[tileID].IsNil() {
				tileID++

				continue
			}
			tile := layer.Tiles[tileID]
			gid := tile.Tileset.FirstGID + tile.ID
			if animation, ok := animationFrames[gid]; ok {
				animation.positions = append(animation.positions, [2]float64{float64(x * data.TileWidth), float64(y * data.TileHeight)})
				layer.Tiles[tileID] = tiled.NilLayerTile
				animationFrames[gid] = animation
			}
			tileID++
		}
	}

	i := 0
	animations := make([]*Animation, len(animationFrames))
	for _, anim := range animationFrames {
		animations[i] = anim
		i++
	}

	return animations, nil
}
