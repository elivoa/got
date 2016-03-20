/*
   Time-stamp: <[templates.go] Elivoa @ Monday, 2016-03-21 01:09:29>
*/
package templates

import (
	"bufio"
	"fmt"
	"github.com/elivoa/got/config"
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/debug"
	"github.com/elivoa/got/logs"
	"github.com/elivoa/got/register"
	"github.com/elivoa/got/templates/transform"
	"html/template"
	"io"
	"os"
	"strings"
	"sync"
)

var logTemplate = logs.Get("Log Template")

var TemplateInitialized bool = false

func FinalInitialize() {
	// TODO add lock.
	TemplateInitialized = true
	Engine.template.Funcs(buildFuncMap())
}

// Engine instance. Unique.
var Engine = NewTemplateEngine()

type TemplateEngine struct {
	template *template.Template
}

func NewTemplateEngine() *TemplateEngine {
	e := &TemplateEngine{
		// init template TODO remove this, change another init method.
		// TODO: use better way to init.
		template: template.New("-"),
	}
	return e
}

/*_______________________________________________________________________________
  Register components
*/

// RegisterComponent register component as tempalte function. ComponentKey is converted to function name by replacing all shash '/' into '_'. Original cased and lowercased key is used. in component invoke.
func (e *TemplateEngine) RegisterComponent(componentKey string, componentFunc interface{}) {
	funcName := fmt.Sprintf("t_%v", strings.Replace(componentKey, "/", "_", -1))
	e.template.Funcs(template.FuncMap{
		funcName:                  componentFunc,
		strings.ToLower(funcName): componentFunc,
	})
}

/*_______________________________________________________________________________
  Render Tempaltes
*/

// RenderTemplate render template into writer.
func (e *TemplateEngine) RenderTemplate(w io.Writer, key string, p interface{}) error {

	// // cache some panic
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		fmt.Println("\n\n====== Panic Occured When rendering template. =============")
	// 		panic(err)
	// 	}
	// }()

	// TODO: process key, with versions.
	err := Engine.template.ExecuteTemplate(w, key, p)
	if err != nil {
		return err
	}
	return nil
}

// 如果不存在没关系
func (e *TemplateEngine) RenderBlock(w io.Writer, templateId, blockId string, p interface{}) error {
	blockKey := fmt.Sprintf("%s%s%s", templateId, config.SPLITER_BLOCK, blockId)
	err := Engine.template.ExecuteTemplate(w, blockKey, p)
	if err != nil {
		return err
	}
	return nil
}

func (e *TemplateEngine) RenderBlockIfExist(w io.Writer, templateId, blockId string, p interface{}) error {
	blockKey := fmt.Sprintf("%s%s%s", templateId, config.SPLITER_BLOCK, blockId)
	if t := Engine.template.Lookup(blockKey); t != nil {
		err := Engine.template.ExecuteTemplate(w, blockKey, p)
		if err != nil {
			return err
		}
	}
	return nil
}

/*_______________________________________________________________________________
  GOT Templates Caches
*/

var l sync.RWMutex

