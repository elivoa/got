/*
  Parse all page/components source files. Cache it's content.

  TODO:
    - Rename function names.
    - Cleanup this file.
    - Make a lightwight parse, only parse modules with core embed.
*/

package parser

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/scanner"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/elivoa/got/perfs"

	"github.com/elivoa/got/core"
	"github.com/elivoa/got/debug"
	"github.com/elivoa/got/utils"
)

var DEBUG_QUICK = false

// methodCall describes a call to c.Render(..)
// It documents the argument names used, in order to propagate them to RenderArgs.
type methodCall struct {
	Path  string // e.g. "myapp/app/controllers.(*Application).Action"
	Line  int
	Names []string
}

type MethodSpec struct {
	Name        string        // Name of the method, e.g. "Index"
	Args        []*MethodArg  // Argument descriptors
	RenderCalls []*methodCall // Descriptions of Render() invocations from this Method.
}

type MethodArg struct {
	Name       string   // Name of the argument.
	TypeExpr   TypeExpr // The name of the type, e.g. "int", "*pkg.UserType"
	ImportPath string   // If the arg is of an imported type, this is the import path.
}

type embeddedTypeName struct {
	ImportPath, StructName string
}

// Maps a controller simple name (e.g. "Login") to the methods for which it is a
// receiver.
type methodMap map[string][]*MethodSpec

//// @cleaned
// Parse the app controllers directory and return a list of the controller types found.
// Returns a CompileError if the parsing fails.
func ParseSource(modules []*core.Module, findOnly bool) (*SourceInfo, *Error) {
	timer := utils.NewTimer()
	defer timer.Log("Parsing Sources Done.")

	if len(modules) == 0 {
		panic("Generating Error: No modules found!!!")
	}

	var (
		srcInfo      *SourceInfo
		compileError *Error
		// sourcePaths  []string = make([]string, len(modules))
	)

	for _, module := range modules {
		sourcePath := module.BasePath

		fmt.Println("!> Generator # Parse Source Folder: ", sourcePath)
		fmt.Println("!> Generator # Module.modulePackagePath --> ", module.PackageName)

		// }

		// for i := 0; i < len(modules); i++ {
		// 	sourcePaths[i] = modules[i].PackageName
		// }

		// for _, sourcePath := range sourcePaths {
		// modulePackagePath := extractPackagePath(sourcePath)

		// if modulePackagePath == "" {
		// 	debug.Log("Skipping code path %v", sourcePath)
		// 	continue
		// }

		i := 0
		// Start walking the directory tree.
		filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {

			// timer := perfs.NewPerfTimer("Walk")

			i += 1
			if DEBUG_QUICK && i > 10 {
				// debug.Log("！！！！！！！！！！！！Special Skipping code path %v", sourcePath)
				return nil
			}

			// skip tmp folder
			if !info.IsDir() || info.Name() == "tmp" {
				return nil
			}

			if err != nil {
				debug.Log("Error scanning source: %v", err)
				return nil
			}

			// Get the import path of the package.
			modulePackagePath := module.PackageName
			pkgPath := modulePackagePath
			if sourcePath != path {
				pkgPath = modulePackagePath + "/" + filepath.ToSlash(path[len(sourcePath)+1:])
			}

			// timer.LapInfo(fmt.Sprintf("pkg: [%s]", pkgPath))

			// TODO: here we get things like this, deep parse & analisys proton value and cache them.
			//       or even generate new source code.
			/*
				module package path is:  got/builtin
				pkg path is:  got/builtin
				pkg path is:  got/builtin/components
				pkg path is:  got/builtin/components/fileupload
				pkg path is:  got/builtin/pages
				pkg path is:  got/builtin/pages/got
				pkg path is:  got/builtin/pages/got/fileupload
			*/

			// Parse files within the path.
			var pkgs map[string]*ast.Package
			fset := token.NewFileSet()

			pkgs, err = parser.ParseDir(fset, path, func(f os.FileInfo) bool {
				return !f.IsDir() && !strings.HasPrefix(f.Name(), ".") && strings.HasSuffix(f.Name(), ".go")
			}, 0)

			// timer.LapInfo("2")

			if err != nil {
				if errList, ok := err.(scanner.ErrorList); ok {
					var pos token.Position = errList[0].Pos

					// return errors.New(fmt.Sprintf("Compilation Error: %v:%v:%v - %v",
					// 	pos.Filename, pos.Line, pos.Column, errList[0].Msg,
					// ))

					compileError = &Error{
						SourceType:  ".go source",
						Title:       "Go Compilation Error",
						Path:        pos.Filename,
						Description: errList[0].Msg,
						Line:        pos.Line,
						Column:      pos.Column,
						// SourceLines: revex.MustReadLines(pos.Filename),
					}
					return compileError
				}
				ast.Print(nil, err)
				// log.Fatalf("Failed to parse dir: %s", err)
				panic(fmt.Sprintf("x> Generator.Parse Error: Failed to parse dir: %s", err))
			}

			// Skip "main" packages.
			delete(pkgs, "main")

			// If there is no code in this directory, skip it.
			if len(pkgs) == 0 {
				return nil
			}

			// There should be only one package in this directory.
			if len(pkgs) > 1 {
				debug.Log("Multiple packages in a single directory: %v", pkgs)
			}

			var pkg *ast.Package
			for _, v := range pkgs {
				pkg = v
			}

			ssi := processPackage(fset, modulePackagePath, pkgPath, path, pkg)

			// timer.LapInfo("4")

			srcInfo = appendSourceInfo(srcInfo, ssi)
			return nil // return walk
		})
	}

	// map it
	tmap := map[string]*StructInfo{}
	for _, ti := range srcInfo.Structs {
		tmap[fmt.Sprintf("%v.%v", ti.ImportPath, ti.StructName)] = ti
	}
	srcInfo.StructMap = tmap
	return srcInfo, compileError
}

