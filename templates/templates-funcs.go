/*
Functions used in tempalte.
Time-stamp: <[templates-funcs.go] Elivoa @ Thursday, 2014-09-11 00:15:40>

This is a full list:

datetime   : 2006-01-02 15:04:05
date       : 2006-01-02

*/
package templates

import (
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/coreservice/coercion"
	"github.com/elivoa/got/util"
	"github.com/elivoa/got/utils"
	"github.com/elivoa/gxl"
	"html/template"
	"math"
	"time"
)

// something to register func map;
var funcMapRegister = template.FuncMap{
	// deprecated
	"eq": equas,

	// date & time
	"formattime":         FormatTime,
	"datetime":           DateTime,
	"date":               Date,
	"smartdatetime":      SmartDateTime,
	"smartvaliddatetime": SmartValidDateTime,

	"prettytime":       BeautyTime,
	"prettyday":        gxl.PrettyDay,
	"prettycurrency":   PrettyCurrency,
	"prettycurrency32": PrettyCurrency32,

	"now":       func() time.Time { return time.Now() },
	"validtime": utils.ValidTime,

	// strings
	"truncate": utils.TrimTruncate,

	// system
	"refer":  GetReferUrl, // get page's refer url, usually used to go back.
	"encode": EncodeContext,

	// steal from stackflow
	"attr": func(s string) template.HTMLAttr { return template.HTMLAttr(s) },
	"safe": func(s string) template.HTML { return template.HTML(s) },
}

func RegisterFunc(funcName string, funcValue interface{}) {
	if TemplateInitialized == true {
		panic("Can't call RegisterFunc() after template is initialized.")
	}
	funcMapRegister[funcName] = funcValue
}

// TODO open this to developer to register global functions.
// func registerBuiltinFuncs(t *template.Template) {
// 	// init functions
// 	t.Funcs(funcMapRegister)
// }

// // TODO open this to developer to register global functions.
// func registerBuiltinFuncs(t *template.Template) {
// 	// init functions
// 	t.Funcs(template.FuncMap{
// 		// deprecated
// 		"eq": equas,

// 		"formattime":    FormatTime,
// 		"datetime":      DateTime,
// 		"date":          Date,
// 		"smartdatetime": SmartDateTime,

// 		"prettytime":     BeautyTime,
// 		"prettyday":      gxl.PrettyDay,
// 		"prettycurrency": PrettyCurrency,

// 		"now":       func() time.Time { return time.Now() },
// 		"validtime": utils.ValidTime,

// 		"encode": EncodeContext,

// 		// steal
// 		"attr": func(s string) template.HTMLAttr {
// 			return template.HTMLAttr(s)
// 		},
// 		"safe": func(s string) template.HTML {
// 			return template.HTML(s)
// 		},
// 	})
// }

/*_______________________________________________________________________________
  Tempalte Functions
*/

func equas(o1 interface{}, o2 interface{}) bool {
	return o1 == o2
}

// {{showtime .CreateTime "2006-01-02 15:04:05"}}
func FormatTime(format string, t time.Time) string {
	return t.Format(format)
}

func BeautyTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func DateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func Date(t time.Time) string {
	return t.Format("2006-01-02")
}

func SmartDateTime(t time.Time) string {
	return coercion.TimeToString(t)
}

// if time is invalid such as 0001-00-00..., return empty
func SmartValidDateTime(t time.Time) string {
	if !utils.ValidTime(t) {
		return ""
	}
	return coercion.TimeToString(t)
}

func PrettyCurrency(d float64) string {
	if math.Mod(d, 1) > 0 {
		return gxl.FormatCurrency(d, 2)
	} else {
		return gxl.FormatCurrency(d, 0)
	}
}

func PrettyCurrency32(d float32) string {
	return PrettyCurrency(float64(d))
}

// c/text ==> c__text
func EncodeContext(s string) string {
	return util.EncodeContext(s)
}

func GetReferUrl(page core.Protoner) string {
	return page.Request().URL.RequestURI()
}
