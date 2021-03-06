package utils

import (
	"fmt"
	"reflect"
	"strconv"
)

// _________________________________________
func PrintAttributes(m interface{}) {
	attrs := Attributes(m)
	fmt.Printf(":: The attributes of (type:%v) is:\n", reflect.TypeOf(m))
	for k, v := range attrs {
		fmt.Printf("  > '%v' %v \n", k, v)
	}
	fmt.Printf("  - %v attributes in total.\n", len(attrs))
}

// _________________________________________
// Example of how to use Go's reflection
// Print the attributes of a Data Model
func Attributes(m interface{}) map[string]reflect.Type {
	typ := reflect.TypeOf(m)
	// if a pointer to a struct is passed, get the type of the dereferenced object
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// create an attribute data structure as a map of types keyed by a string.
	attrs := make(map[string]reflect.Type)
	// Only structs are supported so return an empty result if the passed object
	// isn't a struct
	if typ.Kind() != reflect.Struct {
		fmt.Printf("%v type can't have attributes inspected\n", typ.Kind())
		return attrs
	}

	// loop through the struct's fields and set the map
	for i := 0; i < typ.NumField(); i++ {
		p := typ.Field(i)
		if !p.Anonymous {
			attrs[p.Name] = p.Type
		}
	}

	return attrs
}

/* ----------------------------------------
 * Reflect related utils.
 * ----------------------------------------
 */

func ReflectPrintAttribute(m interface{}) {
	Attributes(m)
}

var (
	nilValue    = reflect.ValueOf(nil)
	emptyString = reflect.ValueOf("")
)

// cache this, use gorilla/schema's method. method delegation.
func Coercion(value string, t reflect.Type) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(value), nil
	case reflect.Int:
		intvalue, err := strconv.Atoi(value)
		if err != nil {
			return nilValue, err
		}
		return reflect.ValueOf(intvalue), nil
	case reflect.Int64:
		intvalue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nilValue, err
		}
		return reflect.ValueOf(intvalue), nil
	case reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nilValue, err
		}
		return reflect.ValueOf(floatValue), nil
	default:
		panic(fmt.Sprintf("Coercion error, type %v not supported.", t))
	}
}

func CoercionNil(t reflect.Type) reflect.Value {
	switch t.Kind() {
	case reflect.String:
		return emptyString
	default:
		return nilValue
	}
}

// ________________________________________________________________________________
// used by got/cache
//

func RemovePointer(typo reflect.Type, removeSlice bool) (t reflect.Type, isSlice bool) {
	t = typo
	if t.Kind() == reflect.Ptr { // remove ptr
		t = t.Elem()
	}
	if removeSlice {
		if isSlice = t.Kind() == reflect.Slice; isSlice { // remove slice
			t = t.Elem()
			if t.Kind() == reflect.Ptr { // remove slice.elem's ptr
				t = t.Elem()
			}
		}
	}
	return
}

func GetRootType(obj interface{}) reflect.Type {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr { // remove ptr
		t = t.Elem()
	}
	return t
}

func GetRootValue(obj interface{}) reflect.Value {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr { // remove ptr
		v = v.Elem()
	}
	return v
}
