/*
   Time-stamp: <[templates.go] Elivoa @ Saturday, 2016-12-10 17:40:21>
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
	"os"
	"strings"
	"sync"
)

var logTemplate = logs.Get("Log Template")

// TODO 1：加全局锁来解决多线程问题。
// TODO 2：多实例最后加锁合并。
// Engine instance. Unique.
var (
	Engine                         = core.NewTemplateEngine() // Root Template
	TemplateInitialized bool       = false
	Lock                sync.Mutex // 全局锁
)

func FinalInitialize() {
	Lock.Lock()
	defer Lock.Unlock()
	TemplateInitialized = true
	Engine.Template().Funcs(buildFuncMap())
}

/*_______________________________________________________________________________
  Register components
*/

// RegisterComponent register component as tempalte function.
// ComponentKey is converted to function name by replacing all shash '/' into '_'.
// Original cased and lowercased key is used. in component invoke.
// Note: 这个方法只在启动时调用。
func RegisterComponent(componentKey string, componentFunc interface{}) {
	funcName := fmt.Sprintf("t_%v", strings.Replace(componentKey, "/", "_", -1))
	funcmap := template.FuncMap{
		funcName:                  componentFunc,
		strings.ToLower(funcName): componentFunc,
	}
	Engine.Template().Funcs(funcmap) // Set function to root TemplateEngine.

	if false { // disabled-code
		// Loop all templates engine instance, set function to them.
		for _, seg := range register.TemplateKeyMap.Keymap {
			fmt.Println("\t>> RegisterComponent ", seg.Name)
			seg.TemplateEngine.Template().Funcs(funcmap)
		}
	}
}

/*_______________________________________________________________________________
  Render Tempaltes
*/

/*_______________________________________________________________________________
  GOT Templates Caches
*/

/*
  Load template and it's contents into memory. Then parse it into template.
  TODO: zip the source to save memory
  TODO: Implement force reload
*/
/*, protonType reflect.Type, key string, filename string*/
func LoadTemplates(registry *register.ProtonSegment, reloadWhenFileChanges bool) (
	/* returns: */ cached bool, engine *core.TemplateEngine, err error) {

	_, templatePath := registry.TemplatePath()

	if logTemplate.Info() {
		logTemplate.Printf("[ParseTemplate] Identity:%v", registry.Identity())
		logTemplate.Printf("[ParseTemplate] FullPath:%v", templatePath)
		logTemplate.Printf("[ParseTemplate] registry:%v", registry.Name)
		logTemplate.Printf("[ParseTemplate] registry Alias:%v", registry.Alias)
	}

	// TODO: 这里的锁有问题，高并发时容易引起资源浪费。 /锁呢？
	if !reloadWhenFileChanges && registry.IsTemplateLoaded {
		// Be Lazy, err is Tempalte not loaded yet!
		if nil == registry.TemplateEngine {
			panic("不可能!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		}
		cached = true
		engine = registry.TemplateEngine
		return
	}

	// load and parse it.
	// fmt.Println("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&")
	// fmt.Println("Lock registry.L.Lock()")
	registry.L.Lock() // write lock
	// fmt.Println("Lock registry.L.Lock() success")
	defer func() {
		registry.L.Unlock()
	}()

	// if file doesn't exist.
	var fileInfo os.FileInfo
	if fileInfo, err = os.Stat(templatePath); os.IsNotExist(err) {
		// set nil to cache
		// Set loaded flag to true even if file not exist. FileNotExist is a normal case.
		fmt.Println("文件不存在， template not exist.!")
		registry.IsTemplateLoaded = true
		return false, nil, err
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
				engine = registry.TemplateEngine
				if engine == nil {
					fmt.Println("绝对不可能")
				}
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
		fmt.Println("\n\n---- [CONTENT TRANSFERED] --------------------------------------------------")
		fmt.Println(registry.ContentTransfered)
		fmt.Println("----------------------------------------------------------------------\n\n-")

		fmt.Println("\n\n---- [IMPORTS IN BLOCK] ----------------------------------------------------")
		fmt.Println(registry.ContentTransfered)
		fmt.Println("----------------------------------------------------------------------\n\n-")
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
	if seg, ok := register.TemplateKeyMap.Keymap[registry.Identity()]; ok {
		fmt.Println("Set engine from x to x; ", engine, seg.TemplateEngine)
		engine = seg.TemplateEngine // set return value
		// cached templates
		fmt.Println("\n\n\n\n\n++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
		fmt.Println("cached templates, ignore", registry.Identity())
	} else {
		// if not cached.
		engine = Engine.Clone()
		registry.TemplateEngine = engine

		if err = ParseTemplate(engine, registry.Identity(), registry.ContentTransfered); err != nil {
			if debug.QuickFixEnabled {
				panic(err)
			} else {
				panic(err)
			}
		}

		// blocks found in template.
		blocks := trans.RenderBlocks()
		if blocks != nil {
			registry.Blocks = map[string]*register.Block{}
			for blockId, html := range blocks {
				block := &register.Block{
					ID:                blockId,
					ContentTransfered: html,
				}
				registry.Blocks[blockId] = block
				blockKey := fmt.Sprintf("%s%s%s", registry.Identity(), config.SPLITER_BLOCK, blockId)
				// fmt.Println("debug info { RenderBlock inline key is} ", blockKey)
				if err = ParseTemplate(engine, blockKey, block.ContentTransfered); err != nil {
					fmt.Printf("~~~ Error when parse template %x", blockKey)
					panic(err)
				}
			}
		}

		registry.SetAssets(trans.Assets) // set assets into registry
		register.TemplateKeyMap.Keymap[registry.Identity()] = registry
	}
	return
}

func ParseTemplate(engine *core.TemplateEngine, key string, content string) error {
	tmpl := engine.InitTemplate(key)
	_, err := tmpl.Parse(content)
	if err != nil {
		return err
	}
	return nil
}
