/*
   Time-stamp: <[templates.go] Elivoa @ Thursday, 2014-07-10 17:12:31>
*/
package templates

import (
	"bufio"
	"fmt"
	"github.com/elivoa/got/config"
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

	registerBuiltinFuncs(e.template) // Register built-in templates.
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
	// TODO: process key, with versions.
	err := Engine.template.ExecuteTemplate(w, key, p)
	if err != nil {
		return err
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

	registry.ContentTransfered = trans.RenderToString()

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
	if err = parseTemplate(registry.Identity(), registry.ContentTransfered); err != nil {
		// TODO: Detailed template parse Error page.
		panic(err)
		// panic(fmt.Sprintf("Error when parse template %x", identity))
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

	t := Engine.template
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
