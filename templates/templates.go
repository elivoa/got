/*
   Time-stamp: <[templates.go] Elivoa @ Sunday, 2014-05-11 00:18:01>
*/
package templates

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/elivoa/got/templates/transform"
	"got/debug"
	"html/template"
	"io"
	"log"
	"os"
	"reflect"
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

/*
 @Param: key is template's key. not including blocks in it.
 @return TemplateUnits.
   TemplateUnits - when tempalte available and parse successful.
   error         - errors occured.
*/
func parseTemplates(key string, filename string) (map[string]*TemplateUnit, error) {

	debug.Log("-   - [ParseTemplate] %v, %v", key, filename)

	// borrowed from html/tempate
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, nil // file not exist, don't panic.
	}

	// open input file
	fi, err := os.Open(filename)
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
	trans.Parse(r)

	templatesToParse := map[string]string{}
	templatesToParse[key] = trans.RenderToString() // main block
	blocks := trans.RenderBlocks()                 // blocks found in template.
	if blocks != nil {
		for blockId, html := range blocks {
			templatesToParse[fmt.Sprintf("%v:%v", key, blockId)] = html
		}
	}

	// fmt.Println(">>>> ------------------------------------------------------------------------------")
	// fmt.Println(">>>> ------------------", filename, "--------------------------------------")
	// fmt.Println(templatesToParse[key])
	// fmt.Println("<<<< ``````````````````````````````````````````````````````````````````````````````````")

	// Old version uses filename as key, I make my own key. not
	// filepath.Base(filename) First template becomes return value if
	// not already defined, and we use that one for subsequent New
	// calls to associate all the templates together. Also, if this
	// file has the same name as t, this file becomes the contents of
	// t, so t, err := New(name).Funcs(xxx).ParseFiles(name)
	// works. Otherwise we create a new template associated with t.
	returns := map[string]*TemplateUnit{}
	t := Templates
	for _key, html := range templatesToParse {
		var tmpl *template.Template
		if t == nil {
			t = template.New(_key)
		}
		if _key == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(_key)
		}

		_, err = tmpl.Parse(html)
		if err != nil {
			return nil, err
		}
		returns[_key] = &TemplateUnit{
			Key:               _key,
			FilePath:          filename,
			ContentOrigin:     "",
			ContentTransfered: html,
			// Template:          tmpl, // only one *template.Template
			IsCached: true,
		}

	}

	return returns, nil
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

// init template cache
var Cache TemplateCache = TemplateCache{
	Templates: map[string]*TemplateUnit{},
}

// TemplateCache cache templates
type TemplateCache struct {
	l sync.RWMutex

	// fullpath as key?
	// Changed to template key as key.
	Templates map[string]*TemplateUnit
}

type TemplateUnit struct {
	Key               string
	FilePath          string
	ContentOrigin     string `json:"-"`
	ContentTransfered string `json:"-"`
	IsBlock           bool   // Is this template a block? false means a main template.
	IsCached          bool   `json:"-"`
}

func (t *TemplateCache) Get(key string) *TemplateUnit {
	var unit *TemplateUnit
	t.l.RLock()
	unit, _ = t.Templates[key]
	t.l.RUnlock()
	return unit
}

// TODO Test & Improve Performance.
// TODO the first return value is not used.
// TODO the name shoud change
func (t *TemplateCache) GetnParse(key string, templatePath string, pageType reflect.Type) (*TemplateUnit, error) {
	/* TODO 这里模板上锁的机制有问题。
	   1. 先上锁判断是否存在，然后初始化，设置的时候上第二道锁；
	      缺点：并发多的时候，会有多个进程同时初始化。
	   2. 解决方案, 用rw嗦，读取写入的时候上多到嗦。
	*/
	t.l.RLock()
	// If has something means template is cached. maybe changed in the .
	_, ok := t.Templates[key] // chagne key as key
	// _, ok := t.Templates[templatePath] // old, path as key.
	t.l.RUnlock()
	if ok {
		return nil, nil
	}

	if !ok {
		units, err := parseTemplates(key, templatePath) // returns map[string]*template.Template, error
		// this tmpls including main block and blocks.
		if err != nil { // error occured
			return nil, err
		}
		if units == nil { // no error and no result.
			err = errors.New(fmt.Sprintf("Templates for '%v' not found!", key))
			return nil, err
		}

		t.l.Lock() // write lock
		// write template and it's blocks into cache.
		for _key, unit := range units {
			t.Templates[_key] = unit
		}
		t.l.Unlock()

		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Println(units[key].ContentTransfered)
		return units[key], nil
	}
	return nil, nil
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
