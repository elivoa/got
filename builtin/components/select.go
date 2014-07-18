package components

import (
	"github.com/elivoa/got/core/exception"
	"got/core"
)

/*
   Select Component Struct

   Key is string.
   Value is string by default.

   TODO:
     support tag `param:"data"`

*/

type Select struct {
	core.Component

	// key as option value and value as label
	Data       map[string]string // option list
	ArrayData  [][]string        // 2 level array/matrix with ordered option list.
	Name       string            // bind name
	Value      string            // current value/value bind
	Order      []string          // TODO: use this order(ordered map?)
	Header     string            // default
	AllowEmpty bool              // if true, the first value is shown as Head or empty string.(default true)
}

func (c *Select) New() *Select {
	return &Select{AllowEmpty: true}
}

func (c *Select) Setup() {
	if nil == c.Data && nil == c.ArrayData {
		panic(exception.NewCoreErrorf("Option Data Not Found in Select component, name: %s", c.Name))
	}
}

// function example
func (c *Select) IsSelected(key string) bool {
	// fmt.Printf("isselected %v == %v\n", c.Value, key)
	if c.Value == key {
		return true
	}
	return false
}

func (c *Select) OptionData() [][]string {
	if c.ArrayData != nil {
		return c.ArrayData
	}
	if nil != c.Data {
		// TODO: order
		data := [][]string{}
		for k, v := range c.Data {
			data = append(data, []string{k, v})
		}
		return data
	}
	return nil
}