func appendSourceInfo(srcInfo1, srcInfo2 *SourceInfo) *SourceInfo {
	if srcInfo1 == nil {
		return srcInfo2
	}

	srcInfo1.Structs = append(srcInfo1.Structs, srcInfo2.Structs...)
	srcInfo1.InitImportPaths = append(srcInfo1.InitImportPaths, srcInfo2.InitImportPaths...)
	for k, v := range srcInfo2.ValidationKeys {
		if _, ok := srcInfo1.ValidationKeys[k]; ok {
			log.Println("Key conflict when scanning validation calls:", k)
			continue
		}
		srcInfo1.ValidationKeys[k] = v
	}
	return srcInfo1
}

// @processing
// pkgImportPath is package path
// pkgPath is full filepath
func processPackage(fset *token.FileSet, modulePackatePath string, pkgImportPath, pkgPath string,
	pkg *ast.Package) *SourceInfo {

	timer := perfs.NewPerfTimer("PACKAGE")

	// fmt.Println("    --------------------------------------------------------")
	// fmt.Printf("    .processPackage: pkgImportPath: %v\n", pkgImportPath)
	relativeModelPath := ""
	if strings.HasPrefix(pkgImportPath, modulePackatePath) {
		relativeModelPath = pkgImportPath[len(modulePackatePath):]
	}
	var (
		structSpecs     []*StructInfo
		initImportPaths []string

		methodSpecs    = make(methodMap)
		validationKeys = make(map[string]map[int]string)

		scanPages = strings.HasSuffix(relativeModelPath, "/pages") ||
			strings.HasPrefix(relativeModelPath, "/pages/")
		scanComponents = strings.HasSuffix(relativeModelPath, "/components") ||
			strings.HasPrefix(relativeModelPath, "/components/")
		scanMixins = strings.HasSuffix(relativeModelPath, "/mixins") ||
			strings.HasPrefix(relativeModelPath, "/mixins/")
	)

	// For each source file in the package...
	for _, file := range pkg.Files {

		// Imports maps the package key to the full import path.
		// e.g. import "sample/app/models" => "models": "sample/app/models"
		imports := map[string]string{}

		// For each declaration in the source file...
		for _, decl := range file.Decls {
			addImports(imports, decl, pkgPath)

			if scanPages || scanComponents || scanMixins {
				// Match and add both structs and methods
				structSpecs = appendStruct(structSpecs, modulePackatePath, pkgImportPath, pkg, decl, imports)
				appendAction(fset, methodSpecs, decl, pkgImportPath, pkg.Name, imports)
				// } else if scanTests {
				// 	structSpecs = appendStruct(structSpecs, pkgImportPath, pkg, decl, imports)
			}

			// If this is a func...
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				// Scan it for validation calls
				lineKeys := getValidationKeys(fset, funcDecl, imports)
				if len(lineKeys) > 0 {
					validationKeys[pkgImportPath+"."+getFuncName(funcDecl)] = lineKeys
				}

				// Check if it's an init function.
				if funcDecl.Name.Name == "init" {
					initImportPaths = []string{pkgImportPath}
				}
			}
		}

		timer.LapInfo(fmt.Sprintf("process [%s]", file.Name))

	}

	// Add the method specs to the struct specs.
	for _, spec := range structSpecs {
		spec.MethodSpecs = methodSpecs[spec.StructName]
	}

	return &SourceInfo{
		Structs:         structSpecs,
		ValidationKeys:  validationKeys,
		InitImportPaths: initImportPaths,
	}
}

