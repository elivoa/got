/*
   Time-stamp: <[lifecircle-component.go] Elivoa @ Saturday, 2016-12-10 17:43:22>
*/
package lifecircle

import (
	"bytes"
	"fmt"
	"github.com/elivoa/got/config"
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/logs"
	"github.com/elivoa/got/register"
	"github.com/elivoa/got/route/exit"
	"github.com/elivoa/got/templates"
	"github.com/elivoa/got/utils"
	"html/template"
	"path"
	"reflect"
	"strings"
	"time"
)

var cflog = logs.Get("ComponentFLow")

// ComponentLifeCircle returns template-func to handle component render.
// Param: name e.g.: layout/gotheader
// Param: name - lowercased component key. e.g.: got/Component
func ComponentLifeCircle(name string) func(...interface{}) interface{} {

	// returnd string or template.HTML are inserted into final template.
	// params: 1. container, 2. uniqueID, 3. other parameter pairs...
	return func(params ...interface{}) interface{} {

		if cflog.Info() {
			cflog.Printf("[Component] Render Component %v (ID:%s) ....", name, params[1].(string))
		}

		// Processing component in one page
		// 1. find component by component's type
		// 2. find component's container
		// 3. create component's life
		// 4. process returns.

		// 1. find base component type
		result, err := register.Components.Lookup(name)
		if err != nil || result.Segment == nil {
			panic(fmt.Sprintf("Component %v not found!", name))
		}
		if len(params) < 1 {
			panic(fmt.Sprintf("First parameter of component must be '$' (container)"))
		}

		// 2. find container page or component
		container := params[0].(core.Protoner)
		tid := params[1].(string)
		containerLife := container.FlowLife().(*Life)
		{
			// fmt.Println("~~~~~~==~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
			// fmt.Println("container:", utils.GetRootType(container), "  >>> ", container.Kind())
			// fmt.Println("seed comp:", reflect.TypeOf(result.Segment.Proton))
			// // fmt.Println("component:", lcc.current.rootType, " >>> ", result.Segment.Proton.Kind())
			// // fmt.Println("container:", utils.GetRootType(container), "  >>> ", container.ClientId())
			// // fmt.Println("seed comp:", reflect.TypeOf(result.Segment.Proton))
			// // fmt.Println("component:", lcc.current.rootType, " >>> ", result.Segment.Proton.ClientId())
			// fmt.Println("\n")
		}
		// unused: get lcc from component; use method to get from controler.
		// lcc := context.Get(container.Request(), config.LCC_OBJECT_KEY).(*LifeCircleControl)
		lcc := containerLife.control // container's lcc has the same R and W.
		life := lcc.componentFlow(containerLife, (core.Componenter)(result.Segment.Proton), tid, params[2:])
		life.SetRegistry(result.Segment)

		// templates renders in common flow()
		returns := life.flow()
		if returns.IsBreakExit() {
			lcc.returns = returns
			lcc.rendering = false
			// here don't process returns, handle return in page-flow's end.
			// here only set returns into control and stop the rendering.
		}

		// If returns is not template-renderer (i.e.: redirect or text output),
		// flow breaks and will not reach here.
		// Here returns default template render.
		rr := template.HTML(life.out.String())

		return rr
	}
}

