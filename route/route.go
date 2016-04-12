package route

import (
	"fmt"
	"github.com/elivoa/got/cache"
	"github.com/elivoa/got/config"
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/core/exception"
	"github.com/elivoa/got/core/lifecircle"
	"github.com/elivoa/got/coreservice/sessions"
	"github.com/elivoa/got/debug"
	"github.com/elivoa/got/errorhandler"
	"github.com/elivoa/got/logs"
	"github.com/elivoa/got/register"
	"github.com/elivoa/got/templates"
	"github.com/gorilla/context"
	"net/http"
	"reflect"
	"strings"
	"time"
)

var (
	emptyParameters = []reflect.Value{}
	debugLog        = true
	logRoute        = logs.Get("Router")

	// debug optons
	enable_error_handler = true
)

// RouteHandler is responsible to handler all got request.
func RouteHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path

	// 1. skip special resources. TODO Expand to config. // TODO better this
	if url == "/favicon.ico" {
		return
	}

	var starttime time.Time
	if config.ROUTE_PRINT_TIME {
		fmt.Println(">>> Start Logging time. now is: ", time.Now())
		starttime = time.Now()
	}
	printAccessHeader(r)

	// --------  Error Handling  --------------------------------------------------------------
	defer func() {
		if err := recover(); err != nil {

			if !enable_error_handler {
				if true { // Debug print
					fmt.Println("\n_______________________________________________________________")
					fmt.Println("---- DEBUG: ErrorHandler >> Meet An Error -------------------------")
					// fmt.Println(reflect.TypeOf(err))
					if e, ok := err.(error); ok {
						fmt.Println(debug.StackString(e))
					} else if s, ok := err.(string); ok {
						err = fmt.Errorf(s)
						debug.DebugPrintVariable(err)
					} else {
						debug.DebugPrintVariable(err)
					}
					fmt.Println("-------------------------------------------------------------------")
					fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - --")
					fmt.Println("")
				}

				w.Header().Add("content-type", "text/plain")
				w.Write([]byte(fmt.Sprint("[ErrorHandler can't handler this error!]<br>\n")))
				w.Write([]byte(fmt.Sprint(err)))
				return
			}

			// Error hancler process.

			fmt.Println("\n============= ErrorHandler: Panic Occured. =============")
			// Give control to ErrorHandler if panic occurs.
			if b := errorhandler.Process(w, r, err); b == false {
				// return
				if false { // **** disable this function
					w.Header().Add("content-type", "text/plain")
					w.Write([]byte(fmt.Sprint("[ErrorHandler can't handler this error, it returns false]")))
					w.Write([]byte(fmt.Sprint(err)))
					return
				}
			}

		}

		// clear request scope data store.:: should clear context here? Where to?
		context.Clear(r)

		if config.ROUTE_PRINT_TIME {
			fmt.Println(">>> Request Execution Time: ", time.Now().Sub(starttime))
		}
		printAccessFooter(r)

		// this route is ended here.
	}()
	// --------  Routing...  --------------------------------------------------------------

	// 3. let'sp find the right pages.
	result := lookup(url)
	if logRoute.Trace() {
		logRoute.Printf("Lookup(%s) is:\n%v", url, result)
	}
	if result == nil && !result.IsValid() {
		panic(exception.NewPageNotFoundError(fmt.Sprintf("Page %s not found!", r.URL.Path)))
	}

	// TODO Later: Create New page object every request? howto share some page object? see tapestry5.

	var lcc *lifecircle.LifeCircleControl

	// Check if this is an page request after redirect.
	// if has verification code, this is a redirect page and with some data.
	pageRedirectVerificationKeys, ok := r.URL.Query()[config.VERIFICATION_CODE_KEY]
	if ok && len(pageRedirectVerificationKeys) > 0 {
		fmt.Println("********************************************************************************")
		fmt.Println("********************************************************************************")

		var flash_session_key = config.PAGE_REDIRECT_KEY + pageRedirectVerificationKeys[0]
		sessionId := sessions.SessionId(r, w) // called when needed.
		if targetPageInterface, ok := sessions.GetOk(sessionId, flash_session_key); ok {
			fmt.Println("key is ", flash_session_key)
			fmt.Println("target page interface is ", targetPageInterface)
			if targetPage, ok := targetPageInterface.(core.Pager); ok {
				lcc = lifecircle.NewPageFlowFromExistingPage(w, r, targetPage)
				fmt.Println("successfully get targetpage and continue. TODO:!!!!! here is a memory leak!")

				// remove targetpage from session. OR will memery leak!!
				sessions.Delete(sessionId, flash_session_key)
			}
		}
		fmt.Println("********************************************************************************")
		fmt.Println("********************************************************************************")
	}

	// ************************************************************
	// Normal request page flow, create then flow.
	// ************************************************************
	if lcc == nil {
		lcc = lifecircle.NewPageFlow(w, r, result.Segment)
	}
	lcc.SetParameters(result.Parameters)
	lcc.SetEventName(result.EventName) // ?

	// set lcc to session
	lcc.SetToRequest(config.LCC_OBJECT_KEY, lcc)

	// Done: print some information.
	defer func() {
		fmt.Println("---- [defer] ------------------------")
		fmt.Println("Describe the page structure:")
		fmt.Println(lcc.PrintCallStructure())
		// fmt.Println("-- Page Result is ---------")
		// fmt.Println(result)
	}()

	// Process result & returns.
	if !result.IsEventCall() {
		// page render flow
		lcc.PageFlow()
		// handleReturn(lcc, result.Segment)
	} else {
		// event call
		lcc.EventCall(result)

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

// Lookup call register.Lookup, then process some simple error;
// Lookup can detect event call. Event calls on component.
func lookup(url string /*, referer *lifecircle.LifeCircleControl*/) *register.LookupResult {
	result, err := register.Pages.Lookup(url)
	if nil != err {
		panic(err.Error())
	}
	if result == nil || result.Segment == nil {
		panic(fmt.Sprintf("Error: seg.Proton is null. seg: %v", result.Segment))
	}
	if result.Segment.Proton == nil {
		panic(exception.NewPageNotFoundErrorf("Page Not found for"))
	}
	return result
}

// ----  Register Proton  ----------------------------------------------------------------------------

/*
   RegisterProton register structs to the system.
   Every proton(i.e. pages, components, mixins) should be registered by this func.
   Pages and Components are registered automatically by 'parser' and 'generator'.
   Example:
      route.RegisterProton("syd/components/layout", "HeaderNav", "syd", &layout.HeaderNav{})
*/
func RegisterProton(pkgPrefix string, name string, modulePkg string, proton core.Protoner) {
	si, ok := cache.SourceCache.StructMap[fmt.Sprintf("%v.%v", pkgPrefix, name)]
	if !ok {
		panic(fmt.Sprintf("struct info not found: %v.%v ", pkgPrefix, name))
	}

	switch proton.Kind() {
	case core.PAGE:
		register.Pages.Add(si, proton)
	case core.COMPONENT:
		selectors := register.Components.Add(si, proton)

		// register component as func
		for _, selector := range selectors {
			key := strings.Join(selector, "/") // e.g.: got/TestComponent
			templates.RegisterComponent(key, lifecircle.ComponentLifeCircle(strings.ToLower(key)))
			// templates.RegisterComponentAsFunc(key, lifecircle.ComponentLifeCircle(lowerKey))
		}
	case core.MIXIN:
		panic(fmt.Sprint("........ [WARRNING...] Mixin not suported now! ", si))
	case core.STRUCT, core.UNKNOWN:
		panic(fmt.Sprint("........ [Error...] Can't register non proton struct! ", si))
	}
}
