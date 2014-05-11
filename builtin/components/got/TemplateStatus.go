package got

import (
	"fmt"
	"github.com/elivoa/got/route/exit"
	"github.com/elivoa/got/templates"
	"got/core"
	"strings"
)

type TemplateStatus struct {
	core.Component
	Templates []*TemplatesJson
}

type TemplatesJson struct {
	*templates.TemplateUnit
}

func (c *TemplateStatus) Setup() {
	c.Templates = c.TemplatesJson()
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
func (c *TemplateStatus) OnTemplateDetail(templateKey string) *exit.Exit {
	fmt.Printf("-------------------------------------------------------------------------------------")
	if index := strings.LastIndex(templateKey, ":"); index > 0 {
		// has block
		if unit, err := templates.Cache.GetBlockByKey(templateKey[0:index], templateKey[index:]); err != nil {
			if unit != nil {
				return exit.RenderText(unit.ContentTransfered)
			}
		}
	} else {
		// no block
		if unit, err := templates.Cache.GetByKey(templateKey); err != nil {
			if unit != nil {
				return exit.RenderText(unit.ContentTransfered)
			}
		}
	}
	return exit.RenderText("")
}
