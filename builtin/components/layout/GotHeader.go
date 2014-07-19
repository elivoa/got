package layout

import (
	"github.com/elivoa/got/core"
)

type GOTHeader struct {
	core.Component
	Title string
}

func (c *GOTHeader) Setup()        {}
func (c *GOTHeader) BeforeRender() {}
func (c *GOTHeader) AfterRender()  {}
