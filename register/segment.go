package register

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/elivoa/got/config"
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/logs"
	"github.com/elivoa/got/parser"
	"github.com/elivoa/got/utils"
)

// ----  Identity & Template path  ---------------------------------------------------------------

var pathMap = map[core.Kind]string{
	core.PAGE:      "pages",
	core.COMPONENT: "components",
	core.MIXIN:     "mixins",
}

var identityPrefixMap = map[core.Kind]string{
	core.PAGE:      "p/",
	core.COMPONENT: "c/",
	core.MIXIN:     "x/",
}

var conf = config.Config

// ProtonSegment is a tree like structure to hold path to Page/Component
// 1. Support quick lookup to locate a page or component. (TODO need improve performance)
// 2. Each kind of page has one ProtonSegment instance. (one path)
// TODO
//   - refactor this.
//   - change to SegmentNode and SegmentCore, Note contains parent and chindren,
//     core contains inportant and unique information. Such as IsTemplateLoaded,
//     OR it will fail when parse tempalte.
//
//
type ProtonSegment struct {
	// as a tree node
	Name     string                    // segment name
	Alias    []string                  // alias, e.g.(order/OrderEdit): edit, orderedit
	Parent   *ProtonSegment            //
	Children map[string]*ProtonSegment //
	Level    int                       // depth

	// template related
	IsTemplateLoaded  bool
	Blocks            map[string]*Block         // blocks
	ContentOrigin     string                    // template's html
	ContentTransfered string                    // template's transfered html
	EmbedComponents   map[string]*ProtonSegment // lowercased id
	assets            *core.AssetSet            // js and css files.
	combinedAssets    *core.AssetSet            // Combined it's embed assets.
	TemplateEngine    *core.TemplateEngine      // Embed Template Engine

	// Used for debug, when modified a template file, system will reload the tempalte file as veriosn+1.
	templateVersion          int
	TemplateLastModifiedTime time.Time // template file's last-modified-time

	// TODO: Test Performance: New Method
	//   - Test Perforance between `reflect new` and `native func call`
	//   ? Use Generated New function (e.g. NewSomePage) to create new Page? Is This Faster?
	// TODO: Chagne name
	Proton core.Protoner // The base proton segment. Create new one when installed.

	// associated external resources.
	ModulePackage string             // e.g. got/builtin, syd; used in init.
	StructInfo    *parser.StructInfo // from parser package
	module        *core.Module       // associated Module

	// temp caches. generated template unique id and it's path.
	identity     string // cache identity, default the same name with StructName
	templatePath string // cache template path.

	// TODO - try the method that use channel to lock.
	L sync.RWMutex
}

type Block struct {
	ID                string // block's id
	ContentOrigin     string
	ContentTransfered string
}

func (s *ProtonSegment) AddChild(segname string, seg *ProtonSegment) {
	if s.Children == nil {
		s.Children = map[string]*ProtonSegment{}
	}
	s.Children[strings.ToLower(segname)] = seg
}

func (s *ProtonSegment) HasChild(seg string) bool {
	return s.Children != nil && s.Children[strings.ToLower(seg)] != nil
}

// used to update register
func (s *ProtonSegment) Remove() {
	panic("not implement!")
	// TODO implement this. used in auto reload.
}

// unique identity used as template key.
// TODO refactor all Identities of proton. with event call and event path call.
func (s *ProtonSegment) Identity() string {
	if s.identity == "" {
		s.identity = s.generateIdentity()
	}
	return s.identity
}

func (s *ProtonSegment) generateIdentity() string {
	var term = []string{
		path.Join(identityPrefixMap[s.Proton.Kind()], s.StructInfo.ProtonPath()),
		".",
		s.StructInfo.StructName,
	}
	if s.templateVersion > 0 {
		term = append(term, "^v", strconv.Itoa(s.templateVersion))
	}
	return strings.Join(term, "")
}

func (s *ProtonSegment) IncTemplateVersion() {
	s.templateVersion += 1
	s.identity = s.generateIdentity() // regenerate identity whern version changed.
}

