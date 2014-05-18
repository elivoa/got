package components

import (
	bs "github.com/elivoa/got/builtin/services" // builtin services
	"github.com/elivoa/got/core/lifecircle"
	"got/core"
)

/*
   Select Component Struct

   Key is string.
   Value is string by default.

   TODO:
     support tag `param:"data"`

*/
type A struct {
	core.Component

	// parameters

	Label     string
	MainBlock string // TODO
	Event     string // event name

	// TODO: Component inject can inject normal things into interface{}
	// TODO: Component inject can inject normal things array into []interface{}
	// Not only support one string parameter
	// Context   []interface{} // just like things in tapestry.
	Context string // just like things in tapestry.

	// properties
	Href string // A's href

	// services

	// TODO: use interface instead to remove *;
	// TODO: bind services and implements.
	LinkService *bs.LinkService
}

func (c *A) Setup() {
	// TODO: init services, remove this.
	c.LinkService = &bs.LinkService{Life: c.FlowLife().(*lifecircle.Life)}

	// real setup
	c.Href = c.LinkService.GenerateEventUrlIgnoreComponent(c.Event, 1, c.Context)
	// 1 means remove last component A
}
