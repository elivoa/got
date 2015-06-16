package components

import (
	// bs "github.com/elivoa/got/builtin/services" // builtin services
	"bytes"
	"fmt"
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/core/lifecircle"
	"github.com/elivoa/got/route/exit"
	"github.com/elivoa/got/templates"
	"html/template"
)

/*
Delegate Component
*/
type Delegate struct {
	core.Component

	To     string // currectly block id;
	Global bool   // if true, render this block only once at the bottom of body;

	Html template.HTML // output, TODO change to code not template.
}

func (c *Delegate) Setup() *exit.Exit {
	if c.To == "" {
		return nil
	}

	fmt.Println("++++++++++++++ Render template ", c.To)
	if c.Global {
		// render globally once at bottom;
		// TODO
		fmt.Println("++++++++++++++++++ Render Globally")
	} else {
		// render at the place;

		life := c.FlowLife().(*lifecircle.Life)
		containerLife := life.GetContainer()
		reg := containerLife.Registry()

		var out bytes.Buffer
		if err := templates.Engine.RenderBlock(
			&out, reg.Identity(), c.To, containerLife.GetProton()); err != nil {
			panic(err)
		}
		c.Html = template.HTML(out.String())
	}
	return nil // nil means using template
}
