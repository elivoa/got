package core

import (
	"fmt"
	"html/template"
	"io"
)

type TemplateEngine struct {
	template *template.Template // Main Tempalte
	// Lock sync.Mutex // the global template lock
}

func (e *TemplateEngine) Template() *template.Template {
	return e.template
}

func (e *TemplateEngine) InitTemplate(key string) *template.Template {
	if e.template == nil {
		e.template = template.New(key)
	}
	if key != e.template.Name() {
		e.template = e.template.New(key)
	}
	return e.template
}

func (e *TemplateEngine) String() string {
	templateName := " [nil] "
	if nil != e.template {
		templateName = e.template.Name()
	}
	return fmt.Sprintf("TemplateEngine(%s)", templateName)
}

func NewTemplateEngine() *TemplateEngine {
	e := &TemplateEngine{
		// init template TODO remove this, change another init method.
		// TODO: use better way to init.
		template: template.New(" - "),
		// renderDepth: 0,
	}
	return e
}

func (e *TemplateEngine) renderTemplate(w io.Writer, templateKey string, p interface{}) error {
	err := e.template.ExecuteTemplate(w, templateKey, p)
	if err != nil {
		return err
	}
	return nil
}

func (e *TemplateEngine) Clone() *TemplateEngine {
	var newt *template.Template
	if nil != e.template {
		if temp, err := e.template.Clone(); err != nil {
			panic(err)
		} else {
			newt = temp
		}
	}
	newe := &TemplateEngine{
		template: newt,
	}
	fmt.Println(">> TemplateEngine.Clone(), Name is", newt.Name())
	return newe
}

// final combine.
func Combine() {
	fmt.Println("******* The final Combine *******")
}

// RenderTemplate render template into writer.

func (e *TemplateEngine) RenderTemplate(w io.Writer, key string, p interface{}) error {
	return e.renderTemplate(w, key, p) // TODO: process key, with versions.
}

var SPLITER_BLOCK = ":"

// 如果不存在没关系
func (e *TemplateEngine) RenderBlock(w io.Writer, templateId, blockId string, p interface{}) error {
	blockKey := fmt.Sprintf("%s%s%s", templateId, SPLITER_BLOCK, blockId)
	return e.renderTemplate(w, blockKey, p)
}

func (e *TemplateEngine) RenderBlockIfExist(w io.Writer, templateId, blockId string, p interface{}) error {
	blockKey := fmt.Sprintf("%s%s%s", templateId, SPLITER_BLOCK, blockId)
	if t := e.template.Lookup(blockKey); t != nil {
		return e.renderTemplate(w, blockKey, p)
	}
	return nil
}
