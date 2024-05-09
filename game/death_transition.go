package game

import (
	"game/assets"
	"game/utils"
	"game/vars"
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

var (
	grayscaleShader = []byte(`
	package main
	var Force float
	func Fragment(position vec4, texCoord vec2, colorScale vec4) vec4 {
		color := imageSrc0At(texCoord)
		gray := 0.299 * color.r + 0.587 * color.g + 0.114 * color.b
		return vec4(mix(color.r, gray, Force), mix(color.g, gray, Force), mix(color.b, gray, Force), 1)
	}`)
	textColor        = color.RGBA{203, 219, 252, 255}
	fadeImg, textImg *ebiten.Image
	grayShader       *ebiten.Shader
)

func init() {
	fadeImg = ebiten.NewImage(vars.ScreenWidth, vars.ScreenHeight)
	fadeImg.Fill(color.Black)

	textImg = ebiten.NewImage(vars.ScreenWidth, vars.ScreenHeight)
	w, _ := utils.TextSize("Game Over", assets.M6x11Font)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(vars.ScreenWidth-w)/2, 20)
	op.ColorScale.ScaleWithColor(textColor)
	utils.DrawText(textImg, "Game Over", assets.M6x11Font, op)

	shader, err := ebiten.NewShader(grayscaleShader)
	if err != nil {
		log.Panic(err)
	}
	grayShader = shader
}

type DeathTransition struct {
	freezeTime              float64
	fadeTween, overlayTween *gween.Tween
	overlayImg              *ebiten.Image
	actionKey               ebiten.Key
}

func (t *DeathTransition) Init() {
	t.freezeTime = 0.5
	vars.World.Freeze(t.freezeTime)
	vars.World.Speed = 0.5
	t.fadeTween = gween.New(0, 1, 5, ease.InQuad)
	t.overlayTween = gween.New(0, 1, 3, ease.OutQuad)

	t.overlayImg, _ = textImg.SubImage(image.Rect(0, 0, vars.ScreenWidth, vars.ScreenHeight)).(*ebiten.Image)
	t.actionKey = vars.Pad[utils.KeyAction]
	text := "Press " + t.actionKey.String() + " to respawn"
	op := &ebiten.DrawImageOptions{}
	w, h := utils.TextSize(text, assets.M5x7Font)
	op.GeoM.Translate(float64(vars.ScreenWidth-w)/2, vars.ScreenHeight-float64(h)-20)
	op.ColorScale.ScaleWithColor(textColor)
	utils.DrawText(t.overlayImg, text, assets.M5x7Font, op)

	// TODO: Implement death transition.
	// - Slow down time
	// - Fade to grayscale
	// - Fade to black
	// - Show death screen
	// - Press any key to restart
}

func (t *DeathTransition) Update(dt float64) bool {
	if t.freezeTime -= dt; t.freezeTime > 0 {
		return false
	}
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

	overlayAlpha, _ := t.overlayTween.Update(0)
	newScreen := ebiten.NewImage(vars.ScreenWidth, vars.ScreenHeight)
	ops := &ebiten.DrawRectShaderOptions{Uniforms: map[string]any{"Force": overlayAlpha}, Images: [4]*ebiten.Image{screen}}
	newScreen.DrawRectShader(vars.ScreenWidth, vars.ScreenHeight, grayShader, ops)
	screen.DrawImage(newScreen, nil)

	alpha, _ := t.fadeTween.Update(0)
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(fadeImg, op)

	op.ColorScale.Reset()
	op.ColorScale.ScaleAlpha(overlayAlpha)
	screen.DrawImage(t.overlayImg, op)
}
