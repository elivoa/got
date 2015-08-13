/*
   Time-stamp: <[lifecircle-return.go] Elivoa @ Wednesday, 2015-08-05 23:48:10>
*/
package lifecircle

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elivoa/got/config"
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/coreservice"
	"github.com/elivoa/got/coreservice/sessions"
	"github.com/elivoa/got/debug"
	"github.com/elivoa/got/logs"
	"github.com/elivoa/got/route/exit"
	"github.com/elivoa/got/utils"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

/* ________________________________________________________________________________
   Return
*/
func (lcc *LifeCircleControl) refreshThisPage() *exit.Exit {
	return exit.Redirect(lcc.r.URL.Path)
}

var logReturn = logs.Get("Return")

/*
   Return analysis the return values and redirect to the write place.

   @return: true if is redirect and should return.
            false if continue the flow.

   Now can processing: string, string interface...

   Process the return values:
   If the first value is an error. Redirect to Error page.
   If the only one value is string, confirm this a url and rediredt.
   If has at least 2 values and first parameter is the following string.
      template - template path
      json     - json string / json bytes
      text     - plan/text
      redirect - redirect to url
   @Author elivoa@gmail.com
*/
// event call returns, should be locally
func SmartReturn(returns []reflect.Value) *exit.Exit {

	// debuglog("... - [route.Return] handle return  '%v'", returns)

	// returns nothing equals return true
	if returns == nil || len(returns) == 0 {
		return exit.TrueExit() // returns nothing equels to return true
	}

	first := returns[0]
	kind := first.Kind()

	// dereference interface
	if kind == reflect.Interface {
		first = reflect.ValueOf(first.Interface())
		kind = first.Kind()
	}
	if kind == reflect.Ptr {
		first = first.Elem()
		kind = first.Kind()
	}

	// fmt.Println("returns is", first)

	if kind == reflect.Invalid {
		// when declear returns interface and returns nil. matches here.
		return exit.TrueExit()
	}

	firstObject := first.Interface()
	switch firstObject.(type) {
	case bool:
		return exit.Bool(first.Bool())
	case string:
		// now only support (string, string) pair
		if len(returns) <= 1 {
			// One String mode:: e.g.: redirect::some_text_output
			panic("return string must have the second return value. OneStringMode not implemented.")
		}

		// process the second return value.

		ft := strings.ToLower(first.String())
		switch ft {
		case "text", "json":
			return &exit.Exit{ExitType: ft, Value: returns[1].Interface()}
		case "redirect":
			url, err := extractString(1, returns...)
			if err != nil {
				panic(err)
			}
			return exit.Redirect(url) //  &Returns{"redirect", url}
		default:
			debuglog("[Warrning] return type %v not found!", first.String())
			panic(fmt.Sprintf("[Warrning] return type %v not found!", first.String()))
		}
	case error:
		panic(firstObject.(error))
	case exit.Exit:
		exxx := firstObject.(exit.Exit)
		return &exxx
		// case reflect.Invalid: // invalid means return nil
		// 	return exit.TrueExit() // returnTrue()
	default:
		panic("**** Can't parse return value ****")
	}
}