// getFuncName returns a name for this func or method declaration.
// e.g. "(*Application).SayHello" for a method, "SayHello" for a func.
func getFuncName(funcDecl *ast.FuncDecl) string {
	prefix := ""
	if funcDecl.Recv != nil {
		recvType := funcDecl.Recv.List[0].Type
		if recvStarType, ok := recvType.(*ast.StarExpr); ok {
			prefix = "(*" + recvStarType.X.(*ast.Ident).Name + ")"
		} else {
			prefix = recvType.(*ast.Ident).Name
		}
		prefix += "."
	}
	return prefix + funcDecl.Name.Name
}

func addImports(imports map[string]string, decl ast.Decl, srcDir string) {
	genDecl, ok := decl.(*ast.GenDecl)
	if !ok {
		return
	}

	if genDecl.Tok != token.IMPORT {
		return
	}

	for _, spec := range genDecl.Specs {
		importSpec := spec.(*ast.ImportSpec)
		var pkgAlias string
		if importSpec.Name != nil {
			pkgAlias = importSpec.Name.Name
			if pkgAlias == "_" {
				continue
			}
		}
		quotedPath := importSpec.Path.Value           // e.g. "\"sample/app/models\""
		fullPath := quotedPath[1 : len(quotedPath)-1] // Remove the quotes

		// If the package was not aliased (common case), we have to import it
		// to see what the package name is.
		// TODO: Can improve performance here a lot:
		// 1. Do not import everything over and over again.  Keep a cache.
		// 2. Exempt the standard library; their directories always match the package name.
		// 3. Can use build.FindOnly and then use parser.ParseDir with mode PackageClauseOnly
		if pkgAlias == "" {
			pkg, err := build.Import(fullPath, srcDir, 0)
			if err != nil {
				// We expect this to happen for apps using reverse routing (since we
				// have not yet generated the routes).  Don't log that.
				if !strings.HasSuffix(fullPath, "/app/routes") {
					debug.Log("Could not find import: %v", fullPath)
				}
				continue
			}
			pkgAlias = pkg.Name
		}

		imports[pkgAlias] = fullPath
	}
}

