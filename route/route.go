package route

import (
	"fmt"
	"github.com/elivoa/got/config"
	"github.com/elivoa/got/errorhandler"
	"github.com/elivoa/got/route/exit"
	"github.com/gorilla/context"
	"got/cache"
	"got/core"
	"got/core/lifecircle"
	"got/debug"
	"got/register"
	"got/templates"
	"net/http"
	"reflect"
	"strings"
)

var (
	emptyParameters = []reflect.Value{}
	debugLog        = true
)

/*

*/
// RouteHandler is responsible to handler all got request.
func RouteHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path

	// 1. skip special resources. TODO Expand to config.
	// TODO better this
	if url == "/favicon.ico" {
		return
	}

	printAccessHeader(r)

	// --------  Error Handling  --------------------------------------------------------------
	defer func() {
		if err := recover(); err != nil {

			fmt.Println("\n----  MEET SOME ERROR   --------------------------------------------------------")
			fmt.Println(err)

			// if panic occured, turn to errorhandler.
			result := errorhandler.Process(err)
			if nil != result {
				// get current lcc object from request.
				lcc_obj := context.Get(r, config.LCC_OBJECT_KEY)
				lcc := lcc_obj.(*lifecircle.LifeCircleControl)
				if lcc != nil {
					// TODO: what if non-break-returns.
					lcc.HandleExternalReturn(result)
				}
				// fmt.Println(lcc)
				// TODO: continued here....
			}
			/*
				// TODO use a page to render error.
				processPanic(err, r)
				// TODO Render error page 500/404 page.
				http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			*/
		}

		// clear request scope data store.
		context.Clear(r)

		printAccessFooter(r)
	}()
	// --------  Routing...  --------------------------------------------------------------

	// 3. let's find the right pages.
	result := lookup(url)
	if result == nil && !result.IsValid() {
		panic("Page not found! for: " + url)
		// TODO goto 404 page.
		// TODO all path-parameter parse error goes to 404, not 500.
	}
	debug.Log("-601- [RouteFind] %v", result.Segment)

	// TODO: Create New page object every request? howto share some page object? see tapestry5.
	lcc := lifecircle.NewPageFlow(w, r, result.Segment)
	lcc.SetPageUrl(result.PageUrl)
	lcc.SetEventName(result.EventName)

	// print some information.
	defer func() {
		fmt.Println("----------------------------")
		fmt.Println("Describe the page structure:")
		fmt.Println(lcc.PrintCallStructure())
		// ------------------------------------------
	}()

	if !result.IsEventCall() {
		// page render flow
		lcc.PageFlow()
		// handleReturn(lcc, result.Segment)
	} else {
		// event call
		lcc.EventCall(result.EventName)

		// TODO wrong here. this is wrong. sudo refactor lifecircle-return.
		// if lcc not returned, return the current page.
		if lcc.Err != nil {
			panic(lcc.Err.Error())
		}
		// // default return the current page.
		// if result.Segment != nil && lcc.ResultType == "" {
		// 	url := lcc.r.URL.Path
		// 	http.Redirect(lcc.w, lcc.r, url, http.StatusFound)
		// }
	}
}

// Look up page. Can detect event call.
func lookup(url string) *register.LookupResult {
	result, err := register.Pages.Lookup(url)
	if nil != err {
		panic(err.Error())
	}
	if result == nil || result.Segment == nil {
		panic(fmt.Sprintf("Error: seg.Proton is null. seg: %v", result.Segment))
	}
	if result.Segment.Proton == nil {
		// TODO redirect to 404 page.
		panic(fmt.Sprintf("~~~~ Page not found ~~~~"))
	}
	return result
}

// // TODO kill this
// // handle return
// func handleReturn(lcc *lifecircle.LifeCircleControl, seg *register.ProtonSegment) {
// 	// no error, no templates return or redirect.
// 	if seg != nil && lcc.Err == nil && lcc.ResultType == "" {
// 		// find default tempalte to return
// 		identity, templatePath := seg.TemplatePath()
// 		// debug.Log("-755- [TemplateSelect] %v -> %v", identity, templatePath)
// 		if _, err := templates.GotTemplateCache.Get(identity, templatePath); err != nil {
// 			lcc.Err = err
// 		} else {
// 			if err := templates.RenderGotTemplate(lcc.W, identity, lcc.Proton); err != nil {
// 				lcc.Err = err
// 			}
// 		}
// 	}
// 	// panic here if has errors.
// 	if lcc.Err != nil {
// 		panic(lcc.Err.Error())
// 	}
// }

// ----  Register Proton  ----------------------------------------------------------------------------

/*
   RegisterProton register structs to the system.
   Every proton(i.e. pages, components, mixins) should be registered by this fumc.
   Pages and Components registered automatically by 'parser' and 'generator'.
   Example:
      route.RegisterProton("syd/components/layout", "HeaderNav", "syd", &layout.HeaderNav{})
*/
func RegisterProton(pkg string, name string, modulePkg string, proton core.Protoner) {
	si, ok := cache.SourceCache.StructMap[fmt.Sprintf("%v.%v", pkg, name)]
	if !ok {
		panic(fmt.Sprintf("struct info not found: %v.%v ", pkg, name))
	}

	switch proton.Kind() {
	case core.PAGE:
		register.Pages.Add(si, proton)
	case core.COMPONENT:
		selectors := register.Components.Add(si, proton)
		for _, selector := range selectors {
			key := strings.Join(selector, "/")
			lowerKey := strings.ToLower(key)
			templates.RegisterComponentAsFunc(key, lifecircle.ComponentLifeCircle(lowerKey))
		}
	case core.MIXIN:
		fmt.Println("........ [WARRNING...] Mixin not suported now! ", si)
	case core.STRUCT, core.UNKNOWN:
		fmt.Println("........ [Error...] Can't register non proton struct! ", si)
	}
}

// --------------------------------------------------------------------------------
// Exit Helper;;
// These methods are copied to github.com/elivoa/got/route/exit package. To solve circle import.

func TrueExit() *exit.Exit                   { return &exit.Exit{ExitType: "bool", Value: true} }
func FalseExit() *exit.Exit                  { return &exit.Exit{ExitType: "bool", Value: false} }
func Bool(b bool) *exit.Exit                 { return &exit.Exit{ExitType: "bool", Value: b} }
func Redirect(url interface{}) *exit.Exit    { return &exit.Exit{ExitType: "redirect", Value: url} }
func Forward(url interface{}) *exit.Exit     { return &exit.Exit{ExitType: "forward", Value: url} }
func Template(tpl interface{}) *exit.Exit    { return &exit.Exit{ExitType: "template", Value: tpl} }
func RenderText(text interface{}) *exit.Exit { return &exit.Exit{ExitType: "text", Value: text} }
func RenderJson(json interface{}) *exit.Exit { return &exit.Exit{ExitType: "json", Value: json} }
func Error(err interface{}) *exit.Exit       { return &exit.Exit{ExitType: "error", Value: err} }