package game

import (
	_ "embed" // Embed is used to embed the shader file.
	"game/comps/anim"
	"game/core"
	"game/vars"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	Lights    = true
	lightSize = 16
)

var (
	normalMapImage = ebiten.NewImage(vars.ScreenWidth, vars.ScreenHeight)
	diffuseImage   = ebiten.NewImage(vars.ScreenWidth, vars.ScreenHeight)
	lights         []light
	shaderTime     float32
	shader         *ebiten.Shader
	//go:embed light.kage
	shaderData []byte
)

type light struct{ x, y, size float64 }

func shaderLoad(worldMap *core.Map, lightTileGID uint32) {
	if !Lights {
		return
	}

	var err error
	if shader, err = ebiten.NewShader(shaderData); err != nil {
		log.Fatal(err)
	}
	positions := worldMap.FindTilePosition(lightTileGID)
	lights = make([]light, len(positions)+1)
	lights[len(lights)-1] = light{0, 0, 0}
	for i, pos := range positions {
		lights[i] = light{pos[0] + 4, pos[1] + 4, lightSize}
	}
}

func shaderUpdate(dt float64) { shaderTime += float32(dt) }

func shaderDrawLights(pipeline *core.Pipeline, screen *ebiten.Image) {
	if !Lights {
		return
	}

	normalMapImage.Fill(anim.NormalMaskColor)
	pipeline.Compose(vars.PipelineNormalMapTag, normalMapImage)
	diffuseImage.Fill(color.Black)
	cx, cy := vars.World.Camera.Position()
	for i, light := range lights {
		x, y := light.x-cx, light.y-cy
		w, h := float64(vars.ScreenWidth), float64(vars.ScreenHeight)
		if x < -2*w || y < -2*h || x > 3*w || y > 3*h {
			// TODO: This continue breaks shader if there are no lights near, the screen turns black.
			//continue
		}

		op := &ebiten.DrawRectShaderOptions{
			Uniforms: map[string]any{
				"LightPosSize": []float32{float32(x), float32(y), float32(light.size)},
				"Time":         shaderTime + float32(10*i),
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
