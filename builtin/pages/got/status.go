package got

import (
	"errors"
	"fmt"
	"github.com/elivoa/got/cache"
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/register"
	"github.com/elivoa/got/route/exit"
	"html/template"
	"strings"
)

// TODO render blocks as tree structure.
// TODO order templates display.
// TODO show pages.
// TODO show components.
type Status struct {
	core.Page

	// parameters
	Tab string `path-param:"1"` // tab

	Modules *register.ModuleCache
}

func (p *Status) SetupRender() *exit.Exit {
	// PrintSourceCaches()
	register.DeubgPrintTypeMaps()

	if false {
		panic(errors.New("CheDan ====== "))
	}

	p.Modules = register.Modules

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

func (p *Status) SourceCache() template.HTML {
	c := cache.SourceCache
	for idx, v := range c.Structs {
		fmt.Printf(">%d\t  %s:'%s.%s'\n", idx, core.KindLabels[v.ProtonKind], v.PackageName, v.StructName)
	}
	return template.HTML("")
}

func (p *Status) StructInfo() template.HTML {
	c := cache.StructCache
	html := c.String()
	fmt.Println(html)
	html = strings.Replace(html, " ", "&nbsp;", -1)
	html = strings.Replace(html, "\n", "<br>", -1)
	return template.HTML(html)
}

func PrintSourceCaches() {
	source := cache.SourceCache
	for k, v := range source.StructMap {
		fmt.Println("  in source map: ", k, v)
	}
}

func (p *Status) Style(tab string) string {
	if p.Tab == tab {
		return "active"
	}
	return ""
}

func (p *Status) OnTab(tab string) *exit.Exit {
	return exit.Redirect(fmt.Sprintf("/got/status/%s", tab))
}

func (p *Status) OnGotoHome() *exit.Exit {
	// return exit.Forward(p.IndexPage) //
	return exit.Forward("/")
}
