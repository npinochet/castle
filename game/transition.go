package game

import (
	"game/assets"
	"game/entity"
	"game/utils"
	"game/vars"
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

var grayscaleShader = []byte(`
	package main
	var Force float
	func Fragment(_ vec4, texCoord vec2, _ vec4) vec4 {
		color := imageSrc0At(texCoord)
		gray := 0.299 * color.r + 0.587 * color.g + 0.114 * color.b
		return vec4(mix(color.r, gray, Force), mix(color.g, gray, Force), mix(color.b, gray, Force), 1)
	}
`)

type Transition interface {
	Init()
	Update(dt float64) bool
	Draw(screen *ebiten.Image)
}

type DeathTransition struct {
	fadeTween, overlayTween *gween.Tween
	fadeImg, overlayImg     *ebiten.Image
	actionKey               ebiten.Key
	grayShader              *ebiten.Shader
}

func (t *DeathTransition) Init() {
	vars.World.Speed = 0.5
	t.fadeTween = gween.New(0, 1, 5, ease.InQuad)
	t.overlayTween = gween.New(0, 1, 3, ease.OutQuad)

	t.fadeImg = ebiten.NewImage(vars.ScreenWidth, vars.ScreenHeight)
	t.fadeImg.Fill(color.Black)

	t.overlayImg = ebiten.NewImage(vars.ScreenWidth, vars.ScreenHeight)
	w, _ := utils.TextSize("Game Over", assets.M6x11Font)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(vars.ScreenWidth-w)/2, 20)
	utils.DrawText(t.overlayImg, "Game Over", assets.M6x11Font, op)

	t.actionKey = vars.Player.(*entity.Player).Pad[utils.KeyAction]
	text := "Press " + t.actionKey.String() + " to respawn"
	w, h := utils.TextSize(text, assets.M5x7Font)
	op.GeoM.Reset()
	op.GeoM.Translate(float64(vars.ScreenWidth-w)/2, vars.ScreenHeight-float64(h)-20)
	utils.DrawText(t.overlayImg, text, assets.M5x7Font, op)

	shader, err := ebiten.NewShader(grayscaleShader)
	if err != nil {
		log.Panic(err)
	}
	t.grayShader = shader

	// TODO: Implement death transition.
	// - Slow down time
	// - Fade to grayscale
	// - Fade to black
	// - Show death screen
	// - Press any key to restart
}

func (t *DeathTransition) Update(dt float64) bool {
	t.fadeTween.Update(float32(dt))
	t.overlayTween.Update(float32(dt))
	if ebiten.IsKeyPressed(t.actionKey) {
		Reset()

		return true
	}

	return false
}

func (t *DeathTransition) Draw(screen *ebiten.Image) {
	if t.fadeTween == nil {
		return
	}

	alpha, _ := t.overlayTween.Update(0)
	ops := &ebiten.DrawRectShaderOptions{}
	ops.Uniforms = map[string]any{"Force": alpha}
	ops.Images[0] = ebiten.NewImageFromImage(screen.SubImage(image.Rect(0, 0, vars.ScreenWidth, vars.ScreenHeight)))
	screen.DrawRectShader(vars.ScreenWidth, vars.ScreenHeight, t.grayShader, ops)

	alpha, _ = t.fadeTween.Update(0)
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(t.fadeImg, op)

	alpha, _ = t.overlayTween.Update(0)
	op.ColorScale.Reset()
	op.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(t.overlayImg, op)
}
