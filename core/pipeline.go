package core

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

// TODO: Add a cache dispose system to impove performance, there's no need to dispose the layer every frame

type DrawFunc func(image *ebiten.Image)

type Pipeline struct {
	layers map[string][]layerDraw
	sizes  map[string]int
}

type layerDraw struct {
	layer    int
	drawFunc DrawFunc
}

func NewPipeline() *Pipeline {
	return &Pipeline{layers: map[string][]layerDraw{}, sizes: map[string]int{}}
}

func (p *Pipeline) AddDraw(imageTag string, layer int, drawFunc DrawFunc) {
	if size := p.sizes[imageTag]; len(p.layers[imageTag]) > size {
		p.layers[imageTag][size] = layerDraw{layer, drawFunc}
	} else {
		p.layers[imageTag] = append(p.layers[imageTag], layerDraw{layer, drawFunc})
	}

	p.sizes[imageTag]++
}

func (p *Pipeline) Compose(imageTag string, image *ebiten.Image) {
	layer := p.layers[imageTag] // TODO: sort layers correctly, optimize
	slices.SortStableFunc(layer, func(a, b layerDraw) int { return a.layer - b.layer })
	for i := 0; i < p.sizes[imageTag]; i++ {
		layer[i].drawFunc(image)
	}
	p.sizes[imageTag] = 0
}

func (p *Pipeline) Dispose(imageTag string) { p.sizes[imageTag] = 0 }
