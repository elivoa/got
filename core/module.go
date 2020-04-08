package core

import "fmt"

// e.g: for /Users/xxx/src/module/pages/xxx
/*
  Notes:
    {{.PackagePath}}.{{.VarName}} // calling methods

*/
type Module struct {
	Name            string // seems no use.
	Version         string // Application version.
	VarName         string // module name, should be the same with struct name.
	BasePath        string // full file path of this module. e.g.: /Users/xxx/src/
	PackageName     string
	Description     string
	IsStartupModule bool

	// SourcePath      string // full source path = BasePath + SourcePath.
	// PackagePath string // package path. e.g.: /module
	// Register config the module in more details.
	// This method only called by generated code.
	Register func() // manually register page and components.
}

func (m *Module) Key() string {
	return fmt.Sprintf("%s@%s", m.Name, m.Version)
}

// Path returns /User/xxx/src/module
func (m *Module) Path() string {
	return m.BasePath
	// if m.SourcePath != "" {
	// 	return path.Join(m.BasePath, m.SourcePath)
	// } else {
	// 	return path.Join(m.BasePath, m.PackagePath)
	// }
}

func (m *Module) String() string {
	return fmt.Sprintf("Module:%s>%s", m.Key(), m.Path())
}

/* Example */
var ___Example_Module___ Module = Module{
	Name:            "syd",
	Version:         "1.0",
	VarName:         "SYDModule",
	BasePath:        "/Users/bogao/develop/gitme/gotapestry/src",
	PackageName:     "syd",
	Description:     "SYD Selling System Main module.",
	IsStartupModule: true,
}
