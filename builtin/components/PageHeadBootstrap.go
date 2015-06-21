package components

import (
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/core/lifecircle"
)

/*
Delegate Component
*/
type PageHeadBootstrap struct {
	core.Component
}

// Special: Called by block in lifecircle-component.go
func (c *PageHeadBootstrap) Assets() *core.AssetSet {
	life := c.FlowLife().(*lifecircle.Life)
	var container = life.GetContainer()
	if nil != container {
		reg := container.Registry()
		return reg.CombinedAssets()
	}
	return nil
}
