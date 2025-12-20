package entity

import (
	"game/comps/textbox"
	"game/core"
	"game/libs/bump"
)

func init() { core.RegisterEntityName("Text", NewText) }

type Text struct {
	*core.BaseEntity
	textbox *textbox.Comp
}

func NewText(x, y, w, h float64, props *core.Properties) *Text {
	text := "No Text"
	if props.Custom["text"] != "" {
		text = props.Custom["text"]
	}
	entity := &core.BaseEntity{X: x, Y: y, W: w, H: h}
	textObj := &Text{
		BaseEntity: entity,
		textbox:    &textbox.Comp{Text: text, Area: func() bump.Rect { return bump.NewRect(entity.Rect()) }},
	}
	textObj.Add(textObj.textbox)

	return textObj
}

func (sd *Text) Init()          {}
func (sd *Text) Update(float64) {}
func (sd *Text) Destroy()       {}
