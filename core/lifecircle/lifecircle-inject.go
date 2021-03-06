package lifecircle

import (
	"fmt"
	"github.com/elivoa/got/config"
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/coreservice/coercion"
	"github.com/elivoa/got/debug"
	"github.com/elivoa/got/utils"
	"github.com/gorilla/schema"
	"reflect"
	"strconv"
	"strings"
)

// injectBasic injects request&response into current Life.
func (lcc *LifeCircleControl) injectBasic() *LifeCircleControl {
	lcc.injectBasicTo(lcc.current.proton)
	return lcc
}

// InjectBasicTo will inject R & W into proton, this is not necessary, make this an option.
func (lcc *LifeCircleControl) injectBasicTo(proton core.Protoner) {
	if logger.Debug() {
		logger.Printf("[Inject] Inject Basic Information:\n")
	}
	if logger.Debug() && false {
		logger.Printf("Inject proton.r <= lcc.r\n")
		logger.Printf("Inject proton.w <= lcc.w\n")
	}
	proton.SetRequest(lcc.r)
	proton.SetResponseWriter(lcc.w)
}

func (lcc *LifeCircleControl) injectPath() *LifeCircleControl {
	lcc.injectPathTo(lcc.current.proton)
	return lcc
}

// value must be Proton struct, not ptr
// everything in lcc is belong to the root page. parameter proton is inject target.
// TODO remove reflect.
func (lcc *LifeCircleControl) injectPathTo(proton core.Protoner) {

	value := reflect.ValueOf(proton)
	t, _ := utils.RemovePointer(value.Type(), false)

	values := make(map[string][]string) // used to inject
	// pathParams := extractPathParameters(lcc.r.URL.Path, lcc.pageUrl, lcc.eventName)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		var fieldKey = f.Name

		// if type is gxl.*; i.e.: a -> a.Int
		var gxlSuffix string = analysisTranslateSuffix(f.Type)
		if gxlSuffix != "" {
			fieldKey += gxlSuffix
		}

		// parse TAG: path-param | TODO Cache this.
		tagValue := f.Tag.Get(config.TAG_path_injection)
		if tagValue != "" {
			pathParamIndex, err := strconv.Atoi(tagValue)
			if err != nil {
				panic(fmt.Sprintf("TAG path-param must be numbers. not %v.", tagValue))
			}
			if pathParamIndex <= len(lcc.parameters) {
				values[fieldKey] = []string{lcc.parameters[pathParamIndex-1]}

				if logger.Debug() {
					logger.Printf("Inject path param: '%s' <= '%s'\n", f.Name, values[fieldKey])
				}
				proton.SetInjected(f.Name, true)
			}
		}
	}
	InjectValues(proton, values)
}

func (lcc *LifeCircleControl) injectURLParameter() *LifeCircleControl {
	lcc.injectURLParameterTo(lcc.current.proton)
	return lcc
}

func (lcc *LifeCircleControl) injectURLParameterTo(proton core.Protoner) {
	t := utils.GetRootType(proton)

	values := make(map[string][]string) // used to inject using gorilla
	queries := lcc.r.URL.Query()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		var fieldKey = f.Name

		// if type is gxl.*; i.e.: a -> a.Int
		var gxlSuffix string = analysisTranslateSuffix(f.Type)
		if gxlSuffix != "" {
			fieldKey += gxlSuffix
		}

		// query param: in url query
		tagValue := f.Tag.Get(config.TAG_url_injection)
		if tagValue != "" {
			if tagValue == "." {
				tagValue = f.Name
			}
			if v, ok := queries[tagValue]; ok {
				values[f.Name] = v

				if logger.Debug() {
					logger.Printf("Inject url param: '%s' <= '%s'\n", f.Name, v)
				}
				proton.SetInjected(f.Name, true)
				continue
			}
		}

	}
	InjectValues(proton, values)
}

func (lcc *LifeCircleControl) injectComponentParameters(params []interface{}) *LifeCircleControl {
	lcc.injectComponentParametersTo(lcc.current.proton, params)
	return lcc
}

