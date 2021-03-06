/*
TODO:
   - Make an util to generate file and run it.
*/

package parser

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"text/template"

	"github.com/elivoa/got/config"
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/debug"
	"github.com/elivoa/got/utils/revex"
)

var importErrorPattern = regexp.MustCompile("cannot find package \"([^\"]+)\"")

// Build the app:
// 1. Generate the the main.go file.
// 2. Run the appropriate "go build" command.
// Requires that revex.Init has been called previously.
// Returns the path to the built binary, and an error if there was a problem building it.
func HackSource(modules []*core.Module) (app *App, compileError *Error) {
	if len(modules) == 0 {
		panic("Generating Error: No modules found!!!")
	}

	// First, clear the generated files (to avoid them messing with ProcessSource).
	cleanSource("generated")

	sourceInfo, compileError := ParseSource(modules, true) // find only = true
	if compileError != nil {
		return nil, compileError
	}

	// set it to cache

	// // Add the db.import to the import paths.
	// if dbImportPath, found := revex.Config.String("db.import"); found {
	// 	sourceInfo.InitImportPaths = append(sourceInfo.InitImportPaths, dbImportPath)
	// }
	importPaths := make(map[string]string)
	// pageSpecs := Code path1 is not in GOPATH[]*StructInfo{}
	typeArrays := [][]*StructInfo{sourceInfo.Structs}
	for _, specs := range typeArrays {
		for _, spec := range specs {
			switch spec.ProtonKind {
			case core.PAGE, core.COMPONENT, core.MIXIN:
				addAlias(importPaths, spec.ImportPath, spec.PackageName)
			}
		}
	}

	// Used by register.RegisterModule(
	str_modules := [][]string{}
	for _, module := range modules {
		// Note: github.com/elivoa/got/builtin -> github.com/elivoa/got/builtin builtin
		_, n := path.Split(module.PackageName)
		alias := addAlias(importPaths, module.PackageName, n)
		str_modules = append(str_modules, []string{alias, module.VarName})
	}

	// Generate two source files.
	data := map[string]interface{}{
		// "Controllers":    sourceInfo.ControllerSpecs(), // empty, leave it there
		// "ValidationKeys": sourceInfo.ValidationKeys,    // empty, levae it there
		// "ModulePaths": modulePaths,
		"ImportPaths": importPaths,
		"Structs":     sourceInfo.Structs,
		"Modules":     modules,
		"STRModules":  str_modules,
		"ProtonKindLabel": map[core.Kind]string{
			core.PAGE:      "core.PAGE",
			core.COMPONENT: "core.COMPONENT",
			core.MIXIN:     "core.MIXIN",
		},
		// "TestSuites":     sourceInfo.TestSuites(),
	}
	genSource("generated", "main.go", MAIN, data)

	// // Read build config.
	// buildTags := revex.Config.StringDefault("build.tags", "")

	// Build the user program (all code under app).
	// It relies on the user having "go" installed.
	goPath, err := exec.LookPath("go")
	if err != nil {
		debug.Log("Go executable not found in PATH.")
	}

	// Loads
	// packagePath := config.Config.StartupModule.PackagePath
	// pkg, err := build.Default.Import(packagePath, "", build.FindOnly)
	// if err != nil {
	// 	debug.Log("Failure importing %v", revex.ImportPath)
	// }

	// binNamex := path.Join(pkg.BinDir, path.Base(config.Config.AppBasePath))
	// this will generate: /Users/bogao/develop/go/src/

	binName := path.Join(config.Config.StartupModule.Path(), "generated", "main")

	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	gotten := make(map[string]struct{})
	for {
		fmt.Println("!> *******************************************************************")
		fmt.Println(goPath, "build", "-o", binName, path.Join(config.Config.StartupModule.PackageName, "generated"))
		buildCmd := exec.Command(goPath, "build",
			// "-tags", buildTags,
			"-o", binName, path.Join(config.Config.StartupModule.PackageName, "generated"))

		// {
		// 	fmt.Println("\n ========== Run Command build =======")

		// 	fmt.Println("> args: ")
		// 	fmt.Println("goPath", goPath)
		// 	fmt.Println("binName", binName)

		// 	fmt.Println("> Run Command: ==========================   ", buildCmd)
		// 	fmt.Println("Command: Details:")
		// 	fmt.Println("Command: Path: ", buildCmd.Path)
		// 	fmt.Println("Command: Args: ", buildCmd.Args)
		// 	fmt.Println("Command: Env: ", buildCmd.Env)
		// 	fmt.Println("Command: Dir: ", buildCmd.Dir)
		// }

		// This step will generate xxx/bin/go
		output, err := buildCmd.CombinedOutput()
		fmt.Println("Output is: ", string(output))

		// If the build succeeded, we're done.
		if err == nil {
			fmt.Println("!> Generator.build success. binName is ", binName)
			fmt.Println("!TODO> Fatal: The location of binName should be optimized. same application in same source folder may be conflict.")
			return NewApp(binName), nil
		}

		// On error goes here!
		fmt.Println("\n===== Error Occured when Building main.go ================")
		fmt.Println(err.Error())

		// -============================ return
		if true {
			panic(string(output))
		}

		// --------------- What does the following done ----------------------?
		{ // -- not reachable --
			// See if it was an import error that we can go get.
			matches := importErrorPattern.FindStringSubmatch(string(output))
			if matches == nil {
				return nil, newCompileError(output)
			}

			// Ensure we haven't already tried to go get it.
			pkgName := matches[1]
			if _, alreadyTried := gotten[pkgName]; alreadyTried {
				return nil, newCompileError(output)
			}
			gotten[pkgName] = struct{}{}

			// Execute "go get <pkg>"
			getCmd := exec.Command(goPath, "get", pkgName)
			// debug.Log("Exec: ", getCmd.Args)
			getOutput, err := getCmd.CombinedOutput()
			if err != nil {
				panic(string(getOutput))
				// revex.ERROR.Println(string(getOutput))
				return nil, newCompileError(output)
			}
		}
		// Success getting the import, attempt to build again.
	}

	panic("Not reachable")
	return nil, nil
}

