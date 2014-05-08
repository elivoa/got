package components

import (
	"fmt"
	"got/core"
	// "html/template"
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

	Href      string // A's href
	Label     string
	MainBlock string

	// key as option value and value as label
	Data   *map[string]string // option list
	Name   string             // bind name
	Value  string             // current value/value bind
	Order  []string           // TODO: use this order(ordered map?)
	Header string             //
}

func (c *A) Setup() {
	fmt.Println("------------------------------- A link initialized. ---------------")
}

// function example
func (c *A) IsSelected(key string) bool {
	fmt.Printf("isselected %v == %v\n", c.Value, key)
	if c.Value == key {
		return true
	}
	return false
}
