package route

import (
	"fmt"
	"github.com/elivoa/got/route/exit"
	"log"
	"net/http"
	rd "runtime/debug"
	"strings"
)

// return the first non-empty target.
func RedirectDispatch(targets ...string) *exit.Exit {
	for _, target := range targets {
		if strings.TrimSpace(target) != "" {
			return exit.Redirect(target)
		}
	}
	panic("Can't Dispatch any of these redirects.")
	// return "error", "Can't Dispatch any of these redirects."
}

// --------------------------------------------------------------------------------
// handle components
// --------------------------------------------------------------------------------

// helper
func printAccessHeader(r *http.Request) {
	fmt.Println("\n\n............................................................" +
		"........................................")
	fmt.Println(".. Request.URL.Path = ", r.URL.Path)
	// session := utils.Session(r)
	// fmt.Println(".. Session ID       = ", session.ID)
	fmt.Println("............................................................" +
		"........................................")

	// log.Printf(">>> access %v\n", r.URL.Path)
	// log.Printf("> w is %v\n", reflect.TypeOf(w))
	// log.Printf("> w is %v\n", reflect.TypeOf(req))
}

func printAccessFooter(r *http.Request) {
	//debug.Log("^ ^ ^ ^ ^ ^ ^ ^ PAGE RENDER END ^ ^ ^ ^ ^ ^ ^ ^ ^ ^")
	fmt.Println("-----------------------------^         PAGE RENDER END           " +
		"-----------------------------------")
	fmt.Println("................................................................." +
		"...................................")
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
