package components

import (
	// bs "github.com/elivoa/got/builtin/services" // builtin services
	"bytes"
	"fmt"
	"github.com/elivoa/got/builtin/data"
	"github.com/elivoa/got/config"
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

	// inner fields
	containerLife *lifecircle.Life // TODO how to hide lifecircle infomation?
}

func (c *Delegate) Setup() *exit.Exit {
	if c.To == "" {
		return nil
	}
	c.ensureContainer()

	if c.Global {

		// render globally once at bottom;
		// 只需要执行第一个component即可，既然是全局唯一的那么一个页面中所有Component执行的结果应该是一样的，
		// 并且只需要一个；因此这里只执行第一个即可；
		var blockUniqueKey string = c.GlobalUniqueKey()
		if !data.Has(c.Request(), blockUniqueKey) {
			data.AddHtml(c.Request(), blockUniqueKey, template.HTML(c.executeBlock()))
		}

	} else {
		// render at the place;
		c.Html = template.HTML(c.executeBlock())
	}
	return nil // nil means using template
}

func (c *Delegate) ensureContainer() {
	life := c.FlowLife().(*lifecircle.Life)
	c.containerLife = life.GetContainer()
}

// execute block and return it's html as string;
func (c *Delegate) executeBlock() string {
	reg := c.containerLife.Registry()

	var out bytes.Buffer
	if err := templates.Engine.RenderBlock(
		&out, reg.Identity(), c.To, c.containerLife.GetProton()); err != nil {
		panic(err)
	}
	return out.String()
}

func (c *Delegate) GlobalUniqueKey() string {
	return fmt.Sprintf("%s%s%s",
		c.containerLife.Registry().Identity(), config.SPLITER_BLOCK, c.To)
}
