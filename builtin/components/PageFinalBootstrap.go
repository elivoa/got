package components

import (
	// bs "github.com/elivoa/got/builtin/services" // builtin services
	"github.com/elivoa/got/builtin/bootstrap"
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/core/lifecircle"
)

/*
Delegate Component
*/
type PageFinalBootstrap struct {
	core.Component

	Bootstraps *bootstrap.PageBootstraps
	Assets     *core.AssetSet
}

func (c *PageFinalBootstrap) Setup() {
	c.Bootstraps = bootstrap.GetBootstraps(c.Request())

	// replace with
	life := c.FlowLife().(*lifecircle.Life)
	var container = life.GetContainer()
	if nil != container {
		reg := container.Registry()
		c.Assets = reg.CombinedAssets()
	}
}
