package main

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const scale = 4

var (
	width, height = 160, 96
	nanoFont      *text.GoTextFace
)

type Game struct{}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		os.Exit(0)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {

	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {

	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	msg := "Heelow World!\nYeah! how are you today?\nI'm fine, thank you!"
	w, h := text.Measure(msg, nanoFont, nanoFont.Size+1)

	img := ebiten.NewImage(int(w), int(h))
	img.Fill(color.RGBA{0, 0, 255, 255})
	opi := &ebiten.DrawImageOptions{}
	opi.GeoM.Translate(10, 10)
	screen.DrawImage(img, opi)

	op := &text.DrawOptions{}
	op.LineSpacing = nanoFont.Size + 1
	op.GeoM.Translate(10, 11)
	text.Draw(screen, msg, nanoFont, op)

}

func (g *Game) Layout(_, _ int) (int, int) {
	return width, height
}

func main() {
	fontFile, err := os.Open("assets/nano.ttf")
	if err != nil {
		panic(err)
	}
	defer fontFile.Close()
	data, err := io.ReadAll(fontFile)
	if err != nil {
		panic(err)
	}
	face, err := text.NewGoTextFaceSource(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	nanoFont = &text.GoTextFace{Source: face, Size: 6}

	fmt.Println(nanoFont.Metrics())

	ebiten.SetWindowSize(width*scale, height*scale)
	if err := ebiten.RunGameWithOptions(&Game{}, &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryOpenGL}); err != nil {
		panic(err)
	}
}
