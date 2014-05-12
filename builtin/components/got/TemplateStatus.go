package got

import (
	"bytes"
	"fmt"
	"github.com/elivoa/got/route/exit"
	"github.com/elivoa/got/util"
	"got/core"
	"got/register"
	"strings"
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
	var buffer bytes.Buffer
	buffer.WriteString("Details for Template: ")
	key := util.DecodeContext(templateKey)
	buffer.WriteString(key)
	buffer.WriteString("\n\n")

	isblock := false
	blockname := ""
	if index := strings.LastIndex(key, ":"); index > 0 {
		blockname = templateKey[index:]
		key = key[0:index]
		isblock = true
	}
	// TODO: can't see templates that are not initialized.
	// currentSeg := FollowComponentByIds(lcc.page.registry, result.ComponentPaths)

	if seg, ok := register.TemplateKeyMap.Keymap[key]; ok {
		if !isblock {
			buffer.WriteString(seg.ContentTransfered)
		} else {
			if block, ok := seg.Blocks[blockname]; ok {
				buffer.WriteString(block.ContentTransfered)
			}
		}
		return exit.RenderText(buffer.String())
	}
	// panic(fmt.Sprintf("Template Not Found for %v", util.DecodeContext(templateKey)))
	buffer.WriteString(fmt.Sprintf("\n\nTemplate Not Found for %v", util.DecodeContext(templateKey)))
	return exit.RenderText(buffer.String())

}