// If this Decl is a struct type definition, it is summarized and added to specs.
// Else, specs is returned unchanged.
func appendStruct(specs []*StructInfo, modulePackatePath string, pkgImportPath string, pkg *ast.Package, decl ast.Decl, imports map[string]string) []*StructInfo {
	// Filter out non-Struct type declarations.
	spec, found := getStructTypeDecl(decl)
	if !found {
		return specs
	}
	structType := spec.Type.(*ast.StructType)

	// At this point we know it's a type declaration for a struct.
	// Fill in the rest of the info by diving into the fields.
	// Add it provisionally to the Controller list -- it's later filtered using field info.
	protonInfos := &StructInfo{
		StructName:    spec.Name.Name,
		ImportPath:    pkgImportPath,
		PackageName:   pkg.Name,
		ModulePackage: modulePackatePath,
	}

	// TODO cache all field info that i need.

	for _, field := range structType.Fields.List {
		// If field.Names is set, it's not an embedded type.
		if field.Names != nil {
			continue
		}

		// A direct "sub-type" has an ast.Field as either:
		//   Ident { "AppController" }
		//   SelectorExpr { "rev", "Controller" }
		// Additionally, that can be wrapped by StarExprs.
		fieldType := field.Type
		pkgName, typeName := func() (string, string) {
			// Drill through any StarExprs.
			for {
				if starExpr, ok := fieldType.(*ast.StarExpr); ok {
					fieldType = starExpr.X
					continue
				}
				break
			}

			// If the embedded type is in the same package, it's an Ident.
			if ident, ok := fieldType.(*ast.Ident); ok {
				return "", ident.Name
			}

			if selectorExpr, ok := fieldType.(*ast.SelectorExpr); ok {
				if pkgIdent, ok := selectorExpr.X.(*ast.Ident); ok {
					return pkgIdent.Name, selectorExpr.Sel.Name
				}
			}

			return "", ""
		}()

		// if = core.Component is component
		// fmt.Printf("        > pkgName: %v, \ttypeName: %v\n", pkgName, typeName)

		// Set ProtonKind; TODO: HardCoded
		if pkgName == "core" {
			switch typeName {
			case "Page":
				protonInfos.ProtonKind = core.PAGE
			case "Component":
				protonInfos.ProtonKind = core.COMPONENT
			case "Mixin":
				protonInfos.ProtonKind = core.MIXIN
			default:
				protonInfos.ProtonKind = core.STRUCT
			}
		}

		// fmt.Println("    > ", core.KindLabels[protonInfos.ProtonKind], ":", spec.Name.Name)

		// If a typename wasn't found, skip it.
		if typeName == "" {
			continue
		}

		// Find the import path for this type.
		// If it was referenced without a package name, use the current package import path.
		// Else, look up the package's import path by name.
		var importPath string
		if pkgName == "" {
			importPath = pkgImportPath
		} else {
			var ok bool
			if importPath, ok = imports[pkgName]; !ok {
				log.Print("Failed to find import path for ", pkgName, ".", typeName)
				continue
			}
		}

		protonInfos.embeddedTypes = append(protonInfos.embeddedTypes, &embeddedTypeName{
			ImportPath: importPath,
			StructName: typeName,
		})
	}

	return append(specs, protonInfos)
}

////# I canged this method to cache

