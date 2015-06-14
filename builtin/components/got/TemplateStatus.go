package got

import (
	"bytes"
	"fmt"
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/register"
	"github.com/elivoa/got/route/exit"
	"github.com/elivoa/got/util"
	"sort"
	"strings"
)

type TemplateStatus struct {
	core.Component
	Templates TemplatesJsonCollection // []*TemplatesJson

	//
}

type TemplatesJson struct {
	Key      string
	FilePath string
	Seg      *register.ProtonSegment `json:""`
}

// for sort
type TemplatesJsonCollection []*TemplatesJson

func (p TemplatesJsonCollection) Len() int           { return len(p) }
func (p TemplatesJsonCollection) Less(i, j int) bool { return p[i].Key < p[j].Key }
func (p TemplatesJsonCollection) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (c *TemplateStatus) Setup() {
	c.Templates = c.TemplatesJson()
}

func (c *TemplateStatus) TemplatesJson() []*TemplatesJson {
	var json = TemplatesJsonCollection{}
	for key, seg := range register.TemplateKeyMap.Keymap {
		_, path := seg.TemplatePath()
		json = append(json, &TemplatesJson{
			Key:      key,
			FilePath: path,
			Seg:      seg,
		})
	}
	// TODO sort it;
	sort.Sort(json)
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
		fmt.Println("::::::::::::::::::::::", blockname)
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
