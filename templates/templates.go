/*
   Time-stamp: <[templates.go] Elivoa @ Monday, 2014-05-12 22:34:51>
*/
package templates

import (
	"bufio"
	"fmt"
	"github.com/elivoa/got/config"
	"github.com/elivoa/got/templates/transform"
	"got/debug"
	"got/register"
	"html/template"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

// Templates stores all templates.
var Templates *template.Template

func init() {
	// init template TODO remove this, change another init method.
	// TODO: use better way to init.
	Templates = template.New("-")

	// Register built-in templates.
	registerBuiltinFuncs()
}

/*_______________________________________________________________________________
  Register components
*/

// register components as template function call.
func RegisterComponentAsFunc(name string, f interface{}) {
	funcName := fmt.Sprintf("t_%v", strings.Replace(name, "/", "_", -1))
	lowerFuncName := strings.ToLower(funcName)
	Templates.Funcs(template.FuncMap{funcName: f, lowerFuncName: f})
}

/*_______________________________________________________________________________
  Render Tempaltes
*/

// RenderTemplate render template into writer.
func RenderTemplate(w io.Writer, key string, p interface{}) error {
	err := Templates.ExecuteTemplate(w, key, p)
	if err != nil {
		return err
	}
	return nil
}

/*_______________________________________________________________________________
  GOT Templates Caches
*/

// TODO: Integrate to register.

// // init template cache
// var Cache TemplateCache = TemplateCache{
// 	// Templates: map[reflect.Type]*TemplateUnit{},
// 	Keymap: map[string]*TemplateUnit{},
// }

// TemplateCache cache templates
// type TemplateCache struct {
// 	l sync.RWMutex

// 	// Templates map[reflect.Type]*TemplateUnit // type as key
// 	Keymap map[string]*register.TemplateUnit // template key as key
// }

// Get cached TemplateUnit by proton type.

var l sync.RWMutex

// Parse and cache page or component's template, return the cached one.
// func GetnParse(key string, templatePath string, protonType reflect.Type) (*TemplateUnit, error) {
// 	/* TODO 这里模板上锁的机制有问题。
// 	   1. 先上锁判断是否存在，然后初始化，设置的时候上第二道锁；
// 	      缺点：并发多的时候，会有多个进程同时初始化。
// 	   2. 解决方案, 用rw嗦，读取写入的时候上多到嗦。
// 	*/
// 	fmt.Println(register.Components)
// 	// if !ok {
// 	forceLoad := false
// 	tu, cached, err := t.LoadTemplates(protonType, key, templatePath, forceLoad)
// 	if err != nil { // error occured
// 		return nil, err
// 	}
// 	if tu == nil { // no error and no result.
// 		err = errors.New(fmt.Sprintf("Templates for '%v' not found!", key))
// 		return nil, err
// 	}
// 	if !cached { // return the cached one.
// 		// parse templates
// 		t.l.Lock() // write lock
// 		ParseTemplate(tu)
// 		t.l.Unlock()
// 	}
// 	return tu, nil
// }

/*
  Load template and it's contents into memory. Then parse it into template.
  TODO: zip the source
  TODO: implement force reload
*/
/*, protonType reflect.Type, key string, filename string*/
func LoadTemplates(registry *register.ProtonSegment, forceReload bool) (cached bool, err error) {

	identity, templatePath := registry.TemplatePath()
	debug.Log("-   - [ParseTemplate] %v", templatePath)

	// TODO: 这里的锁有问题，高并发时容易引起资源浪费。
	if !forceReload { // read cache.
		if registry.IsTemplateLoaded {
			// Be Lazy, err is Tempalte not loaded yet!
			cached = true
			return // return cached version.
		}
	}

	// load and parse it.
	registry.L.Lock() // write lock
	defer registry.L.Unlock()

	// if file doesn't exist.
	if _, err = os.Stat(templatePath); os.IsNotExist(err) {
		// set nil to cache
		registry.IsTemplateLoaded = true
		return
	} else if err != nil {
		return // other file error.
	}

	// open input file
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
	trans.Parse(r) // then trans has components

	registry.IsTemplateLoaded = true
	registry.ContentTransfered = trans.RenderToString()

	// append components
	// TODO aspsign component id.
	if nil != trans.Components && len(trans.Components) > 0 {
		if nil == registry.EmbedComponents {
			registry.EmbedComponents = map[string]*register.ProtonSegment{}
		}
		// to be continued.
		for componentId, componentInfos := range trans.Components {
			var specifiedIndex int
			var hasSpecified bool
			var serial int
			// find specified, [only one for one id]
			for idx, compInfo := range componentInfos {
				if compInfo.IDSpecified {
					specifiedIndex = idx
					hasSpecified = true
					serial = 1
					break
				}
			}
			for idx, compInfo := range componentInfos {
				var realId string
				if hasSpecified && idx == specifiedIndex {
					realId = strings.ToLower(componentId)
				} else {
					if serial == 0 {
						realId = strings.ToLower(componentId)
					} else {
						realId = fmt.Sprintf("%s_%d", componentId, serial)
					}
				}
				registry.EmbedComponents[realId] = compInfo.Segment
				serial += 1
			}
		}
	}

	// var realId string
	// if count == 0 {
	// 	realId = componentId
	// } else {
	// 	realId = fmt.Sprintf("%s_%d", componentId, count)
	// }
	// t.Components[realId] = component

	// parse tempalte
	if err = parseTemplate(identity, registry.ContentTransfered); err != nil {
		// TODO: Detailed template parse Error page.
		panic(fmt.Sprintf("Error when parse template %x", identity))
	}

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
	// add to cache
	register.TemplateKeyMap.Keymap[registry.Identity()] = registry
	return
}

func parseTemplate(key string, content string) error {
	// Old version uses filename as key, I make my own key. not
	// filepath.Base(filename) First template becomes return value if
	// not already defined, and we use that one for subsequent New
	// calls to associate all the templates together. Also, if this
	// file has the same name as t, this file becomes the contents of
	// t, so t, err := New(name).Funcs(xxx).ParseFiles(name)
	// works. Otherwise we create a new template associated with t.

	t := Templates
	var tmpl *template.Template
	if t == nil {
		t = template.New(key)
	}
	if key == t.Name() {
		tmpl = t
	} else {
		tmpl = t.New(key)
	}

	_, err := tmpl.Parse(content)
	if err != nil {
		return err
	}
	return nil
}

// --------------------------------------------------------------------------------
// log
//
var debugLog = true

func debuglog(format string, params ...interface{}) {
	if debugLog {
		log.Printf(format, params...)
	}
}
