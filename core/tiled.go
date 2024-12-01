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

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
	"github.com/lafriks/go-tiled/render"
)

const viewPropName = "view"

const secondToMillisecond = 1000

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
	tileImages map[uint32]*ebiten.Image
	animations map[uint32]*animation
	imageTag   string
}

type Tile struct {
	X, Y     float64
	Position int
	Image    *ebiten.Image
	Tag      string
}

type Map struct {
	Data                *tiled.Map // TODO: set as private after fakewall is done
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
		layersSet[i+1], err = extractLayers(data, ext, &extVariationFS{ext: ext, target: "png", fs: fs}, i+1 == len(drawImagesTags)-1)
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
			/*cx, cy := camera.Position()
			for _, anim := range set[i].animations {
				for _, pos := range anim.positions {
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Translate(pos[0]-cx, pos[1]-cy)
					localAnim := anim
					pipeline.Add(set[i].imageTag, layerDepth, func(screen *ebiten.Image) {
						screen.DrawImage(localAnim.frames[localAnim.current].image, op)
					})
				}
			}*/
		}
	}
}

func (m *Map) FindObjectID(id int) (*tiled.Object, error) {
	for _, group := range m.Data.ObjectGroups {
		for _, obj := range group.Objects {
			if int(obj.ID) == id {
				return obj, nil
			}
		}
	}

	return nil, tiled.ErrInvalidObjectPoint
}

