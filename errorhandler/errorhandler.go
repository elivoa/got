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
	"fmt"
	"github.com/elivoa/got/route/exit"
	"got/debug"
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
	AddHandler("access denied", reflect.TypeOf(exceptions.AccessDeniedError{}), HandleAccessDeniedError)
}

// AddHandler add a new handler for some kind of error/panic.
// Function f returns an interface{} that are treated as GOT returns.
func AddHandler(name string, errType reflect.Type, f func() *exit.Exit) error {
	handlers[errType] = HandlerPair{
		name:    name,
		errType: errType,
		handler: f,
	}
	return nil
}

// Process match the error and then goto the right place.
func Process(err interface{}) *exit.Exit {

	t := reflect.TypeOf(err)
	if handlerPair, ok := handlers[t]; ok {
		fmt.Println(">>>>>>>>>> ", "enter handler..", handlerPair)
		return handlerPair.handler()
	} else {
		// common error
		debug.DebugPrintVariable(err)
		return Handle500()
	}
}

// struct
type HandlerPair struct {
	name    string // name, where this uses.
	errType reflect.Type
	handler func() *exit.Exit // will be the same as normal return interface{}
}

// --------------------------------------------------------------------------------
// ---- Some built-in handlers --------
// With limited right of returns.
// --------------------------------------------------------------------------------

func Handle404() *exit.Exit {
	// TODO: Pass current url to 404 page, which page is 404.
	fmt.Println("Handle 404")
	// TODO: Return one helper structure. and support it.
	//	return "redirect", "/error404"
	return exit.Redirect("/error404")
}

func Handle500() *exit.Exit {
	// TODO pass error to 500 page, what panic?
	fmt.Println("Handle 500")
	return exit.Redirect("/error500")
}

func HandleAccessDeniedError() *exit.Exit {
	// TODO: show more information in this page.
	return exit.Redirect("/permissiondenied")
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
