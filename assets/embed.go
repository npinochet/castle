package assets

import (
	_ "embed" // For embedding assets.
	"log"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	//go:embed m5x7.ttf
	m5x7File []byte
	//go:embed m6x11.ttf
	m6x11File []byte
	//go:embed Bitty.ttf
	bittyFile []byte
	//go:embed TeenyTiny.otf
	tinyFile []byte
)

var (
	M5x7Font                font.Face
	M6x11Font               font.Face
	BittyFont               font.Face
	TinyFont                font.Face
	defaultDPI, defaultSize = 96.0, 12.0
)

func init() {
	loadFont(m5x7File, &M5x7Font, nil)
	loadFont(m6x11File, &M6x11Font, nil)
	loadFont(bittyFile, &BittyFont, nil)
	loadFont(tinyFile, &TinyFont, &opentype.FaceOptions{Size: 3, DPI: 120})
}

func loadFont(data []byte, target *font.Face, opt *opentype.FaceOptions) {
	tt, err := opentype.Parse(data)
	if err != nil {
		log.Panic(err)
	}
	if opt == nil {
		opt = &opentype.FaceOptions{Size: defaultSize, DPI: defaultDPI}
	}
	*target, err = opentype.NewFace(tt, opt)
	if err != nil {
		log.Panic(err)
	}
}
