package core

import (
	"fmt"
	"game/libs/bump"
	"game/libs/camera"
	"image"
	"io/fs"
	"log"
	"math"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
	"github.com/lafriks/go-tiled/render"
)

const viewPropName = "view"

const (
	defaultCollisionPriority = -2
	secondToMillisecond      = 1000
)

var LayerIndex = 2

type Properties struct {
	FlipX, FlipY bool
	View         *tiled.Object
	Custom       map[string]string
}

type EntityContructor func(x, y, w, h float64, props *Properties) Entity

type animationFrame struct {
	image    *ebiten.Image
	duration float64
}

type animation struct {
	frames    []*animationFrame
	positions [][2]float64
	timer     float64
	current   int
}

type layerData struct {
	image      *ebiten.Image
	animations map[uint32]*animation
	imageTag   string
}

type Map struct {
	data                *tiled.Map
	layersSet           [][]*layerData
	backgroundLayersNum int
}

type extVariationFS struct {
	ext, target string
	fs          fs.FS
}

func (fs *extVariationFS) Open(name string) (fs.File, error) {
	dir, file := path.Split(name)
	if fileExt := strings.Split(file, "."); fileExt[1] == fs.target {
		name = path.Join(dir, fmt.Sprintf("%s_%s.%s", fileExt[0], fs.ext, fileExt[1]))
	}

	return fs.fs.Open(name)
}

func NewMap(mapPath string, backLayersNum int, fs fs.FS, drawImagesTags ...string) *Map {
	data, err := tiled.LoadFile(mapPath, tiled.WithFileSystem(fs))
	if err != nil {
		log.Println("Error parsing Tiled map:", err)
	}

	layersSet := make([][]*layerData, len(drawImagesTags))
	layersSet[0], err = extractLayers(data, drawImagesTags[0], fs, false)
	if err != nil {
		log.Println("Error extracting layers data from Tiled map:", err)
	}

	for i, ext := range drawImagesTags[1:] {
		layersSet[i+1], err = extractLayers(data, ext, &extVariationFS{ext: ext, target: "png", fs: fs}, i == len(drawImagesTags)-1)
		if err != nil {
			log.Println("Error extracting layers data from Tiled map:", err)
		}
	}

	return &Map{data, layersSet, backLayersNum}
}

func (m *Map) Update(dt float64) {
	for _, layers := range m.layersSet {
		for _, layer := range layers {
			for _, anim := range layer.animations {
				if anim.timer += dt; anim.timer >= anim.frames[anim.current].duration {
					anim.current = (anim.current + 1) % len(anim.frames)
					anim.timer = 0
				}
			}
		}
	}
}

func (m *Map) Draw(pipeline *Pipeline, camera *camera.Camera) {
	for i := range m.layersSet[0] {
		layerDepth := -LayerIndex
		if i >= m.backgroundLayersNum {
			layerDepth = LayerIndex
		}
		for _, set := range m.layersSet {
			localImage, _ := set[i].image.SubImage(camera.Bounds()).(*ebiten.Image)
			pipeline.Add(set[i].imageTag, layerDepth, func(screen *ebiten.Image) { screen.DrawImage(localImage, nil) })
			cx, cy := camera.Position()
			for _, anim := range set[i].animations {
				for _, pos := range anim.positions {
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Translate(pos[0]-cx, pos[1]-cy)
					localAnim := anim
					pipeline.Add(set[i].imageTag, layerDepth, func(screen *ebiten.Image) {
						screen.DrawImage(localAnim.frames[localAnim.current].image, op)
					})
				}
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

func (m *Map) FindTilePosition(gid uint32) [][2]float64 {
	var positions [][2]float64
	for _, layer := range m.layersSet[0] {
		if anim, ok := layer.animations[gid]; ok {
			positions = append(positions, anim.positions...)
		}
	}
	for _, layer := range m.data.Layers {
		for y := 0; y < m.data.Height; y++ {
			for x := 0; x < m.data.Width; x++ {
				if tile := layer.Tiles[y*m.data.Width+x]; !tile.IsNil() && tile.Tileset.FirstGID+tile.ID == gid {
					positions = append(positions, [2]float64{float64(x * m.data.TileWidth), float64(y * m.data.TileHeight)})
				}
			}
		}
	}

	return positions
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
				default:
					props.Custom[prop.Name] = prop.Value
				}
			}
			world.AddWithID(construct(obj.X, obj.Y, obj.Width, obj.Height, props), uint(obj.ID))
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

func extractLayers(data *tiled.Map, imageTag string, fs fs.FS, last bool) ([]*layerData, error) {
	renderer, err := render.NewRendererWithFileSystem(data, fs)
	if err != nil {
		return nil, fmt.Errorf("tiled map unsupported for rendering: %w", err)
	}

	layersData := make([]*layerData, len(data.Layers))
	for i, layer := range data.Layers {
		anims, err := extractLayerAnimations(data, fs, i, last)
		if err != nil {
			return nil, fmt.Errorf("map: error extracting %s animations: %w", layer.Name, err)
		}
		renderer.Clear()
		if err := renderer.RenderLayer(i); err != nil {
			return nil, fmt.Errorf("map: tiled layer %s unsupported for rendering: %w", layer.Name, err)
		}
		layersData[i] = &layerData{ebiten.NewImageFromImage(renderer.Result), anims, imageTag}
	}

	return layersData, nil
}

func extractLayerAnimations(data *tiled.Map, fs fs.FS, layerIndex int, last bool) (map[uint32]*animation, error) {
	// TODO: this first part can be run once, it's not consistant with the rest of the file
	animationFrames := map[uint32]*animation{}
	for _, tileset := range data.Tilesets {
		sf, err := fs.Open(filepath.ToSlash(tileset.GetFileFullPath(tileset.Image.Source)))
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
			frames := make([]*animationFrame, len(tile.Animation))
			for i, frame := range tile.Animation {
				duration := float64(frame.Duration) / secondToMillisecond
				fameImg := ebiten.NewImageFromImage(imaging.Crop(img, tileset.GetTileRect(frame.TileID)))
				frames[i] = &animationFrame{image: fameImg, duration: duration}
			}
			animationFrames[tileset.FirstGID+tile.ID] = &animation{frames: frames}
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
				animationFrames[gid] = animation
				if last {
					layer.Tiles[tileID] = tiled.NilLayerTile
				}
			}
			tileID++
		}
	}

	return animationFrames, nil
}
