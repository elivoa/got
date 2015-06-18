package components

import (
	// bs "github.com/elivoa/got/builtin/services" // builtin services
	"github.com/elivoa/got/builtin/bootstrap"
	"github.com/elivoa/got/core"
)

/*
Delegate Component
*/
type PageFinalBootstrap struct {
	core.Component

	Bootstraps *bootstrap.PageBootstraps
}

func (c *PageFinalBootstrap) Setup() {
	c.Bootstraps = bootstrap.GetBootstraps(c.Request())
}