// TemplatePath returns the tempalte key and it's full path.
func (s *ProtonSegment) TemplatePath() (string, string) {
	if s.templatePath == "" {
		module := s.Module()
		if s.templatePath == "" {
			if !strings.HasPrefix(s.StructInfo.ImportPath, module.PackageName) {
				panic("不可能！！！！！")
			}
			filePath := s.StructInfo.ImportPath[len(module.PackageName):]
			s.templatePath = filepath.Join(
				module.BasePath,
				filePath,
				fmt.Sprintf("%v%v", s.StructInfo.StructName, conf.TemplateFileExtension),
			) // TODO Configthis
		}
	}
	return s.Identity(), s.templatePath
}

// Find it's Module
func (s *ProtonSegment) Module() *core.Module {
	if s.module == nil {
		if s.StructInfo != nil {
			module := Modules.Get(s.StructInfo.ModulePackage)
			if module == nil {
				panic(fmt.Sprint("Can't find module for ", s.StructInfo.ModulePackage))
			}
			s.module = module
		}
	}
	return s.module
}

// ________________________________________________________________________________
// parse url and put it into segments.
// return what?
// TODO return [][]string:
//   order, list
//   order, orderlist
//   order/create/OrderCreateDetail
// TODO: alias not correct.
//
func (s *ProtonSegment) Add(si *parser.StructInfo, p core.Protoner) [][]string {

	// TODO segment has structinfo
	src := si.ModulePackage
	segments := strings.Split(si.ProtonPath(), "/")
	if len(segments) > 0 && segments[0] == "" {
		segments = segments[1:]
	}
	segments = append(segments, si.StructName)

	// if si.StructName == "Status" {
	// 	fmt.Println("++++++++++++++++++++++++", si.StructName)
	// }

	dlog("-___- [Register %v] %v::%v", pathMap[p.Kind()], src, segments)
	dlog("-___- [%v's URL is] %v", pathMap[p.Kind()], si)

	// add to registerc
	var (
		currentSeg     = s             // always use root segment.
		prevSeg        = "//nothing//" // previous lowercase seg
		prevSegs       = []string{}    // previous lowercase seg[]
		selectorPrefix = []string{}    // tempvalue
	)

	// 1. process path segments to reach the end, without last node.
	for idx, seg := range segments[0:(len(segments) - 1)] {
		var lowerSeg = strings.ToLower(seg)

		var segment *ProtonSegment
		if currentSeg.HasChild(seg) {
			segment = currentSeg.Children[seg]
			// TODO detect conflict
		} else {
			// Add path to segment.
			segment = &ProtonSegment{
				Name:   seg,
				Parent: currentSeg,
				Level:  idx,
			}
			// dlog("!!!! add path to structure: seg: %v\n", seg) // ------------------
			currentSeg.AddChild(segment.Name, segment)
		}
		currentSeg = segment
		selectorPrefix = append(selectorPrefix, seg)
		prevSeg = lowerSeg
		prevSegs = append(prevSegs, lowerSeg)
	}

	// 2. process last node
	// ~ 2 ~ overlapped keywords: i.e.: order/OrderEdit ==> order/edit
	var (
		seg           string   = segments[len(segments)-1]
		lowerSeg      string   = strings.ToLower(seg)
		shortLowerSeg string   = lowerSeg
		shortSeg      string   = seg           // shorted seg with case
		finalSegs     []string = []string{seg} // alias
	)

	// fmt.Printf("-www- [RegisterPage] enter last node %v; \n", seg)
	// fmt.Printf("-www- --------- %v; %v \n", lowerSeg, prevSeg)

	// Match origin paths: /order/create/OrderCreateIndex, in this example we can ignore
	// prefix 'Order' and the ignore 'Create', and 'Index' can be automatically ignored.
	for _, p := range prevSegs {
		// dlog("prevSegs  =  ", prevSegs)
		// dlog("+++ strings.HasPrefix: %v, %v = %v\n", p, shortLowerSeg, strings.HasPrefix(shortLowerSeg, p))
		if strings.HasPrefix(shortLowerSeg, p) {
			shortSeg = strings.TrimSpace(shortSeg[len(p):])
			shortLowerSeg = strings.TrimSpace(shortLowerSeg[len(p):])
			if shortSeg != "" && shortSeg != p { // e.g. order/create/OrderCreateIndex
				finalSegs = append(finalSegs, shortSeg)
				dlog("!!!!!!!!!!!!!! p:%v, add to final: %v >> %v\n", p, shortSeg, finalSegs)
			}
		}
	}

	// TODO remove suffix in the same way.
	// TODO kill this.
	// /order/Create[Order] - ignore [Order]
	if strings.HasSuffix(lowerSeg, prevSeg) {
		dlog("+++ Match suffix, \n")
		s := seg[:len(seg)-len(prevSeg)]
		if s != "" {
			finalSegs = append(finalSegs, s) // ?? nothing reach here?
			dlog("!!!!!!!!!!!!!! [SUFFIX] p:%v, add to final: %v >> %v\n", p, shortSeg, finalSegs)
		} else {
			// fallback TODO
			// currentSeg.Src = src
			currentSeg.Proton = p
		}
	}

	// judge empty/index

	// remove index if any
	// /order/Order[Index] - fall back to /order/
	// /api/suggest/Suggest - fall back to /api/suggest
	if p.Kind() == core.PAGE && strings.HasSuffix(shortLowerSeg, "index") {
		// dlog("+++ Match Index, \n")
		var trimlen = len(shortLowerSeg) - len("index")
		if trimlen >= 0 {
			shortSeg = shortSeg[:trimlen]
			shortLowerSeg = shortLowerSeg[:trimlen]
		}
	}

	// Fallback if needed.
	if shortSeg == "" {
		// e.g.: /api/suggest/Suggest --> /api/suggest
		// dlog("+++++ Fallback.\n") // ------------------------------------------
		currentSeg.Proton = p
		currentSeg.StructInfo = si
	} else {
		if shortSeg != seg && (len(finalSegs) <= 1 || shortSeg == finalSegs[len(finalSegs)-2]) {
			// e.g.: /order/OrderDetailIndex --> /order/detail
			finalSegs = append(finalSegs, shortSeg)
			dlog("!!!!!!!!!!!!!! [NO-fallback] add-to-final: %v;; seg:%s \n", finalSegs, seg)
		}
	}

	dlog(">>>>> FinalSegs: %v\n", finalSegs) // ------------------------------------------
	finalSegs = utils.SortStringByLength(finalSegs)
	dlog(">>>>> FinalSegs: %v\n", finalSegs) // ------------------------------------------

	// 4. finally add segment struct to chains.
	segment := &ProtonSegment{
		Name:       finalSegs[0], // Name is a bitch.
		Alias:      finalSegs,
		Parent:     currentSeg,
		Level:      len(segments) - 1,
		Proton:     p,
		StructInfo: si,
	}

	selectors := [][]string{} // finally return this to register components.

	for _, s := range finalSegs {
		currentSeg.AddChild(s, segment) // link segment together.

		// add selector
		selector := []string{}
		selector = append(selector, selectorPrefix...)
		selector = append(selector, s)
		selectors = append(selectors, selector)
	}

	// add the first segment into typemap
	switch segment.Proton.Kind() {
	case core.PAGE:
		PageTypeMap[reflect.TypeOf(p).Elem()] = segment
	case core.COMPONENT:
		ComponentTypeMap[reflect.TypeOf(p).Elem()] = segment
	case core.MIXIN:
		MixinTypeMap[reflect.TypeOf(p).Elem()] = segment
	}

	// dlog(">>>>> Selectors: %v\n", selectors)
	return selectors
}

