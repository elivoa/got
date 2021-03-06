/*
   Time-stamp: <[lifecircle.go] Elivoa @ Saturday, 2016-12-10 17:32:03>
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
	"github.com/elivoa/got/utils"
	"github.com/gorilla/context"
	"log"
	"net/http"
	"path"
	"reflect"
)

/* ________________________________________________________________________________

Injection Order:
  Page Render:       new -> path -> url ->
  Event Call:        new -> page's path -> page's url -> event func's parameter -> call
  Component Render:  new -> page's path -> page's url -> component parameter
  Form Submit:       new -> page's path -> page's url -> form values

TODO:
 . Callback functions
 . Cache Method to inject, borrowed form gorilla/schema
*/

// LifecircleControl controls flow of a Request.
type LifeCircleControl struct {
	// basic services
	w http.ResponseWriter
	r *http.Request

	// request related
	// pageUrl    string   // matched page url, Activate parameter part // no use.
	eventName  string   // an event call on page, not page render
	parameters []string // path parameters

	// lifes
	page    *Life // The Root Page Life
	current *Life // The Current Life

	// referer, used in event calls.
	referer *LifeCircleControl

	// Returns: store returned things.  || type:[template|redirect]
	rendering bool       // set to false to stop render.
	returns   *exit.Exit //
	Err       error      // error if something error. TODO change to MultiError.

	isForward bool // if true, ignore form submit.
}

// Life is a Page, Component, or others in the page render lifecircle.
type Life struct {
	// targets: are these too many?
	proton   core.Protoner // Pager or Componenter value
	rootType reflect.Type  // proton's root type
	v        reflect.Value // proton's reflect.Value
	kind     core.Kind     // enum: page|component
	name     string        // page name or component name
	tid      string        // component tid set in tempalte.

	registry *register.ProtonSegment // Is this really useful?

	// tree structure. TODO need children?
	control   *LifeCircleControl
	container *Life
	embedmap  map[string]*Life // what's this use?
	//embed     []*Life // not used

	// store rendered html.
	out bytes.Buffer

	// no use?
	// Path string // ???
}

// newControl create a new LifeCircleConstrol.
func newControl(w http.ResponseWriter, r *http.Request) *LifeCircleControl {
	lcc := &LifeCircleControl{w: w, r: r}
	return lcc
}

// createPage set the root page to lcc(LifeCircleControl).
func (lcc *LifeCircleControl) createPageLife(seed core.Protoner) *Life {
	life := newLife(seed)
	life.control = lcc
	lcc.page = life
	lcc.current = life
	debuglog("-710- [flow] New LifeCircleControl: %v[%v].", life.kind, life.name)
	return life
}

// create page by existing page value.
func (lcc *LifeCircleControl) createPageLifeFromExistingPage(page core.Pager) *Life {
	life := &Life{}
	life.v = reflect.ValueOf(page) // utils.GetRootValue(page) // newProtonInstance(seed)
	life.proton = life.v.Interface().(core.Protoner)
	life.name = fmt.Sprint(reflect.TypeOf(life.proton).Elem()) // remove dependence of fmt
	life.rootType = utils.GetRootType(page)
	life.kind = life.proton.Kind()
	life.proton.SetFlowLife(life)

	life.control = lcc
	lcc.page = life
	lcc.current = life
	debuglog("-710- [flow] New LifeCircleControl by page: %v[%v].", life.kind, life.name)
	return life
}

// appendComponent appends an embed component to lcc and chain it.
func (lcc *LifeCircleControl) appendComponent(seed core.Protoner, tid string) *Life {
	return lcc.current.appendComponent(seed, tid)
}

func (l *Life) appendComponent(seed core.Protoner, tid string) *Life {
	if seed.Kind() == core.PAGE {
		panic("Can't embed a Page!")
	}
	life := newLife(seed)
	life.tid = tid
	life.control = l.control
	life.container = l
	// append
	if l.embedmap == nil {
		l.embedmap = make(map[string]*Life)
	}
	l.embedmap[tid] = life
	l.control.current = life
	return life
}

// create new proton value from the seed.
func newLife(seed core.Protoner) *Life {
	life := &Life{}
	life.v = newProtonInstance(seed)
	life.proton = life.v.Interface().(core.Protoner)
	life.name = fmt.Sprint(reflect.TypeOf(life.proton).Elem()) // remove dependence of fmt
	life.rootType = utils.GetRootType(seed)
	life.kind = life.proton.Kind()
	life.proton.SetFlowLife(life)
	return life
}