/// add by elivoa
func calcImports(src *SourceInfo) map[string]string {
	aliases := make(map[string]string)
	typeArrays := [][]*StructInfo{src.Structs /*, src.TestSuites()*/}
	for _, specs := range typeArrays {
		for _, spec := range specs {
			// fmt.Println("  > i:", spec.ImportPath, " o:", spec.PackageName)
			switch spec.ProtonKind {
			case core.PAGE, core.COMPONENT, core.MIXIN:
				addAlias(aliases, spec.ImportPath, spec.PackageName)
			}

			//# method imports, don't import this.
			//#q
			// for _, methSpec := range spec.MethodSpecs {
			// 	for _, methArg := range methSpec.Args {
			// 		if methArg.ImportPath == "" {
			// 			continue
			// 		}

			// 		addAlias(aliases, methArg.ImportPath, methArg.TypeExpr.PkgName)
			// 	}
			// }
		}
	}

	// Add the "InitImportPaths", with alias "_"
	// for _, importPath := range src.InitImportPaths {
	// 	if _, ok := aliases[importPath]; !ok {
	// 		aliases[importPath] = "_"
	// 	}
	// }

	return aliases
}

////@ cleared
func cleanSource(dirs ...string) {
	for _, dir := range dirs {
		tmpPath := path.Join(
			config.Config.SrcPath,
			// config.Config.StartupModule.PackagePath,
			dir)
		fmt.Printf("!> Generator: Remove Folder %s\n", tmpPath)
		err := os.RemoveAll(tmpPath)
		if err != nil {
			revex.ERROR.Println("!> [ERROR in Generator]: Failed to remove dir:", err)
		}
	}
}

//// @cleaned
// genSource renders the given template to produce source code, which it writes
// to the given directory and file.
func genSource(dir, filename, templateSource string, args map[string]interface{}) {
	// generate source
	sourceCode := ExecuteTemplate(
		template.Must(template.New("").Parse(templateSource)),
		args)

	// Create a fresh dir.
	tmpPath := path.Join(
		config.Config.SrcPath,
		// config.Config.StartupModule.PackagePath,
		dir)

	fmt.Printf("Generator :> Generating '%s/main.go'\n", tmpPath)

	err := os.RemoveAll(tmpPath)
	if err != nil {
		revex.ERROR.Println("Failed to remove dir:", err)
	}
	err = os.Mkdir(tmpPath, 0777)
	if err != nil {
		revex.ERROR.Fatalf("Failed to make tmp directory: %v", err)
	}

	// Create the file
	file, err := os.Create(path.Join(tmpPath, filename))
	defer file.Close()
	if err != nil {
		revex.ERROR.Fatalf("Failed to create file: %v", err)
	}
	_, err = file.WriteString(sourceCode)
	if err != nil {
		revex.ERROR.Fatalf("Failed to write to file: %v", err)
	}
}

// Execute a template and returns the result as a string.
func ExecuteTemplate(tmpl revex.ExecutableTemplate, data interface{}) string {
	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		panic(err.Error())
	}
	return b.String()
}

