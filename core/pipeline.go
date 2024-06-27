package core

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

// TODO: Add a cache dispose system to impove performance, there's no need to dispose the layer every frame

type DrawFunc func(image *ebiten.Image)

type Pipeline struct {
	layers map[string][]layerDraw
}

type layerDraw struct {
	layer    int
	drawFunc DrawFunc
}

func NewPipeline() *Pipeline { return &Pipeline{layers: map[string][]layerDraw{}} }

func (p *Pipeline) AddDraw(imageTag string, layer int, drawFunc DrawFunc) {
	p.layers[imageTag] = append(p.layers[imageTag], layerDraw{layer, drawFunc})
}

func (p *Pipeline) Compose(imageTag string, image *ebiten.Image) {
	layer := p.layers[imageTag]
	slices.SortStableFunc(layer, func(a, b layerDraw) int { return a.layer - b.layer })
	for _, l := range layer {
		l.drawFunc(image)
	}
	delete(p.layers, imageTag)
}
