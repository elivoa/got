/*
Error Handler:
  1. A struct,
  2. Can register some handle methods or some handle redirect.
  3. Has match method to match panic type.
  4. [[[to be continued.]]]

Function Designs:
  . Support Orders, which match first. Defaultly use orders when add. TODO, future improve.


Develop Tips:
  Try this package with none OO thoughts.


*/
package errorhandler

import (
	"errors"
	"fmt"
	pBuiltin "github.com/elivoa/got/builtin/pages"
	"github.com/elivoa/got/core/lifecircle"
	"github.com/elivoa/got/debug"
	"github.com/elivoa/got/route/exit"
	"github.com/elivoa/got/utils"
	"log"
	"net/http"
	"reflect"
	rd "runtime/debug"
	"syd/exceptions"
)

var handlers map[reflect.Type]HandlerPair

func init() {
	handlers = make(map[reflect.Type]HandlerPair, 0) // it's a slice.

	// register handlers
	// TODO should move this to generate?

	// AddHandler("string-panic-handler", reflect.TypeOf(""), Handle500)
	AddHandler("page not found", reflect.TypeOf(exceptions.PageNotFoundError{}), Handle404)
	AddHandler("access denied", reflect.TypeOf(exceptions.AccessDeniedError{}), HandleAccessDeniedError)
	AddHandler("not login", reflect.TypeOf(exceptions.LoginError{}), RedirectHandler("/account/login"))
	// register errors.errorString; this is common error
	AddHandler("error handler", reflect.TypeOf(errors.New("TEST")).Elem(), Handle500)
}

// AddHandler add a new handler for some kind of error/panic.
// Function f returns an interface{} that are treated as GOT returns.
func AddHandler(name string, errType reflect.Type,
	f func(w http.ResponseWriter, r *http.Request, err interface{}) *exit.Exit) error {
	handlers[errType] = HandlerPair{
		name:    name,
		errType: errType,
		handler: f,
	}
	return nil
}

// Process match the error and then goto the right place.
// TODO: return what
func Process(w http.ResponseWriter, r *http.Request, err interface{}) bool {

	if true { // Debug print
		fmt.Println("\n________________________________________________________________________________")
		fmt.Println("---- DEBUG: ErrorHandler >> Meet An Error --------------------------------------")
		debug.DebugPrintVariable(err)
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - --")
		fmt.Println("")
	}

	t := utils.GetRootType(err)
	if handlerPair, ok := handlers[t]; ok {
		// Handler found, process
		result := handlerPair.handler(w, r, err)
		if nil != result {
			// get current lcc object from request.
			if lcc, ok := lifecircle.CurrentLifecircleControl(r); ok {
				lcc.HandleExternalReturn(result)
				return true
			}
		}
		return false
	} else {
		// common error
		// TODO: change this into environment settings.
		if true {
			fmt.Println("\n________________________________________________________________________________")
			fmt.Println("----  ErrorHandler not found, panic directly for debug use ---------------------")
			fmt.Println("----  What is in Handlers? ")
			for k, v := range handlers {
				fmt.Println("       ", k, "  -->  ", v)
			}
			fmt.Println("----  What is Expected Error?")
		}
		panic(err)
		// return Handle500()
	}
}

// struct
type HandlerPair struct {
	name    string // name, where this uses.
	errType reflect.Type
	// will be the same as normal return interface{}
	handler func(w http.ResponseWriter, r *http.Request, err interface{}) *exit.Exit
}

// --------------------------------------------------------------------------------
// ---- Some built-in handlers --------
// With limited right of returns.
// --------------------------------------------------------------------------------

func Handle404(w http.ResponseWriter, r *http.Request, err interface{}) *exit.Exit {
	pageObj := lifecircle.CreatePage(w, r, reflect.TypeOf(pBuiltin.Error404{}))
	if pageObj != nil {
		if page, ok := pageObj.(*pBuiltin.Error404); ok {
			page.Error = err
			return exit.Forward(page)
		}
	}
	return exit.Redirect("/error404")
}

func Handle500(w http.ResponseWriter, r *http.Request, err interface{}) *exit.Exit {
	fmt.Println(`
ERRORER     ERROR     ERROR
ER         ER   OR   ER   OR
ERR       ER     ER ER     OR
  ERROR   ER     ER ER     OR
      ER  ER     ER ER     OR
ERR   ER   ER   OR   RO   OR
  ERROR     ERROR     ORERR
`)
	fmt.Println("500 Error Page: error is")
	printError(err)

	pageObj := lifecircle.CreatePage(w, r, reflect.TypeOf(pBuiltin.Error500{}))
	if pageObj != nil {
		if page, ok := pageObj.(*pBuiltin.Error500); ok {
			page.Error = err
			return exit.Forward(page)
		}
	}
	return exit.Redirect("/error500")
}

func printError(err interface{}) {
	if e, ok := err.(error); ok {
		debug.Error(e)
	}
	fmt.Printf("Error is %v\n", err)
}

func HandleAccessDeniedError(w http.ResponseWriter, r *http.Request, err interface{}) *exit.Exit {
	pageObj := lifecircle.CreatePage(w, r, reflect.TypeOf(pBuiltin.PermissionDenied{}))
	if pageObj != nil {
		if page, ok := pageObj.(*pBuiltin.PermissionDenied); ok {
			page.Error = err
			return exit.Forward(page)
		}
	}
	return exit.Redirect("/permissiondenied")
}

func RedirectHandler(url string) func(w http.ResponseWriter, r *http.Request, err interface{}) *exit.Exit {
	return func(w http.ResponseWriter, r *http.Request, err interface{}) *exit.Exit {
		return exit.Redirect(url)
	}
}

// --------------------------------------------------------------------------------
// ----- helper functions --------
// --------------------------------------------------------------------------------

func PrintALLHandlers() {
	// TODO: pritn all error handlers.
	fmt.Println("TODO : Print all error handlers.")
}

// processPanic only print panic info to the stdout.
func processPanic(err interface{}, r *http.Request) {
	log.Print("xxxxxxxx  PANIC  xxxxxxxxxxxxx", yibaix)
	log.Printf("x URL: %-80v x", r.URL.Path)
	log.Printf("x panic: %-80v x", err)
	log.Print("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", yibaix)
	fmt.Println("StackTrace >>")
	rd.PrintStack()
	fmt.Println()
}

var yibaix = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