// Looks through all the method args and returns a set of unique import paths
// that cover all the method arg types.
// Additionally, assign package aliases when necessary to resolve ambiguity.
// func calcImportAliases(src *SourceInfo) map[string]string {
// 	aliases := make(map[string]string)
// 	typeArrays := [][]*StructInfo{src.ControllerSpecs() /*, src.TestSuites()*/}
// 	for _, specs := range typeArrays {
// 		for _, spec := range specs {
// 			addAlias(aliases, spec.ImportPath, spec.PackageName)

// 			for _, methSpec := range spec.MethodSpecs {
// 				for _, methArg := range methSpec.Args {
// 					if methArg.ImportPath == "" {
// 						continue
// 					}

// 					addAlias(aliases, methArg.ImportPath, methArg.TypeExpr.PkgName)
// 				}
// 			}
// 		}
// 	}

// 	// Add the "InitImportPaths", with alias "_"
// 	for _, importPath := range src.InitImportPaths {
// 		if _, ok := aliases[importPath]; !ok {
// 			aliases[importPath] = "_"
// 		}
// 	}

// 	return aliases
// }

func addAlias(aliases map[string]string, importPath, pkgName string) string {
	alias, ok := aliases[importPath]
	if ok {
		return alias
	}
	alias = makePackageAlias(aliases, pkgName)
	aliases[importPath] = alias
	return alias
}

func makePackageAlias(aliases map[string]string, pkgName string) string {
	i := 0
	alias := pkgName
	for containsValue(aliases, alias) {
		alias = fmt.Sprintf("%s%d", pkgName, i)
		i++
	}
	return alias
}

func containsValue(m map[string]string, val string) bool {
	for _, v := range m {
		if v == val {
			return true
		}
	}
	return false
}

// Parse the output of the "go build" command.
// Return a detailed Error.
func newCompileError(output []byte) *Error {
	errorMatch := regexp.MustCompile(`(?m)^([^:#]+):(\d+):(\d+:)? (.*)$`).
		FindSubmatch(output)
	if errorMatch == nil {
		revex.ERROR.Println("Failed to parse build errors:\n", string(output))
		return &Error{
			SourceType:  "Go code",
			Title:       "Go Compilation Error",
			Description: "See console for build error.",
		}
	}

	// Read the source for the offending file.
	var (
		relFilename    = string(errorMatch[1]) // e.g. "src/revex/sample/app/controllers/app.go"
		absFilename, _ = filepath.Abs(relFilename)
		line, _        = strconv.Atoi(string(errorMatch[2]))
		description    = string(errorMatch[4])
		compileError   = &Error{
			SourceType:  "Go code",
			Title:       "Go Compilation Error",
			Path:        relFilename,
			Description: description,
			Line:        line,
		}
	)

	fileStr, err := revex.ReadLines(absFilename)
	if err != nil {
		compileError.MetaError = absFilename + ": " + err.Error()
		revex.ERROR.Println(compileError.MetaError)
		return compileError
	}

	compileError.SourceLines = fileStr
	return compileError
}

const MAIN = `// DO NOT EDIT THIS FILE -- GENERATED CODE
package main

import (
    "github.com/elivoa/got/config"
    "github.com/elivoa/got/parser"
    "github.com/elivoa/got/route"
    "fmt"
    _got "github.com/elivoa/got"
    "github.com/elivoa/got/register"
	"github.com/elivoa/got/cache"

	{{range $k, $v := $.ImportPaths}}
    {{$v}} "{{$k}}"{{end}}
)

func main() {
    fmt.Println("\n=============== STARTING GENERATED CODE ================================================")

    // setup config.ModulePath
    {{range .STRModules}}
    config.Config.RegisterModule({{index . 0}}.{{index . 1}}){{end}}

    // parse source again.
    sourceInfo, compileError := parser.ParseSource(config.Config.Modules, false) // deep parse
    if compileError != nil {
        panic(compileError.Error())
    }

    // The first cache is runtime system's cache.
    // Important things are put into sourceInfo.
    // TODO: Put everything into SourceCache
    cache.SourceCache = sourceInfo

    // register real module
    register.RegisterModule({{range .STRModules}}{{index . 0}}.{{index . 1}},{{end}})

    // register pages & components
    {{range .Structs}}{{if .IsProton}}
    route.RegisterProton("{{.ImportPath}}", "{{.StructName}}", "{{.ModulePackage}}", &{{index $.ImportPaths .ImportPath}}.{{.StructName}}{}){{end}}{{end}}

    // start the server
    _got.Start()
}
`

// // cache StructCache
// if !findOnly {
// 	fmt.Println("********************************************************************************")
// 	for _, si := range srcInfo.Structs {
// 		// cache.StructCache.GetCreate(si.IsProton)
// 		fmt.Println(si)
// 	}
// }
