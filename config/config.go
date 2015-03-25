package config

import (
	"fmt"
	"github.com/elivoa/got/core"
	"path"
	"reflect"
	"time"
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
var Domain string = "syd.com" // TODO goupi

// Life circle related.
var (
	SESSIONID_KEY         = "JSESSIONID"
	SESSION_TIMEOUT       = 3000  // s
	VERIFICATION_CODE_KEY = "_vc" // return page redirect verification key
	// SESSION_USED          = "__SESSION_USED_NEED_JSESSIONID__"

	LCC_OBJECT_KEY    = "__lifecircle_control_key__"
	LCC_REFERER       = "__LCC_REFERER__"
	PAGE_REDIRECT_KEY = "__page_redirect__"

	USER_TOKEN_SESSION_KEY string = "USER_TOKEN_SESSION_KEY"
	TIMEZONE_SESSION_KEY          = "USER_TIMEZONE_KEY"
)

// Framework level configs.
// Warrning: Change this will affact all templates. Don't chagne these.

// TODO need some examples.
var (
	SPLITER_BLOCK            = ":"
	SPLITER_EMBED_COMPONENTS = "."
	SPLITER_EVENT            = ":"
)

var (
	// injection tag keywords:
	TAG_path_injection      = "path-param"
	TAG_url_injection       = "query"
	TAG_page_injection      = "page"
	TAG_component_injection = "component"

	// value injection / value coercion
	IgnoreInjectionParseError bool = true
	ValidTimeFormats               = []string{"2006-01-02", "2006-01-02 15:04:05"}
	DefaultTime                    = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	DefaultTimeReflectValue        = reflect.ValueOf(DefaultTime)
)

// logic configs
var (
	LIST_PAGE_SIZE = 20
)

// if true, check file if modified each time call an template render.
// This will be an performance loss. TODO: Should be monitor file change and reparse if chagne.
var (
	ProductionMode = false // full debug information when debug.
	ReloadTemplate = true
)

// Debug Output Settings;
var (
	ROUTE_PRINT_TIME = true
)
