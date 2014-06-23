/*
  Time-stamp: <[got.go] Elivoa @ Tuesday, 2014-06-17 20:16:50>

  TODO:
    - Add Hooks: OnAppStart, AfterAppStart, ...
*/

package got

import (
	"fmt"
	"github.com/elivoa/got/builtin"
	"github.com/elivoa/got/config"
	"github.com/elivoa/got/parser"
	"github.com/elivoa/got/register"
	"github.com/elivoa/got/route"
	"github.com/elivoa/got/utils"
	"got/core"
	"net/http"
)

// build phrase. only set config.
var Config *config.Configure

func init() {
	Config = config.Config
}

// BuildStart generates Start code and run server.
func BuildStart() {

	// register built-in module
	config.Config.RegisterModule(builtin.BuiltinModule)
	// config.Config.RegisterModulePath(builtin.BuiltinModule.Path(), "BuiltinModule", false)

	printRegisteredModulePaths()

	// Generate startup codes.

	// generate proton register sourcecode and compile and run.
	timer := utils.NewTimer()
	fmt.Println("> Generating startup codes...")
	
	app, err := parser.HackSource(Config.Modules)
	if err != nil {
		panic(fmt.Sprintf("build error: %v", err.Error()))
	}
	timer.Log("Generating startup codes Done!")

	// Start the server.
	// TODO: Make my own startup codes.
	app.Port = config.Config.Port

	fmt.Println("TODO: Why port can't be set correctly.   // set app.port to ", config.Config.Port)

	fmt.Println("\n>>>> Run Application")

	appcmd := app.Cmd()
	appcmd.Run() // run and not return

	// 	>>> process goes out here, to generated main.go
}

func printRegisteredModulePaths() {
	// print registered modules.
	fmt.Println("> Registered Module paths:")
	for _, module := range Config.Modules {
		fmt.Printf("    - module: %v.%v\n", module.PackagePath, module.VarName)
	}
}

// <<< called by generated code, start the server.
func Start() {
	welcome()

	// processing modules
	var startupModuleKey string
	var startupModule *core.Module
	for key, module := range register.Modules.Map() {
		// register startup module later.(the last one)
		if module.IsStartupModule {
			startupModuleKey = key
			startupModule = module
			continue
		}

		fmt.Println("!> GOT: Register Module ", key)
		if module.Register != nil {
			module.Register()
		}
	}
	// register startup module
	fmt.Println("!> GOT: Register Startup Module ", startupModuleKey)
	if startupModule.Register != nil {
		startupModule.Register()
	}

	fmt.Println("\n!> GOT:  Register static file paths:\n----------------------------------------")

	// mapping static paths.
	for _, pair := range config.Config.StaticResources {
		fmt.Printf("\t%s -> %s (dir: %s)\n", pair[0], pair[1], http.Dir(pair[1]))
		http.Handle(pair[0],
			http.StripPrefix(pair[0], http.FileServer(http.Dir(pair[1]))),
		)
	}

	// got url matcher
	http.HandleFunc("/", route.RouteHandler)

	fmt.Println(">> got started...")
	http.ListenAndServe(fmt.Sprintf(":%v", Config.Port), nil)
}

// welcome print welcome message to screen.
func welcome() {
	fmt.Println("")
	fmt.Println("``````````````````````````````````````````````````")
	fmt.Println("`  GOT WebFramework     (EARLY BUILD 3)          `")
	fmt.Println("`                                                `")
	fmt.Println("``````````````````````````````````````````````````")
	fmt.Printf("Server Started, Listen localhost:%v\n\n", Config.Port)
	// PrintRegistry()
}

// ________________________________________________________________________________
// Print GOT Evnironment
//
func PrintRegistry() {
	register.Modules.PrintALL()

	fmt.Println("\n---- Pages ---------------------")
	register.Pages.PrintALL()

	fmt.Println("\n---- Components ---------------------")
	register.Components.PrintALL()

	fmt.Println("\n---- Mixins ---------------------")
	fmt.Println("... no mixins avaliable ...")

	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Println()
}