func (m *Map) FindObjectFromTileID(id uint32, objectGroupName string) (*tiled.Object, error) {
	var objects []*tiled.Object
	for _, group := range m.Data.ObjectGroups {
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
		tile, err := m.Data.TileGIDToTile(obj.GID)
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
	for _, layer := range m.Data.Layers {
		for y := range m.Data.Height {
			for x := range m.Data.Height {
				if tile := layer.Tiles[y*m.Data.Width+x]; !tile.IsNil() && tile.Tileset.FirstGID+tile.ID == gid {
					positions = append(positions, [2]float64{float64(x * m.Data.TileWidth), float64(y * m.Data.TileHeight)})
				}
			}
		}
	}

	return positions
}

func (m *Map) TilesFromPosition(x, y float64) ([]*Tile, error) {
	mapX, mapY := int(x)/m.Data.TileWidth, int(y)/m.Data.TileHeight
	if mapX < 0 || mapY < 0 || mapX >= m.Data.Width || mapY >= m.Data.Height {
		return nil, fmt.Errorf("map: position out of bounds: %f, %f", x, y)
	}
	position := mapY*m.Data.Width + mapX
	for layerIndex := len(m.Data.Layers) - 1; layerIndex >= 0; layerIndex-- {
		layer := m.Data.Layers[layerIndex]
		tile := layer.Tiles[position]
		if tile.IsNil() {
			continue
		}

		tiles := make([]*Tile, len(m.layersSet))
		for i, layerSet := range m.layersSet {
			tileImage := &Tile{
				X: float64(mapX * m.Data.TileWidth), Y: float64(mapY * m.Data.TileHeight),
				Position: position,
				Image:    layerSet[layerIndex].tileImages[tile.Tileset.FirstGID+tile.ID],
				Tag:      layerSet[layerIndex].imageTag,
			}
			tiles[i] = tileImage
		}

		return tiles, nil
	}

	return nil, fmt.Errorf("map: no tile found at position: %f, %f", x, y)
}

// TODO: Find a way to update the map after removing a tile, re-render
func (m *Map) RemoveTiles(tiles []*Tile) error {
	for _, tile := range tiles {
		for i := len(m.Data.Layers) - 1; i >= 0; i-- {
			layer := m.Data.Layers[i]
			if tile := layer.Tiles[tile.Position]; tile.IsNil() {
				continue
			}
			layer.Tiles[tile.Position] = tiled.NilLayerTile
		}
	}

	return nil
}

func (m *Map) LoadBumpObjects(space *bump.Space, objectGroupName string) {
	var objects []*tiled.Object
	for _, group := range m.Data.ObjectGroups {
		if objectGroupName == group.Name {
			objects = group.Objects

			break
		}
	}

	for _, obj := range objects {
		if obj.Polygons != nil {
			left, right, top, bottom := 0.0, 0.0, 0.0, 0.0
			for _, p := range *obj.Polygons[0].Points {
				left, right, top, bottom = math.Min(left, p.X), math.Max(right, p.X), math.Min(top, p.Y), math.Max(bottom, p.Y)
			}
			contains := [4]bool{} // topLeft, topRight, bottomLeft, bottomRight.
			for _, p := range *obj.Polygons[0].Points {
				switch {
				case p.X == left && p.Y == top:
					contains[0] = true
				case p.X == right && p.Y == top:
					contains[1] = true
				case p.X == left && p.Y == bottom:
					contains[2] = true
				case p.X == right && p.Y == bottom:
					contains[3] = true
				}
			}
			slope := bump.Full
			for i, ok := range contains {
				if ok {
					continue
				}
				switch i {
				case 0:
					slope = bump.BottomRightSlope
				case 1:
					slope = bump.BottomLeftSlope
				case 2:
					slope = bump.TopRightSlope
				case 3:
					slope = bump.TopLeftSlope
				}

				break
			}
			rect := bump.Rect{X: obj.X + left, Y: obj.Y + top, W: right - left, H: bottom - top, Type: slope}
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
		space.Set(obj, bump.Rect{X: obj.X, Y: obj.Y, W: obj.Width, H: obj.Height, Type: bump.Full}, tags...)
	}
}

func (m *Map) LoadEntityObjects(world *World, objectGroupName string, entityBindMap map[uint32]EntityContructor) {
	var objects []*tiled.Object
	for _, group := range m.Data.ObjectGroups {
		if objectGroupName == group.Name {
			objects = group.Objects

			break
		}
	}

	for _, obj := range objects {
		tile, err := m.Data.TileGIDToTile(obj.GID)
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
			entity := construct(obj.X, obj.Y, obj.Width, obj.Height, props)
			x, y, _, h := entity.Rect()
			entity.SetPosition(x, y-h+float64(m.Data.TileHeight)) // TODO: Adjust the Y on doors and other broken objects
			world.AddWithID(entity, uint(obj.ID))

			/*
				TODO: Adjust the X when flipped too
				if props.FlipX {
					imageOffset = doorW - tileSize
					x -= imageOffset
				}
				if props.FlipX {
					imageOffset = chestW - tileSize*2
					x -= chestW - tileSize
				}
			*/
		}
	}
}

func (m *Map) GetObjectsRects(objectGroupName string) ([]bump.Rect, bool) {
	var objects []*tiled.Object
	for _, group := range m.Data.ObjectGroups {
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

func extractLayers(data *tiled.Map, imageTag string, fs fs.FS, removeAnimatedTiles bool) ([]*layerData, error) {
	renderer, err := render.NewRendererWithFileSystem(data, fs)
	if err != nil {
		return nil, fmt.Errorf("map: tiled map unsupported for rendering: %w", err)
	}

	tileImages, err := extractTileImages(data, fs)
	if err != nil {
		return nil, fmt.Errorf("map: error extracting tileset tile images: %w", err)
	}
	layersData := make([]*layerData, len(data.Layers))
	skipped := 0
	for i, layer := range data.Layers {
		if !layer.Visible {
			skipped++

			continue
		}
		// TODO: This should always removeAnimatedTiles, to remove animated tile from RenderLayer, but it should also not remove it for the next extractLayers call
		anims, err := extractLayerAnimations(data, tileImages, i, removeAnimatedTiles)
		if err != nil {
			return nil, fmt.Errorf("map: error extracting %s animations: %w", layer.Name, err)
		}
		renderer.Clear()
		if err := renderer.RenderLayer(i); err != nil {
			return nil, fmt.Errorf("map: tiled layer %s unsupported for rendering: %w", layer.Name, err)
		}
		layersData[i-skipped] = &layerData{ebiten.NewImageFromImage(renderer.Result), tileImages, anims, imageTag}
	}
	layersData = layersData[:len(layersData)-skipped]

	return layersData, nil
}

func extractTileImages(data *tiled.Map, fs fs.FS) (map[uint32]*ebiten.Image, error) {
	tileImages := map[uint32]*ebiten.Image{}

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

		for tileID := range uint32(tileset.TileCount) {
			tileImages[tileset.FirstGID+tileID] = ebiten.NewImageFromImage(cropImage(img, tileset.GetTileRect(tileID)))
		}
	}

	return tileImages, nil
}

func extractLayerAnimations(data *tiled.Map, tileImages map[uint32]*ebiten.Image, layerIndex int, removeAnimatedTiles bool) (map[uint32]*animation, error) {
	// TODO: this first part can be run once, it's not consistant with the rest of the file
	animationFrames := map[uint32]*animation{}
	for _, tileset := range data.Tilesets {
		for _, tile := range tileset.Tiles {
			if len(tile.Animation) == 0 {
				continue
			}
			frames := make([]*animationFrame, len(tile.Animation))
			for i, frame := range tile.Animation {
				frames[i] = &animationFrame{
					image:    tileImages[tileset.FirstGID+frame.TileID],
					duration: float64(frame.Duration) / secondToMillisecond,
				}
			}
			animationFrames[tileset.FirstGID+tile.ID] = &animation{frames: frames}
		}
	}

	layer := data.Layers[layerIndex]
	tileID := 0
	for y := range data.Height {
		for x := range data.Width {
			if layer.Tiles[tileID].IsNil() {
				tileID++

				continue
			}
			tile := layer.Tiles[tileID]
			gid := tile.Tileset.FirstGID + tile.ID
			if animation, ok := animationFrames[gid]; ok {
				animation.positions = append(animation.positions, [2]float64{float64(x * data.TileWidth), float64(y * data.TileHeight)})
				if removeAnimatedTiles {
					layer.Tiles[tileID] = tiled.NilLayerTile
				}
			}
			tileID++
		}
	}

	return animationFrames, nil
}

func cropImage(img image.Image, crop image.Rectangle) image.Image {
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	simg, ok := img.(subImager)
	if !ok {
		panic("image does not support cropping")
	}

	return simg.SubImage(crop)
}
