package main

import (
	"io"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const scale = 6

var (
	pixelPerfect       = true
	apply              = 1.0
	texture            = 0.0
	time               int
	width, height      int
	image, normalImage *ebiten.Image
	shader             *ebiten.Shader
)

type Game struct{}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		os.Exit(0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if apply == 1 {
			apply = 0
		} else {
			apply = 1
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if texture == 1 {
			texture = 0
		} else {
			texture = 1
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	time++
	cx, cy := ebiten.CursorPosition()
	op := &ebiten.DrawRectShaderOptions{
		Uniforms: map[string]any{
			"Texture": texture,
			"Apply":   float32(apply),
			"Cursor":  []float32{float32(cx), float32(cy)},
			"Time":    float32(time) / 60,
		},
		Images: [4]*ebiten.Image{image, normalImage},
	}
	if !pixelPerfect {
		op.GeoM.Scale(scale, scale)
	}
	screen.DrawRectShader(width, height, shader, op)
}

func (g *Game) Layout(_, _ int) (int, int) {
	if pixelPerfect {
		return width, height
	}

	return width * scale, height * scale
}

func main() {
	var err error
	image, _, err = ebitenutil.NewImageFromFile("./image.png")
	if err != nil {
		log.Fatal(err)
	}
	normalImage, _, err = ebitenutil.NewImageFromFile("./normal.png")
	if err != nil {
		log.Fatal(err)
	}
	width, height = image.Bounds().Dx(), image.Bounds().Dy()

	shaderFile, err := os.Open("./light.kage")
	if err != nil {
		log.Fatal(err)
	}
	data, err := io.ReadAll(shaderFile)
	if err != nil {
		log.Fatal(err)
	}
	shader, err = ebiten.NewShader(data)
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowSize(width*scale, height*scale)
	ebiten.SetWindowTitle("Shader Demo")
	if err := ebiten.RunGameWithOptions(&Game{}, &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryOpenGL}); err != nil {
		log.Fatal(err)
	}
}