// ----  Lookup & Results  ------------------------------------------------------------------------------

type LookupResult struct {
	Segment        *ProtonSegment
	ComponentPaths []string // component path ids, for calling event.
	EventName      string
	Parameters     []string // parameters reconized.
}

func (lr *LookupResult) IsEventCall() bool {
	return lr.EventName != ""
}

func (lr *LookupResult) IsValid() bool {
	if lr.Segment != nil && lr.Segment.Proton != nil {
		return true
	}
	return false
}

func (lr *LookupResult) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(">> [LookupResult]{\n"))
	buffer.WriteString(fmt.Sprintf("\tSegment:%v,\n", lr.Segment))
	// buffer.WriteString(fmt.Sprintf("\tPageUrl:%v,\n", lr.PageUrl))
	buffer.WriteString(fmt.Sprintf("\tComponentPaths:%v,\n", lr.ComponentPaths))
	buffer.WriteString(fmt.Sprintf("\tEventName:%v,\n", lr.EventName))
	buffer.WriteString(fmt.Sprintf("\tParameters:%v,\n", lr.Parameters))
	buffer.WriteString("  }\n")
	return buffer.String()
}

var average_lookup_time int

var lookupLogger = logs.Get("URL Lookup")

// Lookup the structure, find the right page or component.
// Can detect event calls, event calls on embed components.
// TODO performance
// 例如： /got/Status.TemplateStatus:TemplateDetail/c__got:TemplateStatus
// 当遇到第一个.的时候，后面的为components. 当再遇到：的时候，后面的是方法名，/截断作为参数。
func (s *ProtonSegment) Lookup(url string) (result *LookupResult, err error) {
	// pre-process url
	trimedUrl := strings.Trim(url, " ")
	if !strings.HasSuffix(trimedUrl, "/") {
		trimedUrl += "/"
	}

	if lookupLogger.Debug() {
		lookupLogger.Printf("[Lookup] URL: '%v' (trimmed)", url)
		if lookupLogger.Trace() {
			lookupLogger.Printf("[Lookup] Trimmed URL: '%v'", trimedUrl)
		}
	}

	var (
		level         int = -1
		buffer        bytes.Buffer
		segments           = []string{}
		parameterPart bool = false
	)
	result = &LookupResult{
		ComponentPaths: []string{}, // init component paths.
		Parameters:     []string{},
	}

	segment := s // loop channel object
	for _, c := range trimedUrl {
		switch c {
		default:
			buffer.WriteRune(c)
			continue
		case '/':
			level += 1
		}

		if lookupLogger.Trace() {
			lookupLogger.Printf("[Lookup] for url segment: '%v'", buffer.String())
		}

		// arrive here means words finished. process segment
		seg := buffer.String()
		segments = append(segments, seg)
		buffer.Reset()

		// skip the first / segment.
		if level == 0 && seg == "" {
			continue
		}

		if lookupLogger.Debug() {
			lookupLogger.Printf("[Lookup] Step: Level %v Seg:[ %-10v ] segment:[ %-20v ]\n",
				level, seg, segment)
		}

		// parameter mode
		if parameterPart {
			result.Parameters = append(result.Parameters, seg)
			continue
		}

		// parth lookup mode

		// If contains ":", this is an event call. Or parameter.
		// and match stops here, others are parameters of event.
		if index := strings.Index(seg, ":"); index > 0 {
			result.EventName = seg[index+1:]
			array := strings.Split(seg[0:index], ".")
			seg = strings.ToLower(array[0])
			result.ComponentPaths = array[1:]

			level = level + 1
			parameterPart = true
		}

		if segment.Children == nil || len(segment.Children) == 0 || !segment.HasChild(seg) {
			if lookupLogger.Debug() {
				lookupLogger.Printf("- - - [Lookup] match finished.")
			}
			// Match finished, this must be the first paramete
			result.Parameters = append(result.Parameters, seg)
			parameterPart = true
			// break				//
		} else {
			// fmt.Println("going into next step: ", segment)
			segment = segment.Children[strings.ToLower(seg)]
		}

	}

	// get page url
	// pageUrl := strings.Join(segments[:level], "/")
	// if result.EventName != "" {
	// 	// TODO: bugs here. can', .
	// 	index := strings.LastIndex(pageUrl, ".")
	// 	fmt.Println("................", index, " >>> ", pageUrl)
	// 	pageUrl = pageUrl[:index]
	// }
	// log.Printf("- - - [Lookup] 'pageurl is' %v  (including event)\n", pageUrl)

	if nil == segment {
		err = errors.New("Lookup Failed.")
	}
	result.Segment = segment
	// result.PageUrl = pageUrl
	if lookupLogger.Debug() {
		// lookupLogger.Printf("- - - [Lookup] Result is %v", result)
	}
	return
}

