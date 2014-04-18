package got

import (
	"fmt"
	"got/core"
	"got/register"
	"got/templates"
	"html/template"
)

// TODO render blocks as tree structure.
// TODO order templates display.
// TODO show pages.
// TODO show components.
type Status struct {
	core.Page

	Modules    *register.ModuleCache
	Pages      *register.ProtonSegment
	Components *register.ProtonSegment
	Tpls       []*template.Template
}

func (p *Status) SetupRender() {
	fmt.Println(">>>>>>>>>>> set up redner s >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	p.Tpls = templates.Templates.Templates()
	p.Modules = register.Modules
	p.Pages = &register.Pages
}

func (p *Status) AfterRender() {
}

func (p *Status) OnClickTemplate(name string) {
	fmt.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")
	fmt.Println("click template: ", name)
	t := templates.Templates.Lookup(name)
	fmt.Println(t)
	fmt.Println(t.Delims)
}