// --------------------------------------------------------------------------------
//
// Create a new Component Flow.
// param:
//   container - real container object.
//   component - current component base object.
//   params - parameters in the component grammar.
//
// Note: I maintain StructCachep here in the flow create func. This occured only when
//       page or component arep rendered. Directly post to a page can not invoke structcache init.
//
// TODO: Performance Improve to Component in Loops.
//
func (lcc *LifeCircleControl) componentFlow(
	containerLife *Life, componentSeed core.Componenter, tid string, params []interface{}) *Life {

	if cflog.Debug() {
		cflog.Printf("----- [Component flow] ------------------------------------------------")
		cflog.Printf("- C - [Component Container] Type: %v, ComponentType:%v, tid:%s\n",
			containerLife.rootType, reflect.TypeOf(componentSeed), tid)
	}

	// debug.InspectComponentParameters(tid, params)

	// Store type in StructCache, Store instance in ProtonObject.
	// Warrning: What if use component in page/component but it's not initialized?
	// Tid= xxx in template must the same with fieldname in .go file.
	//

	// 1. cache in StructInfoCache. (application scope)
	containerSI := scache.GetCreate(containerLife.rootType, containerLife.kind)
	if containerSI == nil {
		panic(fmt.Sprintf("StructInfo for %v can't be null!", containerLife.rootType))
	}

	// TODO: Is below useful? Can i remove these codes?

	t := utils.GetRootType(componentSeed)
	containerSI.CacheEmbedProton(t, tid, componentSeed.Kind()) // TODO: is this useful?

	// 2. store in proton's embed field. (request scope)
	life, found := containerLife.embedmap[tid]
	if !found {
		// create component life!
		life := containerLife.appendComponent(componentSeed, tid)
		containerLife.proton.SetEmbed(tid, life.proton)
	} else {
		// already exist, must be in a loop.
		lcc.current = life
		// already found. maybe this component is in a loop.
		lcc.current.out.Reset()    // components in loop is one instance. reset after each use.
		life.proton.IncLoopIndex() // increase loop index
	}

	// 3. inject parameters into component.
	lcc.injectBasicTo(lcc.current.proton)
	lcc.injectComponentParameters(params) // inject component parameters to current life
	return lcc.current
}

// return (name, is setManually); t must not be ptr.
func determinComponentTid(params []interface{}, t reflect.Type) (tid string, specifiedId bool) {
	for idx, p := range params {
		if idx%2 == 0 {
			key := strings.ToLower(p.(string))
			if key == "tid" || key == "t:id" {
				tid = params[idx+1].(string)
				// use specified id, can't be duplicated (judge in transform.go)
				specifiedId = true
				return
			}
		}
	}
	// if fall here, measn no specified id.
	// TODO: get id
	tid = path.Ext(t.String())[1:]
	return
}

// --------------------------------------------------------------------------------

// flow controls the common lifecircles, including pages and components.
func (l *Life) flow() (returns *exit.Exit) {
	// There are 2 way to reach here.
	// 1. Page lifecircle, from PageFlow()
	// 2. Component's template-func, from func call. Get lcc from Request.

	// Here follows the flow of tapestry:
	//   http://tapestry.apache.org/component-rendering.html
	//
	// TODO: call lifecircle events with parameter
	// TODO: flat them
	for {
		returns = SmartReturn(l.call("Setup", "SetupRender"))
		if returns.IsBreakExit() {
			return
		}
		if !returns.IsReturnsFalse() {

			for {
				returns = SmartReturn(l.call("BeginRender"))
				if returns.IsBreakExit() {
					return
				}
				if !returns.IsReturnsFalse() {

					for {
						returns = SmartReturn(l.call("BeforeRenderTemplate"))
						if returns.IsBreakExit() {
							return
						}

						// render tempalte
						if !returns.IsReturnsFalse() {

							// Here we ignored BeforeRenderBody and AfterRenderBody.
							// Maybe add it later.
							// May be useful for Loop component?

							// fmt.Println(">>>>>>>>>>>>>>>LOCK LOCK LOCK LOCK LOCK LOCK 3 lock")
							// templates.Lock.Lock() // global lock on templates
							// fmt.Println(">>>>>>>>>>>>>>>LOCK LOCK LOCK LOCK LOCK LOCK 4 lock success")

							l.renderTemplate()

							// fmt.Println(">>>>>>>>>>>>>>>LOCK LOCK LOCK LOCK LOCK LOCK 3 unlock")
							// templates.Lock.Unlock()
							// fmt.Println(">>>>>>>>>>>>>>>LOCK LOCK LOCK LOCK LOCK LOCK s3 unlock success")

							// if any // component breaks it's render, stop all rendering.
							// if l.control.rendering == false {
							// 	returns = nil
							// 	return
							// }
						}

						returns = SmartReturn(l.call("AfterRenderTemplate"))
						if returns.IsBreakExit() {
							return
						}
						if !returns.IsReturnsFalse() {
							break
						}
					}
				}
				returns = SmartReturn(l.call("AfterRender"))
				if returns.IsBreakExit() {
					return
				}
				if !returns.IsReturnsFalse() {
					break
				}
			}
		}

		returns = SmartReturn(l.call("Cleanup", "CleanupRender"))
		if returns.IsBreakExit() {
			return
		}
		if !returns.IsReturnsFalse() {
			break // exit
		}
	}

	// finally I go through all render phrase.
	returns = exit.Template(nil)
	return

}