/*
  Load template and it's contents into memory. Then parse it into template.
  TODO: zip the source to save memory
  TODO: Implement force reload
*/
/*, protonType reflect.Type, key string, filename string*/
func LoadTemplates(registry *register.ProtonSegment, reloadWhenFileChanges bool) (cached bool, err error) {

	_, templatePath := registry.TemplatePath()

	if logTemplate.Info() {
		logTemplate.Printf("[ParseTemplate] Identity:%v", registry.Identity())
		logTemplate.Printf("[ParseTemplate] FullPath:%v", templatePath)
		logTemplate.Printf("[ParseTemplate] registry:%v", registry.Name)
		logTemplate.Printf("[ParseTemplate] registry Alias:%v", registry.Alias)
	}

	// TODO: 这里的锁有问题，高并发时容易引起资源浪费。
	if !reloadWhenFileChanges && registry.IsTemplateLoaded {
		// Be Lazy, err is Tempalte not loaded yet!
		cached = true
		return // return cached version.
	}

	// load and parse it.
	registry.L.Lock() // write lock
	defer registry.L.Unlock()

	// if file doesn't exist.
	var fileInfo os.FileInfo
	if fileInfo, err = os.Stat(templatePath); os.IsNotExist(err) {
		// set nil to cache
		// Set loaded flag to true even if file not exist. FileNotExist is a normal case.
		registry.IsTemplateLoaded = true
		return
	} else if err != nil {
		panic(err) // panic on other file error.
	} else {

		if false {
			fmt.Println("\n==============================================")
			fmt.Println("Registry is  :", registry)
			fmt.Println("Cached Time  : ", registry.TemplateLastModifiedTime)
			fmt.Println("fileinfo Time: ", fileInfo.ModTime())
			fmt.Println("Are they eq? : ", fileInfo.ModTime() == registry.TemplateLastModifiedTime)
		}

		// Normal case: file found and no error.
		if reloadWhenFileChanges == true && registry.IsTemplateLoaded {
			// if not the first time meet this template, process versions.
			if registry.TemplateLastModifiedTime == fileInfo.ModTime() {
				// nothing changed, return cached one.
				cached = true
				return
			} else {
				registry.IncTemplateVersion()
				// >> go through to reload the file.
			}
		}
		// Mark file as loaded.
		registry.IsTemplateLoaded = true
		registry.TemplateLastModifiedTime = fileInfo.ModTime()
	}

	//
	// open input file
	//
	fi, err := os.Open(templatePath)
	if err != nil {
		panic(err)
	}

	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	// make a read buffer
	r := bufio.NewReader(fi)

	// transform
	trans := transform.NewTransformer()
	trans.Parse(r, registry.StructInfo.ProtonKind == core.PAGE) // then trans has components

	registry.ContentTransfered = trans.RenderToString()
	if false {
		// fmt.Println("\n\n---- [CONTENT TRANSFERED] --------------------------------------------------")
		// fmt.Println(registry.ContentTransfered)
		// fmt.Println("----------------------------------------------------------------------\n\n-")

		// fmt.Println("\n\n---- [IMPORTS IN BLOCK] ----------------------------------------------------")
		// fmt.Println(registry.ContentTransfered)
		// fmt.Println("----------------------------------------------------------------------\n\n-")
	}

	// append components
	if nil != trans.Components && len(trans.Components) > 0 {
		if nil == registry.EmbedComponents {
			registry.EmbedComponents = map[string]*register.ProtonSegment{}
		}
		for _, componentInfo := range trans.Components {
			registry.EmbedComponents[strings.ToLower(componentInfo.ID)] = componentInfo.Segment
		}
	}

	// parse tempalte

	// [debug:print template keymaps]
	fmt.Println("============== Key map is : ================")
	for k, v := range register.TemplateKeyMap.Keymap {
		fmt.Printf("\t%s => %s\n", k, v)
	}
	fmt.Println()

	if _, ok := register.TemplateKeyMap.Keymap[registry.Identity()]; ok {
		// cached templates
		fmt.Println("\n\n\n\n\n++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		fmt.Println("cached templates, ignore", registry.Identity())
	} else {
		// if not cached.
		var meet_an_bug = false
		if err = parseTemplate(registry.Identity(), registry.ContentTransfered); err != nil {

			if debug.QuickFixEnabled {
				meet_an_bug = true
				// #01 Quick Fix.
				if strings.Index(err.Error(), "html/template: cannot redefine") == 0 {
					fmt.Printf("[>QUICK-FIX] #01: %v\n", err)
					err = nil
					// ignore all then return. because all tempalte are already registered.
				} else {
					// TODO: Goto Detailed template parse Error page.
					panic(err)
					// panic(fmt.Sprintf("Error when parse template %x", identity))
				}
			} else {
				panic(err)
			}
		}

		if !meet_an_bug {
			blocks := trans.RenderBlocks() // blocks found in template.
			if blocks != nil {
				registry.Blocks = map[string]*register.Block{}
				for blockId, html := range blocks {
					block := &register.Block{
						ID:                blockId,
						ContentTransfered: html,
					}
					registry.Blocks[blockId] = block
					blockKey := fmt.Sprintf("%s%s%s", registry.Identity(), config.SPLITER_BLOCK, blockId)
					if err = parseTemplate(blockKey, block.ContentTransfered); err != nil {
						panic(fmt.Sprintf("Error when parse template %x", blockKey))
					}
				}
			}
		}

		registry.SetAssets(trans.Assets) // set assets into registry

		// debug print assets
		// if nil != registry.Assets {
		// 	registry.Assets.DebugPrintAll()
		// }

		// add to cache
		register.TemplateKeyMap.Keymap[registry.Identity()] = registry
	}
	return
}

func parseTemplate(key string, content string) error {
	// Old version uses filename as key, I make my own key. not
	// filepath.Base(filename) First template becomes return value if
	// not already defined, we use that one for subsequent New
	// calls to associate all the templates together. Also, if this
	// file has the same name as t, this file becomes the contents of
	// t, so t, err := New(name).Funcs(xxx).ParseFiles(name)
	// works. Otherwise we create a new template associated with t.

	// fmt.Printf("[parse tempalte] parseTempalte(%s,<<%s>>);\n", key, content) //content) // REMOVE

	var tmpl *template.Template
	if Engine.template == nil {
		Engine.template = template.New(key)
	}

	if key == Engine.template.Name() {
		tmpl = Engine.template
	} else {
		tmpl = Engine.template.New(key)
		// Engine.template = tmpl
	}

	if false { // -------------------------- debug print templates.
		fmt.Println("--$$$$$$$$$$$$--")
		for _, t := range Engine.template.Templates() {
			fmt.Println("\t", t.Name())
		}
		fmt.Println("<<< $$$")
	}
	_, err := tmpl.Parse(content)

	// fmt.Printf("[parse tempalte] End parseTempalte(%s, << ignored >>);\n", key) // REMOVE
	if err != nil {
		// fmt.Println("[ERROR] : \t", err) // REMOVE
		return err
	}
	return nil
}
