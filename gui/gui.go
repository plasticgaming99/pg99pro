package gui

import (
	"bytes"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/plasticgaming99/pg99pro/gui/resources"
)

type Canvas struct {
	pg99proimage *ebiten.Image
}

func NewCanvas() (c Canvas) {
	r := bytes.NewReader(resources.Pg99proImage)
	e, _, err := ebitenutil.NewImageFromReader(r)
	if err != nil {
		log.Fatal(err)
	}
	c.pg99proimage = e
	return
}

func (c *Canvas) Update() error {
	return nil
}

func (c *Canvas) Draw(i *ebiten.Image) {
	i.DrawImage(c.pg99proimage, nil)
}

func (c *Canvas) Layout(x, y int) (rx, ry int) {
	return 964, 320
}

func Execute() {
	op := ebiten.RunGameOptions{
		SingleThread: true,
	}
	canvas := NewCanvas()
	ebiten.SetWindowSize(964, 320)
	ebiten.RunGameWithOptions(&canvas, &op)
}