// ---- utils ---------------------------------------------------------------------

func (lcc *LifeCircleControl) SetToRequest(key interface{}, value interface{}) {
	context.Set(lcc.r, key, value)
}

func (lcc *LifeCircleControl) GetFromRequest(key interface{}) interface{} {
	return context.Get(lcc.r, key)
}

// ---- Accessors -----------------------------------------------------------------
// func (lcc *LifeCircleControl) SetPageUrl(pageUrl string) *LifeCircleControl {
// 	lcc.pageUrl = pageUrl
// 	return lcc
// }

func (lcc *LifeCircleControl) SetEventName(event string) *LifeCircleControl {
	lcc.eventName = event
	return lcc
}

func (lcc *LifeCircleControl) SetParameters(parameters []string) *LifeCircleControl {
	lcc.parameters = parameters
	return lcc
}

// ---- Life ----------------------------------------------------------------------

func (l *Life) SetRegistry(registry *register.ProtonSegment) {
	l.registry = registry
}

func (l *Life) Registry() *register.ProtonSegment {
	return l.registry
}

// Call Events, and other events.
func (l *Life) call(names ...string) []reflect.Value {
	// fmt.Println("  ----  >> try call: ", names)
	// execute the first available method
	for _, name := range names {
		method := l.v.MethodByName(name)
		if method.IsValid() {
			debuglog("-730- [flow] Call Event: %v::%v().", l.name, name)
			returns := method.Call(emptyParameters)
			return returns
		}
	}
	return nil
}

// Generate base page url, forexample: /got/status. Fallback to the first container.
func (l *Life) GeneratePageUrl() string {
	var rootPageLife = l
	for rootPageLife.container != nil {
		rootPageLife = rootPageLife.container
	}
	seg := register.PageTypeMap[rootPageLife.rootType]
	return path.Join(seg.StructInfo.ProtonPath(), seg.StructInfo.StructName)
}

func (l *Life) Is(kind core.Kind) bool {
	return l.kind == kind
}

func (l *Life) GetRootType() reflect.Type {
	return l.rootType
}

func (l *Life) GetProton() core.Protoner {
	return l.proton
}

func (l *Life) GetContainer() *Life {
	return l.container
}

// ----

// ---- Print Structure --------------------------------------------------------------

func (lcc *LifeCircleControl) PrintCallStructure() string {
	var out bytes.Buffer
	printStructure(&out, lcc.page, 0)
	return out.String()
}

func printStructure(out *bytes.Buffer, life *Life, level int) {
	// print indent
	for i := 0; i < level; i++ {
		out.WriteString("  ")
	}

	out.WriteString("+ ")
	out.WriteString(life.String())
	out.WriteString("\n")
	// TODO ordered map
	if life.embedmap != nil {
		for _, v := range life.embedmap {
			printStructure(out, v, level+1)
		}
	}
}

func (l *Life) String() string {
	return fmt.Sprint(l.proton.Kind(), " ", l.name)
}

// ********************************************************************************
// ********************************************************************************
// ********************************************************************************

// ________________________________________________________________________________

// POST Flow,
//   Events:
//     OnSubmit    - Form submitted, called before inject form values.
//     OnValidate  - Validate form. called after form value injected.
//                   If returns false, render the current page, with errors
//     OnSuccess   - Called if OnValidate returns true.
//
// TODO post to components.
func (lcc *LifeCircleControl) PostFlow() (returns *exit.Exit) {
	// add ParseForm to fix bugs in go1.1.1
	err := lcc.r.ParseForm()
	if err != nil {
		lcc.Err = err
		return nil
	}

	// TODO use another method to retrive FormName. t:id

	formName := lcc.r.PostFormValue("t:id")
	if formName != "" {
		// formName = strings.Join([]string{"From", formName}, "")
		formName = fmt.Sprintf("From%v", formName)
	}
	// fmt.Println("********************************************************************************")
	// fmt.Println(formName)

	// call OnSubmit() method
	// onSubmitEventName := fmt.Sprintf("%v%v", "OnSubmit", formName) // old, change to new
	onSubmitEventName := fmt.Sprintf("%v%v", "OnPrepareForSubmit", formName)
	returns = SmartReturn(lcc.page.call(onSubmitEventName))
	if returns.IsBreakExit() {
		return
	}
	if returns.IsReturnsFalse() {
		return lcc.refreshThisPage()
	}

	// inject form values
	lcc.InjectFormValues()

	// call OnValidate() method
	onValidateEventName := fmt.Sprintf("%v%v", "OnValidate", formName)
	returns = SmartReturn(lcc.page.call(onValidateEventName))
	if returns.IsBreakExit() {
		return
	}
	if returns.IsReturnsFalse() {
		return lcc.refreshThisPage()
	}

	// call success method
	// call OnSuccess() method
	onSuccessEventName := fmt.Sprintf("%v%v", "OnSuccess", formName)
	returns = SmartReturn(lcc.page.call(onSuccessEventName))
	if returns.IsBreakExit() {
		return
	}
	// fmt.Println("***************************************************************************")
	// fmt.Println(onSuccessEventName)

	// something else, validation...
	// post flows stopd here.
	return lcc.refreshThisPage()
}

