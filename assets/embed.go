package assets

import (
	"bytes"
	embed "embed"
	"log"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

var (
	// FS is the embedded file system for all assets.
	//go:embed *.png *.json
	FS embed.FS

	//go:embed m5x7.ttf
	m5x7File []byte
	//go:embed m6x11.ttf
	m6x11File []byte
	//go:embed nano.ttf
	nanoFile []byte
)

var (
	M5x7Font  *text.GoTextFace
	M6x11Font *text.GoTextFace
	NanoFont  *text.GoTextFace
)

//nolint:mnd
func init() {
	M5x7Font = loadFont(m5x7File, 16)
	M6x11Font = loadFont(m6x11File, 16)
	NanoFont = loadFont(nanoFile, 6)
}

func loadFont(data []byte, size float64) *text.GoTextFace {
	face, err := text.NewGoTextFaceSource(bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}

	return &text.GoTextFace{Source: face, Size: size}
}
