package got

import (
	"fmt"
	"github.com/elivoa/got/route/exit"
	"got/core"
	"got/register"
)

type TemplateStatus struct {
	core.Component
	Templates []*TemplatesJson
}

type TemplatesJson struct {
	Key      string
	FilePath string
}

func (c *TemplateStatus) Setup() {
	c.Templates = c.TemplatesJson()
}

func (c *TemplateStatus) TemplatesJson() []*TemplatesJson {
	var json = []*TemplatesJson{}
	for key, seg := range register.TemplateKeyMap.Keymap {
		_, path := seg.TemplatePath()
		json = append(json, &TemplatesJson{
			Key:      key,
			FilePath: path,
		})
	}
	return json
}

// TODO: call this on page[got/status], event call on components are not worked now.
func (c *TemplateStatus) OnTemplateDetail(templateKey string) *exit.Exit {
	fmt.Printf("-------------------------------------------------------------------------------------")
	// if index := strings.LastIndex(templateKey, ":"); index > 0 {
	// 	// has block
	// 	if unit, err := templates.Cache.GetBlockByKey(templateKey[0:index], templateKey[index:]); err != nil {
	// 		if unit != nil {
	// 			return exit.RenderText(unit.ContentTransfered)
	// 		}
	// 	}
	// } else {
	// 	// no block
	// 	if unit, err := templates.Cache.GetByKey(templateKey); err != nil {
	// 		if unit != nil {
	// 			return exit.RenderText(unit.ContentTransfered)
	// 		}
	// 	}
	// }
	return exit.RenderText("")
}