// ________________________________________________________________________________
// Call Events, and other events.
//
// func (lcc *LifeCircleControl) CallEvent(name string) bool {
// 	method := lcc.V.MethodByName(name)
// 	if method.IsValid() {
// 		debuglog("-730- [flow] Call Event: %v::%v().", lcc.Name, name)
// 		returns := method.Call(emptyParameters)
// 		return lcc.Return(returns...)
// 	} else {
// 		// debuglog("    - Event Not Found: %v", name)
// 	}
// 	return false
// }

// Call Events, with parameters. only used by Activate for now.
// TODO performance
func (lcc *LifeCircleControl) CallEventWithURLParameters(name string) bool {
	return lcc._callEventWithURLParameters(name, nil, lcc.page.v)
}

func (lcc *LifeCircleControl) _callEventWithURLParameters(name string, contexts []string, base reflect.Value) bool {
	fmt.Println("\n______________ CALL EVENT ______________________________")

	for idx, context := range contexts {
		// TODO inject values.
		fmt.Printf("-1- param #%d is: %v\n", idx, context)
	}

	//reflect.TypeOf(method).NumIn
	// method := lcc.V.MethodByName(name)
	method := base.MethodByName(name)
	if method.IsValid() {
		t := method.Type()
		fmt.Printf("-2- method is: %v\n", method)
		fmt.Printf("-2- typeof method is: %v\n", t)
		fmt.Printf("-2- param number is: %v\n", t.NumIn())

		numOfParameters := t.NumIn()
		parameters := make([]reflect.Value, numOfParameters)
		for i := 0; i < numOfParameters; i++ {
			pt := t.In(i)
			fmt.Printf("-3- param %v is: %v\n", i, pt)
			if len(contexts) > i {
				v, err := utils.Coercion(contexts[i], pt)
				if err != nil {
					panic(err)
				}
				parameters[i] = v
			} else {
				parameters[i] = utils.CoercionNil(pt)
			}
		}

		// debuglog("-730- [flow] Call Event: %v::%v%v().", lcc.page.name, name, contexts)
		lcc.returns = SmartReturn(method.Call(parameters))
		lcc.HandleBreakReturn()
		return true
	} else {
		// debuglog("    - Event Not Found: %v", name)
	}
	return false
}

// CurrentLifecircleControl returns current lcc object from request.
func CurrentLifecircleControl(r *http.Request) (*LifeCircleControl, bool) {
	// get current lcc object from request.
	lcc_obj := context.Get(r, config.LCC_OBJECT_KEY)
	if lcc_obj == nil {
		// panic("LCC is missing in session") // TODO: what to do.
		return nil, false
	}
	lcc := lcc_obj.(*LifeCircleControl)
	if lcc != nil {
		return lcc, true
	}
	return nil, false
}

// CreatePage creates a new page instance with it's control and life. by given type.
func CreatePage(w http.ResponseWriter, r *http.Request, pageT reflect.Type) interface{} {
	if seg := register.GetPage(pageT); seg != nil {
		var newlcc = NewPageFlow(w, r, seg)
		var page = newlcc.current.proton
		page.SetFlowLife(newlcc.current)
		return page
	}
	panic(fmt.Sprintf("Can't create page instance for %v", pageT))
}

// --------------------------------------------------------------------------------

const (
	debugLog = true
)

var (
	emptyParameters = []reflect.Value{}
	emptyString     = ""
)

func debuglog(format string, params ...interface{}) {
	if debugLog {
		log.Printf(format, params...)
	}
}

// ----------------------
// loggers

var logger = logs.Get("IOC:Inject")

var pageflowLogger = logs.Get("GOT:PageFlow")
