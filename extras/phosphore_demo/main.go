package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var (
	apply            = true
	scale            = 4.0
	shaderScale      = 4.0
	width, height    int
	image, maskImage *ebiten.Image
	shader           *ebiten.Shader
)

type Game struct{}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		os.Exit(0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		apply = !apply
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		newScale(scale + 0.1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		newScale(scale - 0.1)
	}

	return nil
}

func ApplyShader(image *ebiten.Image) *ebiten.Image {
	if image.Bounds().Size() != maskImage.Bounds().Size() {
		return nil
	}
	cx, _ := ebiten.CursorPosition()
	op := &ebiten.DrawRectShaderOptions{
		Uniforms: map[string]any{
			"Scale":  float32(scale),
			"Cursor": float32(cx),
		},
		Images: [4]*ebiten.Image{image, maskImage},
	}
	result := ebiten.NewImage(width, height)
	result.DrawRectShader(width, height, shader, op)

	return result
}

func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	screen.DrawImage(image, op)
	if apply {
		op := &ebiten.DrawImageOptions{}
		if newScreen := ApplyShader(screen); newScreen != nil {
			screen.DrawImage(newScreen, op)
		}
	}
}

func (g *Game) Layout(_, _ int) (int, int) {
	return width, height
}

func newScale(newScale float64) {
	fmt.Println(newScale)
	scale = newScale
	width, height = int(float64(image.Bounds().Dx())*scale), int(float64(image.Bounds().Dy())*scale)
	maskImageSample, _, err := ebitenutil.NewImageFromFile("./phosphore_mask.png")
	if err != nil {
		log.Fatal(err)
	}
	maskImage = ebiten.NewImage(width, height)
	// Repeat the mask image size to fill the entire screen
	for y := 0; y < height; y += maskImageSample.Bounds().Dy() {
		for x := 0; x < width; x += maskImageSample.Bounds().Dx() {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			maskImage.DrawImage(maskImageSample, op)
		}
	}
	ebiten.SetWindowSize(width, height)
}

func main() {
	var err error
	image, _, err = ebitenutil.NewImageFromFile("./image.png")
	if err != nil {
		log.Fatal(err)
	}
	newScale(scale)

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

	ebiten.SetWindowTitle("Shader Demo")
	if err := ebiten.RunGameWithOptions(&Game{}, &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryOpenGL}); err != nil {
		log.Fatal(err)
	}
}
