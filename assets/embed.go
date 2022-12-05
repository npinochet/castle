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
	//go:embed Bitty.ttf
	bittyFile []byte
)

var (
	M5x7Font                font.Face
	BittyFont               font.Face
	defaultDPI, defaultSize = 96.0, 12.0
)

func init() {
	loadFont(m5x7File, &M5x7Font)
	loadFont(bittyFile, &BittyFont)
}

func loadFont(data []byte, target *font.Face) {
	tt, err := opentype.Parse(data)
	if err != nil {
		log.Fatal(err)
	}

	*target, err = opentype.NewFace(tt, &opentype.FaceOptions{Size: defaultSize, DPI: defaultDPI})
	if err != nil {
		log.Fatal(err)
	}
}
