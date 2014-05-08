package got

import (
	"fmt"
	"github.com/elivoa/got/templates"
	"got/core"
	"html/template"
)

type TemplateStatus struct {
	core.Component
	Tpls []*template.Template
}

type TemplatesJson struct {
	*templates.TemplateUnit
}

func (c *TemplateStatus) Setup() {
	c.Tpls = templates.Templates.Templates()
}

func (c *TemplateStatus) TemplatesJson() []*TemplatesJson {
	var json = []*TemplatesJson{}
	for _, unit := range templates.Cache.Templates {
		json = append(json, &TemplatesJson{
			TemplateUnit: unit,
		})
	}
	return json
}

// TODO: call this on page[got/status], event call on components are not worked now.
func (c *TemplateStatus) OnTemplateDetail(templateKey string) {
	fmt.Printf("-------------------------------------------------------------------------------------")
	unit := templates.Cache.Get(templateKey)
	if unit != nil {

	}
}