// HandleBreakReturn means
func (lcc *LifeCircleControl) HandleBreakReturn() {
	r := lcc.returns
	if r == nil {
		return
	}
	switch r.ExitType {

	case "text":
		lcc.return_text("text/plain", r.Value)

	case "json":
		lcc.return_text("text/json", r.Value)

	case "marshaljson":
		jsonbytes, err := json.Marshal(r.Value) // marshal object into json.
		if err != nil {
			panic(err)
		}
		jsonstr := string(jsonbytes)
		lcc.return_text("text/json", jsonstr)

	case "download":
		if pair, ok := r.Value.([]interface{}); ok && len(pair) == 3 {
			mime := pair[0].(string)
			filename := pair[1].(string)
			content := pair[2]

			// now we only return 1 result.
			lcc.w.Header().Add("content-type", mime)
			lcc.w.Header().Set("Content-Disposition", "attachment; filename="+filename)
			lcc.w.Write([]byte(fmt.Sprint(content)))

		} else {
			panic("exit.download format error!")
		}
		// TODO need downloadfile
	case "redirect":
		if str, ok := r.Value.(string); ok {
			http.Redirect(lcc.w, lcc.r, str, http.StatusFound)
			return
		}
		// redirect to page.
		if targetPage, ok := r.Value.(core.Pager); ok {
			if pageflowLogger.Debug() {
				pageflowLogger.Printf("[Redirect] Page redirect to PageInstance. %s", targetPage)
			}
			var verificationCode = rand.Intn(99999) // generate validation code.
			var strVerificationCode = strconv.Itoa(verificationCode)
			url := coreservice.Url.GenerateUrlByPage(targetPage)
			url = coreservice.Url.AppendVerificationCode(url, strVerificationCode) // append code

			// set target page to sesson.flash
			session := sessions.ShortCookieSession(lcc.r)
			var flash_session_key = config.PAGE_REDIRECT_KEY + strVerificationCode
			sessionId := sessions.SessionId(lcc.r, lcc.w)
			sessions.Set(sessionId, flash_session_key, targetPage) // set to session.
			http.Redirect(lcc.w, lcc.r, url, http.StatusFound)     // redirect to page's url

			if pageflowLogger.Debug() {
				fmt.Println(session.Values[flash_session_key])
				fmt.Println(targetPage)
				pageflowLogger.Printf("[Redirect] Set target page into flashes with key: %s",
					flash_session_key)
				pageflowLogger.Printf("[Redirect] Redirect URL is: %s", url)
			}
			return
		}

	case "forward":
		// Now only support forward to page.
		// TODO suppport forward to an URL.

		// forward to InjectedPage,
		if targetPage, ok := r.Value.(core.Pager); ok {
			// Forward to a page object, render this page as pager.
			if pageflowLogger.Debug() {
				pageflowLogger.Printf("[Forward] Page forward to PageInstance. %s", targetPage)
			}
			if pagelife, ok := targetPage.FlowLife().(*Life); ok {
				if pageflowLogger.Trace() {
					pageflowLogger.Printf("[Forward] Page's Life is %s", pagelife)
				}
				// page flow
				newlcc := pagelife.control
				newlcc.isForward = true // disable form submit.
				newlcc.PageFlow()
			} else {
				// error case
				debug.DebugPrintVariable(pagelife)
				panic(">>>> can't find life")
			}

		} else if str, ok := r.Value.(string); ok {
			panic(fmt.Sprintf("Not support forward to string now. %s", str))
		} else {
			panic(fmt.Sprintf("Can't forward to this type, %s.", utils.GetRootValue(r.Value).Kind()))
		}

	case "block":
		// TODO ...

	}
}

// now only used in router, when panic.
func (lcc *LifeCircleControl) HandleExternalReturn(externalExit *exit.Exit) {
	lcc.returns = externalExit
	lcc.HandleBreakReturn()
}

// support second parameter type: string, []byte
func (lcc *LifeCircleControl) return_text(contentType string, value interface{}) {
	// now we only return 1 result.
	lcc.w.Header().Add("content-type", contentType)
	lcc.w.Write([]byte(fmt.Sprint(value)))
	// TODO struct to json
}

// TODO reflect utils.
func extractString(index int, data ...reflect.Value) (string, error) {
	if len(data) <= index {
		err := errors.New(fmt.Sprintf("Need the %v(st/nd) return value.", index))
		return emptyString, err
	}
	if data[index].Kind() == reflect.String {
		return data[index].Interface().(string), nil
	} else {
		err := errors.New(fmt.Sprintf("The %v(st/nd) return value must be string.", index))
		return emptyString, err
	}
}
