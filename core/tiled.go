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

var (
	LayerIndex    = 2
	entityObjects = map[string]EntityContructor{}
)

type Properties struct {
	FlipX, FlipY bool
	View         *tiled.Object
	Custom       map[string]string
}

type EntityContructor func(x, y, w, h float64, props *Properties) Entity

type Tile struct {
	X, Y     float64
	Image    *ebiten.Image
	ImageTag string
}

type animationFrame struct {
	image    *ebiten.Image
	duration float64
}

type animationPosition struct {
	x, y                float64
	flipX, flipY, flipR bool
}

type animation struct {
	frames    []*animationFrame
	positions []animationPosition
	timer     float64
	current   int
}

type layerData struct {
	image      *ebiten.Image
	animations map[uint32]*animation
	fs         fs.FS
}

type Map struct {
	data                *tiled.Map
	layers              map[string][]*layerData
	tileset             map[string][]*ebiten.Image
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

func RegisterEntityName[T Entity](name string, constructor func(x, y, w, h float64, p *Properties) T) {
	entityObjects[name] = func(x, y, w, h float64, p *Properties) Entity { return constructor(x, y, w, h, p) }
}

func NewMap(mapPath string, backLayersNum int, fs fs.FS, drawImagesTags ...string) *Map {
	data, err := tiled.LoadFile(mapPath, tiled.WithFileSystem(fs))
	if err != nil {
		log.Println("Error parsing Tiled map:", err)
	}

	tilesets := map[string][]*ebiten.Image{}
	layers := map[string][]*layerData{}
	for i, tag := range drawImagesTags {
		vfs := fs
		if i > 0 {
			vfs = &extVariationFS{ext: tag, target: "png", fs: fs}
		}
		if tilesets[tag], err = buildTileset(data, vfs); err != nil {
			log.Println("Error building tileset from Tiled map:", err)
		}
		if layers[tag], err = buildLayers(data, vfs, tilesets[tag], i == len(drawImagesTags)-1); err != nil {
			log.Println("Error building layers data from Tiled map:", err)
		}
	}

	m := &Map{data, layers, tilesets, backLayersNum}
	if err := m.render(); err != nil {
		log.Println("Error rendering Tiled map:", err)
	}

	return m
}

func (m *Map) Update(dt float64) {
	for _, layers := range m.layers {
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
	for imageTag, layers := range m.layers {
		for i, layer := range layers {
			layerDepth := -LayerIndex
			if i >= m.backgroundLayersNum {
				layerDepth = LayerIndex
			}
			localImage, _ := layer.image.SubImage(camera.Bounds()).(*ebiten.Image)
			pipeline.Add(imageTag, layerDepth, func(screen *ebiten.Image) { screen.DrawImage(localImage, nil) })
			cx, cy := camera.Position()
			for _, anim := range layer.animations {
				for _, pos := range anim.positions {
					op := &ebiten.DrawImageOptions{}
					var sx, sy, dx, dy float64 = 1, 1, 0, 0
					if pos.flipR {
						op.GeoM.Rotate(math.Pi / 2)
						sx = -1
					}
					if pos.flipX {
						sx, dx = -1, float64(m.data.TileWidth)
						if pos.flipR {
							sx = 1
						}
					}
					if pos.flipY {
						sy, dy = -1, float64(m.data.TileHeight)
					}
					pos.x += dx
					pos.y += dy
					op.GeoM.Scale(sx, sy)
					op.GeoM.Translate(math.Ceil(pos.x-cx), math.Ceil(pos.y-cy))
					pipeline.Add(imageTag, layerDepth, func(screen *ebiten.Image) {
						screen.DrawImage(anim.frames[anim.current].image, op)
					})
				}
			}
		}
	}
}

func (m *Map) FindObjectID(id int) (*tiled.Object, error) {
	for _, group := range m.data.ObjectGroups {
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

loop:
	for _, layers := range m.layers {
		for _, layer := range layers {
			if anim, ok := layer.animations[gid]; ok {
				for _, pos := range anim.positions {
					positions = append(positions, [2]float64{pos.x, pos.y})
				}

				break loop
			}
		}
	}
	for _, layer := range m.data.Layers {
		for y := range m.data.Height {
			for x := range m.data.Width {
				if tile := layer.Tiles[y*m.data.Width+x]; !tile.IsNil() && tile.Tileset.FirstGID+tile.ID == gid {
					positions = append(positions, [2]float64{float64(x * m.data.TileWidth), float64(y * m.data.TileHeight)})
				}
			}
		}
	}

	return positions
}

func (m *Map) TilesFromPosition(x, y float64, removeTiles bool, space *bump.Space) (map[string]*Tile, error) {
	mapX, mapY := int(x)/m.data.TileWidth, int(y)/m.data.TileHeight
	if mapX < 0 || mapY < 0 || mapX >= m.data.Width || mapY >= m.data.Height {
		return nil, fmt.Errorf("map: position out of bounds: %f, %f", x, y)
	}
	// TODO: Check animations too
	/*
		loop:
		for _, layers := range m.layers {
			for _, layer := range layers {
				if anim, ok := layer.animations[gid]; ok {
					for _, pos := range anim.positions {
						positions = append(positions, [2]float64{pos.x, pos.y})
					}

					break loop
				}
			}
		}
	*/

	position := mapY*m.data.Width + mapX
	skipped := 0
	for layerIndex := len(m.data.Layers) - 1; layerIndex >= 0; layerIndex-- {
		layer := m.data.Layers[layerIndex]
		if !layer.Visible {
			skipped++

			continue
		}
		tile := layer.Tiles[position]
		if tile.IsNil() {
			continue
		}

		tiles := map[string]*Tile{}
		for imageTag := range m.layers {
			tileImage := &Tile{
				X: float64(mapX * m.data.TileWidth), Y: float64(mapY * m.data.TileHeight),
				Image:    m.tileset[imageTag][tile.Tileset.FirstGID+tile.ID],
				ImageTag: imageTag,
			}
			tiles[imageTag] = tileImage
		}
		if removeTiles {
			// TODO: This should work, but it's really slow and RAM consuming
			// layer.Tiles[position] = tiled.NilLayerTile
			// if err := m.render(); err != nil { // TODO: This takes to much time, and RAM
			// 	return nil, err
			// }

			for _, layers := range m.layers {
				imageLayerIndex := len(m.layers) - 1 - (len(m.data.Layers) - 1 - layerIndex + skipped)
				emptyTile := ebiten.NewImage(m.data.TileWidth, m.data.TileHeight)
				op := &ebiten.DrawImageOptions{Blend: ebiten.BlendCopy}
				op.GeoM.Translate(float64(mapX*m.data.TileWidth), float64(mapY*m.data.TileHeight))
				layers[imageLayerIndex].image.DrawImage(emptyTile, op)
			}
			if space != nil {
				space.Remove(tile)
			}
		}

		return tiles, nil
	}

	return nil, fmt.Errorf("map: no tile found at position: %f, %f", x, y)
}

func (m *Map) LoadTilesetCollisionObjects(space *bump.Space) {
	for _, tileset := range m.data.Tilesets {
		for _, tile := range tileset.Tiles {
			if len(tile.ObjectGroups) == 0 {
				continue
			}
			obj := tile.ObjectGroups[0].Objects[0]
			rect := bump.Rect{X: obj.X, Y: obj.Y, W: obj.Width, H: obj.Height}
			tags := []bump.Tag{"map"}
			if obj.Class == "ladder" || obj.Type == "ladder" {
				tags = append(tags, "passthrough", "ladder")
			}
			if obj.Class == "passthrough" || obj.Type == "passthrough" {
				tags = append(tags, "passthrough")
			}
			if obj.Polygons != nil {
				rect = polygonRect(obj)
				tags = append(tags, "slope")
			}
			for _, layer := range m.data.Layers {
				for y := range m.data.Height {
					for x := range m.data.Width {
						layerTile := layer.Tiles[y*m.data.Width+x]
						if !layerTile.IsNil() && layerTile.Tileset == tileset && layerTile.ID == tile.ID {
							x, y := float64(x*m.data.TileWidth), float64(y*m.data.TileHeight)
							tileRect := bump.Rect{X: rect.X + x, Y: rect.Y + y, W: rect.W, H: rect.H, Type: rect.Type}
							space.Set(layerTile, tileRect, tags...)
						}
					}
				}
			}
		}
	}
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
		rect := bump.Rect{X: obj.X, Y: obj.Y, W: obj.Width, H: obj.Height, Type: bump.Full}
		tags := []bump.Tag{"map"}
		if obj.Polygons != nil {
			rect = polygonRect(obj)
			tags = append(tags, "slope")
		}
		if obj.Class == "ladder" || obj.Type == "ladder" {
			tags = append(tags, "passthrough", "ladder")
		}
		if obj.Class == "passthrough" || obj.Type == "passthrough" {
			tags = append(tags, "passthrough")
		}
		space.Set(obj, rect, tags...)
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
		var construct EntityContructor
		props := &Properties{Custom: map[string]string{}}
		if obj.GID != 0 {
			tile, err := m.data.TileGIDToTile(obj.GID)
			if err != nil || tile.IsNil() {
				continue
			}
			props.FlipX = tile.HorizontalFlip
			props.FlipY = tile.VerticalFlip
			construct = entityBindMap[tile.ID]
		}
		if construct == nil {
			construct = entityObjects[obj.Name]
		}
		if construct == nil {
			var tid uint32
			if tile, err := m.data.TileGIDToTile(obj.GID); err == nil {
				tid = tile.ID
			}
			log.Printf("Warning: entity name %s or ID %d not found, skipping\n", obj.Name, tid)

			continue
		}
		for _, prop := range obj.Properties {
			switch prop.Name {
			case viewPropName:
				id, _ := strconv.Atoi(prop.Value)
				obj, err := m.FindObjectID(id)
				if err != nil {
					log.Panic("tiled: cannot find view object with id "+prop.Value+": ", err)
				}
				props.View = obj
			default:
				props.Custom[prop.Name] = prop.Value
			}
		}
		entity := construct(obj.X, obj.Y, obj.Width, obj.Height, props)
		x, y, _, h := entity.Rect()
		entity.SetPosition(x, y+obj.Height-h)
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

func (m *Map) GetObjects(objectGroupName string) []*tiled.Object {
	for _, group := range m.data.ObjectGroups {
		if objectGroupName == group.Name {
			return group.Objects
		}
	}

	return nil
}

func (m *Map) GetObjectsRects(objectGroupName string) []bump.Rect {
	objects := m.GetObjects(objectGroupName)
	if objects == nil {
		return nil
	}

	rects := make([]bump.Rect, len(objects))
	for i, obj := range objects {
		rects[i] = bump.Rect{X: obj.X, Y: obj.Y, W: obj.Width, H: obj.Height}
	}

	return rects
}

func (m *Map) render() error {
	skipped := 0
	for i, layer := range m.data.Layers {
		if !layer.Visible {
			skipped++

			continue
		}
		for _, layers := range m.layers {
			layerIndex := i - skipped
			renderer, err := render.NewRendererWithFileSystem(m.data, layers[layerIndex].fs)
			if err != nil {
				return fmt.Errorf("map: tiled map unsupported for rendering: %w", err)
			}
			if err := renderer.RenderLayer(i); err != nil {
				return fmt.Errorf("map: tiled layer %s unsupported for rendering: %w", m.data.Layers[i].Name, err)
			}
			if layers[layerIndex].image != nil {
				layers[layerIndex].image.Deallocate()
			}
			layers[layerIndex].image = ebiten.NewImageFromImage(renderer.Result)
		}
	}

	return nil
}

func buildLayers(data *tiled.Map, fs fs.FS, tileImages []*ebiten.Image, removeAnimatedTiles bool) ([]*layerData, error) {
	layersData := []*layerData{}
	for i, layer := range data.Layers {
		if !layer.Visible {
			continue
		}
		anims, err := extractLayerAnimations(data, tileImages, i, removeAnimatedTiles)
		if err != nil {
			return nil, fmt.Errorf("map: error extracting %s animations: %w", layer.Name, err)
		}
		layersData = append(layersData, &layerData{nil, anims, fs})
	}

	return layersData, nil
}

func buildTileset(data *tiled.Map, fs fs.FS) ([]*ebiten.Image, error) {
	tileImages := []*ebiten.Image{nil}
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

		for tileID := range uint32(tileset.TileCount) { //nolint: gosec
			tileImages = append(tileImages, ebiten.NewImageFromImage(cropImage(img, tileset.GetTileRect(tileID))))
		}
	}

	return tileImages, nil
}

func extractLayerAnimations(data *tiled.Map, tileImages []*ebiten.Image, layerIndex int, removeAnimatedTiles bool) (map[uint32]*animation, error) {
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
				position := animationPosition{
					x: float64(x * data.TileWidth), y: float64(y * data.TileHeight),
					flipX: tile.HorizontalFlip, flipY: tile.VerticalFlip, flipR: tile.DiagonalFlip,
				}
				animation.positions = append(animation.positions, position)
				if removeAnimatedTiles {
					layer.Tiles[tileID] = tiled.NilLayerTile
				}
			}
			tileID++
		}
	}

	return animationFrames, nil
}

func polygonRect(object *tiled.Object) bump.Rect {
	left, right, top, bottom := 0.0, 0.0, 0.0, 0.0
	points := *object.Polygons[0].Points
	for _, p := range points {
		left, right, top, bottom = math.Min(left, p.X), math.Max(right, p.X), math.Min(top, p.Y), math.Max(bottom, p.Y)
	}
	contains := [4]bool{} // topLeft, topRight, bottomLeft, bottomRight.
	for _, p := range points {
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

	return bump.Rect{X: object.X + left, Y: object.Y + top, W: right - left, H: bottom - top, Type: slope}
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