// If decl is a Method declaration, it is summarized and added to the array
// underneath its receiver type.
// e.g. "Login" => {MethodSpec, MethodSpec, ..}
func appendAction(fset *token.FileSet, mm methodMap, decl ast.Decl, pkgImportPath, pkgName string, imports map[string]string) {
	// Func declaration?
	funcDecl, ok := decl.(*ast.FuncDecl)
	if !ok {
		return
	}

	// Have a receiver?
	if funcDecl.Recv == nil {
		return
	}

	// Is it public?
	if !funcDecl.Name.IsExported() {
		return
	}

	////# commment this because I accept all return values.
	// Does it return a Result?
	// if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) != 1 {
	// 	return
	// }
	// selExpr, ok := funcDecl.Type.Results.List[0].Type.(*ast.SelectorExpr)
	// if !ok {
	// 	return
	// }
	// if selExpr.Sel.Name != "Result" {
	// 	return
	// }
	// if pkgIdent, ok := selExpr.X.(*ast.Ident); !ok || imports[pkgIdent.Name] != revex.REVEX_IMPORT_PATH {
	// 	return
	// }

	method := &MethodSpec{
		Name: funcDecl.Name.Name,
	}

	// Add a description of the arguments to the method.
	for _, field := range funcDecl.Type.Params.List {
		for _, name := range field.Names {
			var importPath string
			typeExpr := NewTypeExpr(pkgName, field.Type)
			if !typeExpr.Valid {
				return // We didn't understand one of the args.  Ignore this action. (Already logged)
			}

			if typeExpr.PkgName != "" {
				var ok bool
				if importPath, ok = imports[typeExpr.PkgName]; !ok {
					log.Println("Failed to find import for arg of type:", typeExpr.TypeName(""))
				}
			}
			method.Args = append(method.Args, &MethodArg{
				Name:       name.Name,
				TypeExpr:   typeExpr,
				ImportPath: importPath,
			})
		}
	}

	// fmt.Println("    func:", funcDecl.Name)
	// for _, a := range method.Args {
	// 	fmt.Println("      > ", a.ImportPath, a.Name, a.TypeExpr)
	// }

	// Add a description of the calls to Render from the method.
	// Inspect every node (e.g. always return true).
	method.RenderCalls = []*methodCall{}
	ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
		// Is it a function call?
		callExpr, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Is it calling (*Controller).Render?
		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// The type of the receiver is not easily available, so just store every
		// call to any method called Render.
		if selExpr.Sel.Name != "Render" {
			return true
		}

		// Add this call's args to the renderArgs.
		pos := fset.Position(callExpr.Rparen)
		methodCall := &methodCall{
			Line:  pos.Line,
			Names: []string{},
		}
		for _, arg := range callExpr.Args {
			argIdent, ok := arg.(*ast.Ident)
			if !ok {
				continue
			}
			methodCall.Names = append(methodCall.Names, argIdent.Name)
		}
		method.RenderCalls = append(method.RenderCalls, methodCall)
		return true
	})

	var recvTypeName string
	var recvType ast.Expr = funcDecl.Recv.List[0].Type
	if recvStarType, ok := recvType.(*ast.StarExpr); ok {
		recvTypeName = recvStarType.X.(*ast.Ident).Name
	} else {
		recvTypeName = recvType.(*ast.Ident).Name
	}

	mm[recvTypeName] = append(mm[recvTypeName], method)
}

// Scan app source code for calls to X.Y(), where X is of type *Validation.
//
// Recognize these scenarios:
// - "Y" = "Validation" and is a member of the receiver.
//   (The common case for inline validation)
// - "X" is passed in to the func as a parameter.
//   (For structs implementing Validated)
//
// The line number to which a validation call is attributed is that of the
// surrounding ExprStmt.  This is so that it matches what runtime.Callers()
// reports.
//
// The end result is that we can set the default validation key for each call to
// be the same as the local variable.
func getValidationKeys(fset *token.FileSet, funcDecl *ast.FuncDecl, imports map[string]string) map[int]string {
	var (
		lineKeys = make(map[int]string)

		// Check the func parameters and the receiver's members for the *revex.Validation type.
		validationParam = getValidationParameter(funcDecl, imports)
	)

	ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
		// e.g. c.Validation.Required(arg) or v.Required(arg)
		callExpr, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		// e.g. c.Validation.Required or v.Required
		funcSelector, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		switch x := funcSelector.X.(type) {
		case *ast.SelectorExpr: // e.g. c.Validation
			if x.Sel.Name != "Validation" {
				return true
			}

		case *ast.Ident: // e.g. v
			if validationParam == nil || x.Obj != validationParam {
				return true
			}

		default:
			return true
		}

		if len(callExpr.Args) == 0 {
			return true
		}

		// Given the validation expression, extract the key.
		key := callExpr.Args[0]
		switch expr := key.(type) {
		case *ast.BinaryExpr:
			// If the argument is a binary expression, take the first expression.
			// (e.g. c.Validation.Required(myName != ""))
			key = expr.X
		case *ast.UnaryExpr:
			// If the argument is a unary expression, drill in.
			// (e.g. c.Validation.Required(!myBool)
			key = expr.X
		case *ast.BasicLit:
			// If it's a literal, skip it.
			return true
		}

		if typeExpr := NewTypeExpr("", key); typeExpr.Valid {
			lineKeys[fset.Position(callExpr.End()).Line] = typeExpr.TypeName("")
		}
		return true
	})

	return lineKeys
}