// Inject parameters into component, including informalParameters.
func (lcc *LifeCircleControl) injectComponentParametersTo(proton core.Protoner, params []interface{}) {
	// log.Printf("-621- Component [%v]'s params is: ", seg.Name)
	debugprint := false
	if debugprint {
		for i, p := range params {
			fmt.Printf("\t%3v: %v\n", i, p)
		}
		fmt.Println("\t~ END Params ~")
	}

	component, ok := proton.(core.Componenter)
	if !ok {
		panic("Proton is not Component when injectComponentPatameters.")
	}
	componentStructInfo := scache.GetCreate(reflect.TypeOf(proton), core.COMPONENT)

	data := make(map[string][]string) // used to inject with gorilla/schema
	var key string                    // key is also field name.
	for i, param := range params {
		if i%2 == 0 { // it's key
			if k, ok := param.(string); ok {
				if k == "" {
					panic("Component Parameter Key can't be empty!")
				}
				// After template translate, key is lowercased, here I Capitalize the
				// first letter to make match key with component field for injection.
				// TODO Make lowercased key map in structInfo to solve the loss of Cap.
				key = fmt.Sprintf("%v%v", strings.ToUpper(k[0:1]), k[1:]) // Capitalized
				proton.SetInjected(key, true)
			} else {
				panic("component parameter must be name,value pair.")
			}
		} else { // value
			if param == nil {
				fmt.Println("ERROR: value is nil", key, param)
				continue
			}

			// if it's informal parameters, put save them into informal parameters.
			field := componentStructInfo.FieldInfo(key)
			if nil == field { // it's informal parameters
				component.AddInformalParameter(key, param)
				continue
			}
			// fmt.Println("-------------------------------------------------------------------------------")
			// fmt.Println("-------------------------------------------------------------------------------")
			// fmt.Println("-------------------------------------------------------------------------------")
			// fmt.Println("---------- Field is ", field, " [", key, "]. ")
			// fmt.Println("field's type is : ", field.Type)
			// if field.Type == reflect.TypeOf(time.Time) {

			// }

			// value, then set key to struct.
			switch param.(type) {
			case string:
				// if type is string, use gorilla/schema to inject.
				data[key] = []string{param.(string)}

				if logger.Debug() {
					logger.Printf("Inject Component param: '%s' <= '%s'\n", key, data[key])
				}
			default: // other situation
				v := utils.GetRootValue(proton)
				if logger.Debug() {
					logger.Printf("Inject Component param: '%s' <= '%s'\n", key, param)
				}
				injectField(v, key, param) // TO be continued....
			}

		}
	}
	if debugprint {
		debug.PrintFormMap("~ Component ~ inject component data", data)
	}
	InjectValues(proton, data)
}

//
// inject utils
//

func injectField(target reflect.Value, fieldName string, value interface{}) {
	// target
	t := target
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// field
	field := t.FieldByName(fieldName)
	if !field.IsValid() {
		fmt.Sprintf("Inject Error: Can't set '%v' to %v's %v field, type:'%v'.",
			reflect.TypeOf(value), target.Kind(), fieldName, field.Kind())
		panic(fmt.Sprintf("Fields '%s' not found.", fieldName))
	}
	v := reflect.ValueOf(value)

	if field.Kind() != reflect.Interface &&
		field.Kind() != reflect.Ptr &&
		v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t.FieldByName(fieldName).Set(v)
}

// Inject hidden things, i.e. page, components in page.
// TODO need a better name.
func (lcc *LifeCircleControl) injectHiddenThings() *LifeCircleControl {
	lcc.injectHiddenThingsTo(lcc.current.proton)
	return lcc
}

// Inject hidden things, i.e. page, components in page.
// now: Support inject page.
func (lcc *LifeCircleControl) injectHiddenThingsTo(proton core.Protoner) {
	t := utils.GetRootType(proton)

	// fmt.Println("\n________________________________________________________________________________")
	// fmt.Println("----- DEBUG inject page.--------------------------------------------------------")
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		t := f.Type
		if f.Type.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		// injection
		tagInfo := getTagInfo(f.Tag, "inject")
		if tagInfo != nil {
			switch tagInfo.TagValue {
			case "page":
				if pageObj := CreatePage(lcc.w, lcc.r, t); pageObj != nil {
					page := pageObj.(core.Pager)
					// page.SetInjected(f.Name, true)
					lcc.current.proton.SetInjected(f.Name, true)
					v := utils.GetRootValue(proton)
					fieldValue := v.FieldByName(f.Name)
					fieldValue.Set(reflect.ValueOf(page)) // set new page into value.
				} else {
					panic(fmt.Sprintf("Can't find registry for type: %s", t))
				}
			default:
				panic(fmt.Sprintf("Inject type not supported! %v", tagInfo.TagValue))
			}
		}
	}
}

// inject value into object using gorilla.
func InjectValues(proton core.Protoner, data map[string][]string) {
	if nil == data || len(data) == 0 {
		return
	}
	// if si := scache.GetStructInfo(reflect.TypeOf(proton)); si != nil {
	// 	if fi := si.FieldInfo(proton); fi != nil {
	// 		fmt.Println(fi)
	// 	}
	// }
	// TODO just a temp fix

	if _, ok := data["t:id"]; ok {
		delete(data, "t:id")
	}
	// fmt.Printf("converting %v into %v\n", data, proton)
	if err := coercion.SchemaDecoder.Decode(proton, data); err != nil {
		if multierr, ok := err.(schema.MultiError); ok {
			fmt.Println("\n\n\n============== This is an MultiError =====================")
			for errkey, e := range multierr {
				fmt.Printf("> %s : %v\n", errkey, e)
			}
			panic(multierr)
		} else {
			panic(err) // TODO:  more specific users.
		}
	}
}

// --------------------------------------------------------------------------------
func getTagInfo(tag reflect.StructTag, tagName string) *Tag {
	tagValue := tag.Get(tagName)
	if tagValue == "" {
		return nil
	}
	tagValue = strings.TrimSpace(tagValue)
	tagValues := strings.Split(tagValue, " ")
	taginfo := &Tag{TagName: tagName}
	if tagValues != nil && len(tagValues) > 0 {
		taginfo.TagValue = tagValues[0]
		if len(tagValues) > 1 {
			taginfo.Feature = tagValues[1:]
		}
	}
	return taginfo
}

type Tag struct {
	TagName  string
	TagValue string
	Feature  []string
}