// asset helper
func (s *ProtonSegment) Assets() *core.AssetSet {
	return s.assets
}

func (s *ProtonSegment) SetAssets(assetset *core.AssetSet) {
	s.assets = assetset
	s.combinedAssets = nil // clear cached combined assets.
}

// Get CombinedAssets, initialize it first.
func (s *ProtonSegment) CombinedAssets() *core.AssetSet {
	if nil == s.combinedAssets {
		var combinedAssets = core.NewAssetSet()
		_combine(s, combinedAssets, 0)
		s.combinedAssets = combinedAssets
	}
	return s.combinedAssets
}

func _combine(currentSeg *ProtonSegment, combinedAssets *core.AssetSet, depth int) {
	if depth >= 20 {
		panic("Reach max depth when loop EmbedComponents!")
	}
	// do things
	combinedAssets.CombineAssets(currentSeg.Assets())
	if nil != currentSeg.EmbedComponents {
		for _, es := range currentSeg.EmbedComponents {
			_combine(es, combinedAssets, depth+1)
		}
	}
}

/* ________________________________________________________________________________
   Print Helper
*/

func (s *ProtonSegment) String() string {
	length, path := ".", "--"
	if len(s.Children) > 0 {
		length = fmt.Sprint(len(s.Children))
	}
	if s.StructInfo != nil {
		path = s.StructInfo.ImportPath
	}
	return fmt.Sprintf("N%d%v (%v)[%v]", len(s.Alias), s.Alias, length, path)
	// return fmt.Sprintf("%-20v (%v)[%v]", s.Name, length, path)
}

