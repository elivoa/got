package config

import (
	"fmt"
	"got/core"
	"path"
	"reflect"
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

	AppBasePath  string // e.g. /path/to/home     ;; <b>Startup app base path</b>.
	SrcPath      string // e.g. /path/to/home/src (startup module's src-path)
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

	// Database
	DBPort     int // not used
	DBName     string
	DBUser     string
	DBPassword string
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

		// DB
		// DBPort:     3306,
		// DBName:     "syd",
		// DBUser:     "root",
		// DBPassword: "eserver409$)(",
	}
}

// set app base path and other settings.
func (c *Configure) SetBasepath(appBasePath string) {
	fmt.Printf("Config: Set base path to [%v]\n", appBasePath)

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

// TODO automatically get this. // no use
var Domain string = "syd.com"

// Life circle related.
var (
	LCC_OBJECT_KEY = "__lifecircle_control_key__"
	LCC_REFERER    = "__LCC_REFERER__"
)

// framework level configs.
var (
	TAG_path_injection      = "path-param"
	TAG_url_injection       = "query"
	TAG_page_injection      = "page"
	TAG_component_injection = "component"

	SPLITER_BLOCK            = ":"
	SPLITER_EMBED_COMPONENTS = "."
	SPLITER_EVENT            = ":"
)

// if true, check file if modified each time call an template render.
// This will be an performance loss. TODO: Should be monitor file change and reparse if chagne.
var ReloadTemplate = true