// renderTemplate find and render Template using go way.
func (l *Life) renderTemplate() {
	// reach here means I can find the template and render it.
	// debug.Log("-755- [TemplateSelect] %v -> %v", identity, templatePath)
	var (
		engine *core.TemplateEngine
		err    error
	)
	if _, engine, err = templates.LoadTemplates(l.registry, config.ReloadTemplate); err != nil {
		panic(err)
	}

	if err := engine.RenderTemplate(&l.out, l.registry.Identity(), l.proton); err != nil {
		panic(err)
	}

	// PageHeadBootstrap Replace
	// TODO BIG Performance issue.
	if l.kind == core.PAGE {
		var start = time.Now().UnixNano()
		var dur int64
		if headbs, ok := l.proton.Embed("PageHeadBootstrap"); ok {

			var blockhtml bytes.Buffer
			life := headbs.FlowLife().(*Life)

			// BUG: What if this block not exists??
			if err := engine.RenderBlockIfExist(&blockhtml, life.registry.Identity(),
				"page_head_bootstrap_defer_block", headbs); err != nil {
				panic(err)
			}

			var PLACEHOLDER string = "(____PageHeadBootstrap_replace_to_html____)"
			newbuf := replaceInBuffer(&l.out, PLACEHOLDER, blockhtml.Bytes())
			l.out = *newbuf

			// var html string = l.out.String()
			// if index := strings.Index(html, PLACEHOLDER); index > 0 {
			// 	var newout bytes.Buffer
			// 	newout.WriteString(html[:index])

			// 	// debug print time
			// 	dur = (time.Now().UnixNano() - start)
			// 	newout.WriteString(fmt.Sprintf("<br/><br/><br/>Head duration is: %d ms", dur/1000))
			// 	// append final
			// 	newout.WriteString(html[index+len(PLACEHOLDER):])
			// 	l.out = newout
			// }
		}

		dur = (time.Now().UnixNano() - start)
		fmt.Println("[Performance] Time for ReplacePageHeadBootStrap is: ",
			dur/1000, "ms.")
	}
}

// Extract to tools.
func replaceInBuffer(buf *bytes.Buffer, s string, to []byte) *bytes.Buffer {
	var newbuf bytes.Buffer
	var err error
	var c byte
	var position = 0
	var cursor = 0
	var found_first int
	var matchedbytes = make([]byte, len(s))
	for err == nil {
		if c, err = buf.ReadByte(); err != nil {
			break // eof
		}
		// fmt.Printf("outer: char:%s position:%d\n", string(c), position)

		// fmt.Println("c == s[cursor]:  ", string(c), "==", string(s[cursor]), " cursor: ", cursor)
		if c == s[cursor] || c == s[0] {
			// fmt.Printf("inner: char:%s position:%d cursor:%d\n", string(c), position, cursor)
			if cursor == 0 {
				found_first = position
			} else if c == s[0] {
				// special   中途直接重启match e.g.: ((
				// fmt.Println(">> c == s[0]:  ", string(c), "==", string(s[cursor]), " cursor: ", cursor)
				newbuf.Write(matchedbytes[:cursor])
				cursor = 0
				found_first = position
			}
			// fmt.Println("cursor == len(s)-1:  ", cursor, "=?", len(s))
			if cursor == len(s)-1 { // found match
				// fmt.Println("Found cursor is: ", cursor, " char is ", string(s[cursor]))
				newbuf.Write(to)
				break // found
			}
			matchedbytes[cursor] = c
			// fmt.Println("matchedbytes:", string(matchedbytes))
			cursor += 1
		} else if found_first > 0 {
			// find part, start failed. reset.
			newbuf.Write(matchedbytes[:cursor])
			found_first = 0
			cursor = 0
		}

		if found_first == 0 { // find part, reset
			newbuf.WriteByte(c)
		}
		position += 1
	}
	newbuf.Write(buf.Bytes())
	return &newbuf
}
