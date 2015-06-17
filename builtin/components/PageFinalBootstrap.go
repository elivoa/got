package components

import (
	// bs "github.com/elivoa/got/builtin/services" // builtin services
	"github.com/elivoa/got/builtin/data"
	"github.com/elivoa/got/core"
)

/*
Delegate Component
*/
type PageFinalBootstrap struct {
	core.Component

	Bootstraps *data.PageBootstraps
}

func (c *PageFinalBootstrap) Setup() {
	c.Bootstraps = data.GetBootstraps(c.Request())
}