func (s *ProtonSegment) ToString(name string) string {
	length, path := ".", "--"
	if len(s.Children) > 0 {
		length = fmt.Sprint(len(s.Children))
	}
	if s.StructInfo != nil {
		path = s.StructInfo.ImportPath
	}
	return fmt.Sprintf("%-20v (%v)[%v]", name, length, path)
	// return fmt.Sprintf("%-20v (%v)[%v]", s.Name, length, path)
}

// print all details
func (s *ProtonSegment) PrintALL() {
	s.print(s)
}

func (s *ProtonSegment) StringTree(newline string) string {
	var out bytes.Buffer // = bytes.NewBuffer([]byte{})
	s.treeSegment(&out, "root", s, newline)
	return out.String()
}

func (s *ProtonSegment) treeSegment(out *bytes.Buffer, segmentName string, segment *ProtonSegment, newline string) {
	out.WriteString(fmt.Sprintf("+ %v >> %v ||<strong>%s</strong> %s",
		segment.ToString(segmentName), segment.StructInfo, segment.EmbedComponents, newline))
	for segName, seg := range s.Children {
		for i := 0; i <= seg.Level; i++ {
			out.WriteString("  ")
		}
		if seg != nil {
			seg.treeSegment(out, segName, seg, newline)
		}
	}
}

// TODO: user treeSegment instead.
func (s *ProtonSegment) print(segment *ProtonSegment) string {
	fmt.Printf("+ %v >> %v\n", segment, segment.StructInfo)
	for _, seg := range s.Children {
		for i := 0; i <= seg.Level; i++ {
			fmt.Print("  ")
		}
		if seg != nil {
			seg.print(seg)
		}
	}
	return ""
}

/* ________________________________________________________________________________
   TrimPathSegments
   Return: src, segments
   Param:
     protonType - [page|component]

   e.g. f("/got/builtin/pages/order/list") = "order/list/"

*/
func trimPathSegments(url string, protonType string) (string, []string) {
	segments := strings.Split(url, "/")
	var index, seg = 0, ""
	for index, seg = range segments {
		if seg == protonType {
			break
		}
	}
	var src string = ""
	if index > 0 {
		src = strings.Join(segments[0:index], "/")
	}
	return src, segments[index+1:]
}

// --------------------------------------------------------------------------------
// Tools & helper methods
// --------------------------------------------------------------------------------

var debug_add = true

func dlog(format string, params ...interface{}) {
	if debug_add {
		log.Printf(format, params...)
	}
}
