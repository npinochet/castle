package game

import (
	_ "embed" // Embed is used to embed the shader file.
	"game/assets"
	"game/comps/anim"
	"game/core"
	"game/vars"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	Lights    = true
	Phosphore = true
	lightSize = 16
)

var (
	normalMapImage     = ebiten.NewImage(vars.ScreenWidth, vars.ScreenHeight)
	diffuseImage       = ebiten.NewImage(vars.ScreenWidth, vars.ScreenHeight)
	phosphoreMaskImage = ebiten.NewImage(vars.ScreenWidth*vars.Scale, vars.ScreenHeight*vars.Scale)
	screenCopy         = ebiten.NewImage(vars.ScreenWidth*vars.Scale, vars.ScreenHeight*vars.Scale)
	shaderTime         float32
	lights             []light
	lightShader        *ebiten.Shader
	phosphoreShader    *ebiten.Shader
	//go:embed light.kage
	lightShaderData []byte
	//go:embed phosphore.kage
	phosphoreShaderData []byte
)

type light struct{ x, y, size float64 }

func loadShaders(worldMap *core.Map, lightTileGID uint32) {
	if Lights {
		var err error
		if lightShader, err = ebiten.NewShader(lightShaderData); err != nil {
			log.Fatal(err)
		}
		positions := worldMap.FindTilePosition(lightTileGID)
		lights = make([]light, len(positions)+1)
		lights[len(lights)-1] = light{0, 0, 0}
		for i, pos := range positions {
			lights[i] = light{pos[0] + 4, pos[1] + 4, lightSize}
		}
	}

	if Phosphore {
		var err error
		if phosphoreShader, err = ebiten.NewShader(phosphoreShaderData); err != nil {
			log.Fatal(err)
		}

		maskImage, _, _ := ebitenutil.NewImageFromFileSystem(assets.FS, "phosphore_mask.png")
		width, height := vars.ScreenWidth*vars.Scale, vars.ScreenHeight*vars.Scale
		for y := 0; y < height; y += maskImage.Bounds().Dy() {
			for x := 0; x < width; x += maskImage.Bounds().Dx() {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(x), float64(y))
				phosphoreMaskImage.DrawImage(maskImage, op)
			}
		}
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
		diffuseImage.DrawRectShader(vars.ScreenWidth, vars.ScreenHeight, lightShader, op)
	}
	screen.DrawImage(diffuseImage, &ebiten.DrawImageOptions{CompositeMode: ebiten.CompositeModeMultiply})
}

func shaderDrawPhosphore(_ *core.Pipeline, screen *ebiten.Image) {
	if !Phosphore {
		return
	}

	screenCopy.DrawImage(screen, &ebiten.DrawImageOptions{})
	op := &ebiten.DrawRectShaderOptions{
		Uniforms: map[string]any{"Scale": float32(vars.Scale)},
		Images:   [4]*ebiten.Image{screenCopy, phosphoreMaskImage},
	}
	// panic: ebiten: all the source images must be the same size with the rectangle
	screen.DrawRectShader(vars.ScreenWidth*vars.Scale, vars.ScreenHeight*vars.Scale, phosphoreShader, op)

}
