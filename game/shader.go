package game

import (
	_ "embed"
	"game/core"
	"game/maps"
	"game/vars"
	"image/color"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	Lights    = true
	lightSize = 16
)

var (
	normalMapImage = ebiten.NewImage(vars.ScreenWidth, vars.ScreenHeight)
	diffuseImage   = ebiten.NewImage(vars.ScreenWidth, vars.ScreenHeight)
	normalTiledMap *core.Map
	lights         []light
	shaderTime     float32
	shader         *ebiten.Shader
	//go:embed light.kage
	shaderData []byte
)

type light struct{ x, y, size float64 }
type normalMapFS struct{ fs fs.FS }

func (n *normalMapFS) Open(name string) (fs.File, error) {
	dir, file := filepath.Split(name)
	if fileExt := strings.Split(file, "."); fileExt[1] == "png" {
		name = filepath.Join(dir, fileExt[0]+"_normal."+fileExt[1])
	}

	return n.fs.Open(name)
}

func shaderLoad(mapFile string, lightTileGID uint32) {
	if !Lights {
		return
	}

	var err error
	if shader, err = ebiten.NewShader(shaderData); err != nil {
		log.Fatal(err)
	}
	normalTiledMap = core.NewMap(mapFile, "foreground", "background", &normalMapFS{fs: maps.IntroFS})
	positions := normalTiledMap.FindTilePosition(lightTileGID)
	lights = make([]light, len(positions)+1)
	lights[len(lights)-1] = light{0, 0, 0}
	for i, pos := range positions {
		lights[i] = light{pos[0] + 4, pos[1] + 4, lightSize}
	}
}

func shaderUpdate(dt float64) { shaderTime += float32(dt) }

func shaderDrawLights(screen *ebiten.Image) {
	if !Lights {
		return
	}

	normalMapImage.Fill(color.NRGBA{127, 127, 255, 255})
	diffuseImage.Fill(color.Black)
	cx, cy := vars.World.Camera.Position()
	for _, light := range lights {
		x, y := light.x-cx, light.y-cy
		w, h := float64(vars.ScreenWidth), float64(vars.ScreenHeight)
		if x < -2*w || y < -2*h || x > 3*w || y > 3*h {
			continue
		}

		op := &ebiten.DrawRectShaderOptions{
			Uniforms: map[string]any{
				"Time":         shaderTime,
				"LightPosSize": []float32{float32(x), float32(y), float32(light.size)},
			},
			Images: [4]*ebiten.Image{normalMapImage},
			Blend:  ebiten.BlendLighter,
		}
		op.Blend.BlendOperationRGB = ebiten.BlendOperationMax
		op.Blend.BlendOperationAlpha = ebiten.BlendOperationMax
		diffuseImage.DrawRectShader(vars.ScreenWidth, vars.ScreenHeight, shader, op)
	}
	screen.DrawImage(diffuseImage, &ebiten.DrawImageOptions{CompositeMode: ebiten.CompositeModeMultiply})
}
