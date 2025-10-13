package shader

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
	LightSize = 16
	tileSize  = 8
)

var (
	Lights             = true
	Phosphore          = true
	normalMapImage     = ebiten.NewImage(vars.ScreenWidth, vars.ScreenHeight)
	diffuseImage       = ebiten.NewImage(vars.ScreenWidth, vars.ScreenHeight)
	phosphoreMaskImage = ebiten.NewImage(vars.ScreenWidth*vars.Scale, vars.ScreenHeight*vars.Scale)
	screenCopy         = ebiten.NewImage(vars.ScreenWidth*vars.Scale, vars.ScreenHeight*vars.Scale)
	lights             = []Light{{0, 0, 0}}
	shaderTime         float32
	lightShader        *ebiten.Shader
	phosphoreShader    *ebiten.Shader
	//go:embed light.kage
	lightShaderData []byte
	//go:embed phosphore.kage
	phosphoreShaderData []byte
)

type Light struct{ X, Y, Size float64 }

func Load(worldMap *core.Map, lightGIDs []uint32) {
	var err error
	if lightShader, err = ebiten.NewShader(lightShaderData); err != nil {
		log.Fatal(err)
	}
	for _, gid := range lightGIDs {
		for _, position := range worldMap.FindTilePosition(gid) {
			AddLight(position[0]+tileSize/2, position[1]+tileSize/2, LightSize)
		}
	}

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

func Update(dt float64) { shaderTime += float32(dt) }

func DrawLights(pipeline *core.Pipeline, screen *ebiten.Image) {
	if !Lights {
		return
	}

	normalMapImage.Fill(anim.NormalMaskColor)
	pipeline.Compose(vars.PipelineNormalMapTag, normalMapImage)
	diffuseImage.Fill(color.Black)
	cx, cy := vars.World.Camera.Position()
	for i, light := range lights {
		x, y := light.X-cx, light.Y-cy
		w, h := float64(vars.ScreenWidth), float64(vars.ScreenHeight)
		if x < -2*w || y < -2*h || x > 3*w || y > 3*h {
			// TODO: This continue breaks shader if there are no lights near, the screen turns black.
			//continue
		}

		op := &ebiten.DrawRectShaderOptions{
			Uniforms: map[string]any{
				"LightPosSize": []float32{float32(x), float32(y), float32(light.Size)},
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

func DrawPhosphore(_ *core.Pipeline, screen *ebiten.Image) {
	if !Phosphore {
		return
	}

	screenCopy.DrawImage(screen, &ebiten.DrawImageOptions{})
	op := &ebiten.DrawRectShaderOptions{
		Uniforms: map[string]any{"Scale": float32(vars.Scale)},
		Images:   [4]*ebiten.Image{screenCopy, phosphoreMaskImage},
	}
	screen.DrawRectShader(vars.ScreenWidth*vars.Scale, vars.ScreenHeight*vars.Scale, phosphoreShader, op)
}

func AddLight(x, y, size float64) *Light {
	lights = append(lights, Light{x, y, size})

	return &lights[len(lights)-1]
}
