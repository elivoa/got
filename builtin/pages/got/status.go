package got

import (
	"errors"
	"fmt"
	"github.com/elivoa/got/register"
	"github.com/elivoa/got/route/exit"
	"github.com/elivoa/got/templates"
	"got/cache"
	"got/core"
	"html/template"
	"strings"
	page "syd/pages"
)

// TODO render blocks as tree structure.
// TODO order templates display.
// TODO show pages.
// TODO show components.
type Status struct {
	core.Page

	Modules *register.ModuleCache
	// Pages      *register.ProtonSegment
	// Components *register.ProtonSegment
	Tpls []*template.Template

	// redirect to this page.
	// TODO: 如何Inject一个page？ page的包名太长不好记怎么办？
	IndexPage *page.Index `inject:"page"` //:"something that can't be emptys"`
}

func (p *Status) SetupRender() *exit.Exit {
	// PrintSourceCaches()
	register.DeubgPrintTypeMaps()

	if false {
		panic(errors.New("CheDan ====== "))
	}

	p.Tpls = templates.Templates.Templates()
	p.Modules = register.Modules
	// p.Pages = &register.Pages

	return nil
}

func (p *Status) Pages() template.HTML {
	html := register.Pages.StringTree("<br>")
	html = strings.Replace(html, " ", "&nbsp;", -1)
	return template.HTML(html)
}

func (p *Status) Components() template.HTML {
	html := register.Components.StringTree("<br>")
	html = strings.Replace(html, " ", "&nbsp;", -1)
	return template.HTML(html)
}

func PrintSourceCaches() {
	source := cache.SourceCache
	for k, v := range source.StructMap {
		fmt.Println("  in source map: ", k, v)
	}
}

func (p *Status) AfterRender() {
}

func (p *Status) OnGotoHome() *exit.Exit {
	return exit.Forward(p.IndexPage)
}