// Check to see if there is a *revex.Validation as an argument.
func getValidationParameter(funcDecl *ast.FuncDecl, imports map[string]string) *ast.Object {
	for _, field := range funcDecl.Type.Params.List {
		starExpr, ok := field.Type.(*ast.StarExpr) // e.g. *revex.Validation
		if !ok {
			continue
		}

		selExpr, ok := starExpr.X.(*ast.SelectorExpr) // e.g. revex.Validation
		if !ok {
			continue
		}

		xIdent, ok := selExpr.X.(*ast.Ident) // e.g. rev
		if !ok {
			continue
		}

		if selExpr.Sel.Name == "Validation" && imports[xIdent.Name] == "github.com/robfig/revex" {
			// revex.REVEX_IMPORT_PATH { // NOTE nouse by gb
			return field.Names[0].Obj
		}
	}
	return nil
}

func (s *StructInfo) String() string {
	if s == nil {
		return "[StructInfo is NIL]"
	}
	return s.ImportPath + "." + s.StructName
}

func (s *embeddedTypeName) String() string {
	return s.ImportPath + "." + s.StructName
}

// getStructTypeDecl checks if the given decl is a type declaration for a
// struct.  If so, the TypeSpec is returned.
func getStructTypeDecl(decl ast.Decl) (spec *ast.TypeSpec, found bool) {
	genDecl, ok := decl.(*ast.GenDecl)
	if !ok {
		return
	}

	if genDecl.Tok != token.TYPE {
		return
	}

	if len(genDecl.Specs) != 1 {
		debug.Log("Surprising: Decl does not have 1 Spec: %v", genDecl)
		return
	}

	spec = genDecl.Specs[0].(*ast.TypeSpec)
	if _, ok := spec.Type.(*ast.StructType); ok {
		found = true
	}

	return
}

// TypesThatEmbed returns all types that (directly or indirectly) embed the
// target type, which must be a fully qualified type name,
// e.g. "github.com/robfig/revex.Controller"
func (s *SourceInfo) TypesThatEmbed(targetType string) (filtered []*StructInfo) {
	// Do a search in the "embedded type graph", starting with the target type.
	nodeQueue := []string{targetType}
	for len(nodeQueue) > 0 {
		controllerSimpleName := nodeQueue[0]
		nodeQueue = nodeQueue[1:]
		for _, spec := range s.Structs {
			if ContainsString(nodeQueue, spec.String()) {
				continue // Already added
			}

			// Look through the embedded types to see if the current type is among them.
			for _, embeddedType := range spec.embeddedTypes {

				// If so, add this type's simple name to the nodeQueue, and its spec to
				// the filtered list.
				if controllerSimpleName == embeddedType.String() {
					nodeQueue = append(nodeQueue, spec.String())
					filtered = append(filtered, spec)
					break
				}
			}
		}
	}
	return
}

func ContainsString(list []string, target string) bool {
	for _, el := range list {
		if el == target {
			return true
		}
	}
	return false
}

// func (s *SourceInfo) ControllerSpecs() []*StructInfo {
// 	if s.controllerSpecs == nil {
// 		s.controllerSpecs = s.TypesThatEmbed(reveX.REVEX_IMPORT_PATH + ".Controller")
// 	}
// 	return s.controllerSpecs
// }

// func (s *SourceInfo) TestSuites() []*StructInfo {
// 	if s.testSuites == nil {
// 		s.testSuites = s.TypesThatEmbed(reveX.REVEX_IMPORT_PATH + ".TestSuite")
// 	}
// 	return s.testSuites
// }

// TypeExpr provides a type name that may be rewritten to use a package name.
type TypeExpr struct {
	Expr     string // The unqualified type expression, e.g. "[]*MyType"
	PkgName  string // The default package idenifier
	pkgIndex int    // The index where the package identifier should be inserted.
	Valid    bool
}

// TypeName returns the fully-qualified type name for this expression.
// The caller may optionally specify a package name to override the default.
func (e TypeExpr) TypeName(pkgOverride string) string {
	pkgName := FirstNonEmpty(pkgOverride, e.PkgName)
	if pkgName == "" {
		return e.Expr
	}
	return e.Expr[:e.pkgIndex] + pkgName + "." + e.Expr[e.pkgIndex:]
}

func FirstNonEmpty(strs ...string) string {
	for _, str := range strs {
		if len(str) > 0 {
			return str
		}
	}
	return ""
}

