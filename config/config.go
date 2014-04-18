package config

import (
	"fmt"
	"got/core"
	"path"
)

// ________________________________________________________________________________
// System configs

/*
 * TODO Auto-detect PagePackages
 * ...
 */
var Config = NewConfigure()

type Configure struct {
	Version string // `Framewrok Version`

	AppBasePath  string // e.g. /path/to/home     ;; Startup app base path.
	SrcPath      string // e.g. /path/to/home/src
	StaticPath   string // e.g. /path/to/home/static
	ResourcePath string // e.g. /var/site/data/syd/

	// module path need to import. this is not yours.
	Modules       []*core.Module
	StartupModule *core.Module
	// need BuiltinModule?

	StaticResources [][]string // e.g.: [["/static/", "../"], ...]

	// BasePackages      []string // `Packages that contains Pages and Components etc`
	// PagePackages      []string // `no use`
	// ComponentPackages []string // `...`

	// other system config
	TemplateFileExtension string

	// server
	Port int // start port
}

func NewConfigure() *Configure {
	return &Configure{
		Version:               "0.3.0",
		ResourcePath:          "/tmp/",
		Modules:               []*core.Module{},
		StaticResources:       [][]string{},
		TemplateFileExtension: ".html",

		// BasePackages: []string{"happystroking"},
		// server
		Port: 8080,
	}
}

// set app base path and other settings.
func (c *Configure) SetBasepath(appBasePath string) {
	c.AppBasePath = path.Join(appBasePath, "../")
	c.SrcPath = path.Join(appBasePath, "../", "src")
	c.StaticPath = path.Join(appBasePath, "../", "static")
}

// Register modules
func (c *Configure) RegisterModule(module *core.Module) {
	if module.IsStartupModule {
		if c.StartupModule != nil {
			// panic if not only one startup modules.
			panic(fmt.Sprintln("There are more than one StartupModule, they are: \n  ",
				c.StartupModule.PackagePath, c.StartupModule.VarName, "\n  ",
				module.PackagePath, module.VarName,
			))
		}
		c.StartupModule = module
		c.SetBasepath(module.BasePath)
	}

	Config.Modules = append(Config.Modules, module)

	// fmt.Println("\n____  REGISTER MODULE   ____________________________________________________________")
	// fmt.Println("model.Name = ", module.Name)
	// fmt.Println("model.VarName = ", module.VarName)
	// fmt.Println("model.BasePath = ", module.BasePath)
	// fmt.Println("model.PackagePath = ", module.PackagePath)
	// fmt.Println("model.Description = ", module.Description)
	// fmt.Println("model.IsStartupModule = ", module.IsStartupModule)
	// fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
}

func (c *Configure) AddStaticResource(url string, path string) {
	// TODO warn | after log
	Config.StaticResources = append(Config.StaticResources, []string{url, path})
}

// --------------------------------------------------------------------------------
// Global configs.  /// package level config
// --------------------------------------------------------------------------------

// Life circle related.
var LCC_OBJECT_KEY = "__lifecircle_control_key__"

// TODO automatically get this. // no use
var Domain string = "syd.com"