// This returns the syntactic expression for referencing this type in Go.
func NewTypeExpr(pkgName string, expr ast.Expr) TypeExpr {
	switch t := expr.(type) {
	case *ast.Ident:
		if IsBuiltinType(t.Name) {
			pkgName = ""
		}
		return TypeExpr{t.Name, pkgName, 0, true}
	case *ast.SelectorExpr:
		e := NewTypeExpr(pkgName, t.X)
		return TypeExpr{t.Sel.Name, e.Expr, 0, e.Valid}
	case *ast.StarExpr:
		e := NewTypeExpr(pkgName, t.X)
		return TypeExpr{"*" + e.Expr, e.PkgName, e.pkgIndex + 1, e.Valid}
	case *ast.ArrayType:
		e := NewTypeExpr(pkgName, t.Elt)
		return TypeExpr{"[]" + e.Expr, e.PkgName, e.pkgIndex + 2, e.Valid}
	case *ast.Ellipsis:
		e := NewTypeExpr(pkgName, t.Elt)
		return TypeExpr{"[]" + e.Expr, e.PkgName, e.pkgIndex + 2, e.Valid}
	default:
		log.Println("Failed to generate name for field.")
		ast.Print(nil, expr)
	}
	return TypeExpr{Valid: false}
}

var _BUILTIN_TYPES = map[string]struct{}{
	"bool":       struct{}{},
	"byte":       struct{}{},
	"complex128": struct{}{},
	"complex64":  struct{}{},
	"error":      struct{}{},
	"float32":    struct{}{},
	"float64":    struct{}{},
	"int":        struct{}{},
	"int16":      struct{}{},
	"int32":      struct{}{},
	"int64":      struct{}{},
	"int8":       struct{}{},
	"rune":       struct{}{},
	"string":     struct{}{},
	"uint":       struct{}{},
	"uint16":     struct{}{},
	"uint32":     struct{}{},
	"uint64":     struct{}{},
	"uint8":      struct{}{},
	"uintptr":    struct{}{},
}

func IsBuiltinType(name string) bool {
	_, ok := _BUILTIN_TYPES[name]
	return ok
}

// TODO 这个是不对的。
// @processed
// func extractPackagePath(path string) string {
// 	fmt.Println(".....", path)

// 	if strings.HasSuffix(path, "github.com/elivoa/got/builtin") {
// 		// path = strings.ReplaceAll(path, "gotapestry/github.com/elivoa/got/builtin", "got/builtin")
// 		// fmt.Println("****************", path)
// 		return "github.com/elivoa/got/builtin"
// 		// return path
// 	}

// 	workPath, _ := os.Getwd()
// 	paths := []string{
// 		workPath,
// 	}
// 	if strings.HasSuffix(workPath, "/gotapestry") {
// 		paths = append(paths, filepath.Join(workPath, "../got"))
// 		paths = append(paths, filepath.Join(workPath, "../syd"))
// 	}
// 	for _, p := range filepath.SplitList(build.Default.GOPATH) {
// 		paths = append(paths, p)
// 	}

// 	for _, gopath := range paths {
// 		fmt.Println("---- ", path, gopath, strings.HasPrefix(path, gopath))

// 		if strings.HasPrefix(path, gopath) {
// 			return filepath.ToSlash(path[len(gopath)+1:])
// 		}

// 		srcPath := filepath.Join(gopath, "src") // 这是不对的

// 		if strings.HasPrefix(path, srcPath) {
// 			return filepath.ToSlash(path[len(srcPath)+1:])
// 		}
// 	}
// 	// ...../gotapestry/src/pkg
// 	srcPath := filepath.Join(build.Default.GOROOT, "src", "pkg")
// 	fmt.Println(".....", path, srcPath, strings.HasPrefix(path, srcPath))

// 	if strings.HasPrefix(path, srcPath) {
// 		debug.Log("Code path should be in GOPATH, but is in GOROOT: %v", path)
// 		return filepath.ToSlash(path[len(srcPath)+1:])
// 	}

// 	panic(fmt.Sprintf("Unexpected! Code path1 is not in GOPATH: %v", path))
// }
