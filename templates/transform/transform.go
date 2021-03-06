/*
Transform tapestry like html page into go-template like ones. Keep it functions well.

  Time-stamp: <[transform.go] Elivoa @ Monday, 2016-04-11 23:53:32>
  TODO: remove this package.
  TODO: Doc this well.
  TODO: Error Report: add line and column when error occured.

Tapestry template like components:
  <t:a href="chedan" />



*/
package transform

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"

	"github.com/elivoa/got/cache"
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/register"
	"golang.org/x/net/html"
)

// ---- Transform template ------------------------------------------

// 同一个Page或者Component应该使用同一个Transformer
type Transformater struct {
	tree   *Node            // root node
	z      *html.Tokenizer  // tokenizer
	blocks map[string]*Node // blocks in this tempalte

	// status tag
	tag_2nd_parse_in_import bool // true if in t:import block

	// output!!
	Components map[string]*ComponentInfo // Components' Id --> ComponentInfo

	Assets *core.AssetSet // assets

	// outdated. 值如果是-1， 说明这个是通过t:id的方式指定的ID，不允许重复。
	ComponentCount map[string]int
}

type ComponentInfo struct {
	Name        string // component name
	Segment     *register.ProtonSegment
	ID          string // component ID (How to deal with component that in a loop.)
	Index       int    // how many this components in one page.
	IDSpecified bool
}

func NewTransformer() *Transformater {
	return &Transformater{
		Assets:         core.NewAssetSet(),
		Components:     map[string]*ComponentInfo{},
		ComponentCount: map[string]int{},
	}
}

/*
  Transform tempalte fiels. functions:
  translate <t:some_component ... /> into {{t_some_component ...}}

TODO:
  . Support t:block
  . Range Tag

TODOs:
---- 1 --------------------------------------------------------------------------------
<div t:type="xx"... >some <b>bold text</b></div>
 there will remaining: some meaningful text</div>
 now I ignore these, TODO make this a block and render it.

---- N --------------------------------------------------------------------------------
*/
var compressHtml bool = true

func (t *Transformater) Parse(reader io.Reader, isPage bool) *Transformater {
	z := html.NewTokenizer(reader)
	t.z = z

	// the root node
	root := newNode() // &Node{level: 0}
	t.tree = root
	parent := root

	t.blocks = map[string]*Node{} // init for blocks;

	for {
		tt := z.Next()

		// new the current node.
		node := newNode()

		// after call something all tag is lowercased. but here with case.
		zraw := z.Raw()
		node.raw = make([]byte, len(zraw))
		copy(node.raw, zraw[:])
		zraw = node.raw

		// start parse
		switch tt {
		case html.TextToken:
			// here may contains {{ }}
			if compressHtml {
				node.html.Write(t.TrimTextNode(zraw, node)) // trimed spaces
			} else {
				node.html.Write(zraw)
			}
			parent.AddChild(node)

		case html.StartTagToken:
			node.closed = false
			if b := t.processStartTag(node); !b {
				node.html.Write(zraw)
			}
			// switch node.tagName {
			// case "input", "br", "hr", "link":
			// 	parent.AddChild(node)
			// default:
			parent.AddChild(node)
			parent = node // go in
			// }

		case html.SelfClosingTagToken:
			if b := t.processStartTag(node); !b {
				node.html.Write(zraw)
			}
			parent.AddChild(node)

		case html.EndTagToken:
			k, _ := z.TagName()
			tag := string(k)
			switch tag {
			case "range", "with", "if":
				node.html.WriteString("{{end}}")
			case "hide":
				node.html.WriteString("*/}}")
			case "t:import", "t:block":
				// append nothing, only remove </xxx> tag;
			case "head":
				if isPage {
					if err := t.transformComponent(
						node, []byte("PageHeadBootstrap"), []byte("span"), nil); err != nil {
						panic(err)
					}
				}
				// At the end of body, append a component to process page bootstrap things.
				node.html.Write(zraw)
			case "body":
				// TODO append page-final-bootstrap component:
				if err := t.transformComponent(
					node, []byte("PageFinalBootstrap"), []byte("span"), nil); err != nil {
					panic(err)
				}
				// At the end of body, append a component to process page bootstrap things.
				node.html.Write(zraw)
			default:
				node.html.Write(zraw)
			}
			// TODO: process unclosed tag.
			// if has unclosed tag, just unclose it.
			// find the right tag and close, move wrong tag back.
			if tag == parent.tagName {
				parent.AddChild(node)
				parent.closed = true
				parent = parent.parent
			} else {

				node.parent = parent // only set parent will not link the node to the tree.
				temp := node
				for {
					// if true{break}
					if temp == nil {
						panic(fmt.Sprintf("Tag %v not closed!", temp))
					}
					temp = temp.parent
					if tag == temp.tagName {
						temp.AddChild(node)
						parent = temp.parent
						temp.closed = true
						break
					} else {
						if temp.children != nil {
							// tp := []*Node{}
							for _, c := range temp.children {
								// fmt.Println("      > move <<< ", c.tagName, ";", c.html.String(), ">>>")
								c.Detach()
								temp.parent.AddChild(c)
								temp.closed = true
							}
						}
					}
				}
			}

		// case html.CommentToken: // ignore all comments
		// case html.DoctypeToken: // ignore
		// case html.DoctypeToken: // node.html.Write(zraw)
		case html.ErrorToken:
			if z.Err().Error() == "EOF" { // END parsing template.

				// the second step parsing tempalte; parse t:bock and t:imports;
				t.parseBlocks()

				return t
			} else {
				panic(z.Err().Error())
			}
		default:
			node.html.Write(zraw)
			parent.AddChild(node)
		}
	}
}

// processing every start tag()
// return 1.
//   - true if already write to buffer.
//   - false if need to write Raw() to buffer.
//   2. tagNamep
// Note: go.net/html package lowercased all values,
//
//
func (t *Transformater) processStartTag(node *Node) bool {
	// collect information
	bname, hasAttr := t.z.TagName()
	node.tagName = string(bname) // performance

	var (
		iscomopnent   bool
		componentName []byte
		elementName   []byte
		err           error
	)

	// start with componentp
	if len(bname) >= 2 && bname[0] == 't' && bname[1] == ':' {
		iscomopnent = true
		componentName = bname[2:]
	}

	// another kind of component
	var attrs map[string][]byte
	if hasAttr {
		attrs = map[string][]byte{}
		for {
			key, val, more := t.z.TagAttr()
			if len(key) == 6 && bytes.Equal(key, []byte("t:type")) {
				iscomopnent = true
				componentName = val
				elementName = bname
				// ignore t:type attr
			} else {
				attrs[string(key)] = val // = append(attrs, [][]byte{key, val})
			}
			if !more {
				break
			}
		}
		node.attrs = attrs
	}

	if iscomopnent {
		if err = t.transformComponent(node, componentName, elementName, attrs); err == nil {
			return true
		}
	}

	// --------------------------------------------------------------------------------
	// not a component, process if tag is command
	switch string(bname) {
	case "range":
		t.renderRange(node, attrs)
	case "if":
		t.renderIf(node, attrs)
	case "else":
		node.html.WriteString("{{else}}")
	case "hide":
		node.html.WriteString("{{/*")
	case "t:import", "t:block":
		// process these blocks in the next step;
	// case "t:delegate":
	// 	t.renderDelegate(node, attrs)
	default:
		if err != nil {
			panic(err)
		}
		return false
	}
	return true
}

// parse blocks, parse imports 需要解析第二遍；
// TODO 解析t:imports
func (t *Transformater) parseBlocks() {
	t.blocks = map[string]*Node{}
	t._parseBlocks(t.tree)
	// fmt.Println("\ndebug info { ---- [DEBUG: IMPORTS] ---------------------------------------------------- {")
	// if nil != t.blocks {
	// 	for k, v := range t.blocks {
	// 		fmt.Println(k, "  --  >  ", v)
	// 	}
	// }
	// fmt.Println("} // DEBUG: IMPORTS\n")
}

func (t *Transformater) _parseBlocks(n *Node) {
	if n == nil {
		return
	}
	// import 里面不允许任何其他类型的tag除了 link, style, 也不只允许一层tag.
	if t.tag_2nd_parse_in_import == true {
		// in t:import, parse script links and forbid orther tags.
		switch n.tagName {
		case "script":
			// TODO 这里暂时忽略他们的type；所有url都根据url去重；
			t.Assets.AddScripts(&core.Script{
				Type: n.GetAttrSafe("type"),
				Src:  n.GetAttrSafe("src"),
			})
		case "link":
			t.Assets.AddStyleSheet(&core.StyleLink{
				Type: n.GetAttrSafe("type"),
				Rel:  n.GetAttrSafe("src"),
				Href: n.GetAttrSafe("href"),
			})
		case "": // ignored
		default:
			panic(fmt.Sprintf("Template Structure Error: '%s' are not allowed in t:import.", n.tagName))
		}
	} else {
		// not in t:import, do block parse and import-enter.
		if n.tagName == "t:block" { // parse block
			t._secondparse_block(n)
			return // not go deeper, just get block and return;
		}
		if n.tagName == "t:import" { // enter t:import
			t.tag_2nd_parse_in_import = true
			// Go deeper
			if n.children != nil {
				for _, node := range n.children {
					t._parseBlocks(node)
				}
			}
			t.tag_2nd_parse_in_import = false // return back
		} else {
			// normal tag go deeper.
			if n.children != nil {
				for _, node := range n.children {
					t._parseBlocks(node)
				}
			}
		}

	}

}

// extract blocks in templates' structure tree.
func (t *Transformater) _secondparse_block(n *Node) {
	var foundId bool = false
	var id string
	if n.attrs != nil {
		for k, v := range n.attrs {
			// TODO add another parameters;
			if strings.ToLower(k) == "id" {
				id = string(v)
				foundId = true
				break
			}
		}
	}
	if !foundId {
		panic("Can't find `id` attribute in t:block tag!")
	}

	// check id conflict
	if _, ok := t.blocks[id]; ok {
		panic(fmt.Sprintf("Block ID Conflict, ID: %s", id))
	} else {
		t.blocks[id] = n.Detach()
	}
}

func (t *Transformater) renderDelegate(node *Node, attrs map[string][]byte) {
	node.html.Write(node.raw)
}

func (t *Transformater) builtinComponentFunction(name string) func(*Node, map[string][]byte) {
	switch name {
	case "range":
		return t.renderRange
	default:
		panic(fmt.Sprintf("Builtin component %v not found!", name))
	}
}

func (t *Transformater) renderRange(node *Node, attrs map[string][]byte) {
	node.html.WriteString("{{range ")
	if nil != attrs {
		if _var, ok := attrs["var"]; ok {
			node.html.Write(_var)
			node.html.WriteString(":=")
		}
		if source, ok := attrs["source"]; ok {
			node.html.Write(source)
		}
	}
	node.html.WriteString("}}")
}

func (t *Transformater) renderIf(node *Node, attrs map[string][]byte) {
	node.html.WriteString("{{if ")
	if nil != attrs {
		var (
			_var []byte
			ok   bool
		)
		if _var, ok = attrs["t"]; !ok {
			if _var, ok = attrs["test"]; !ok {
				panic("`If` must have attribute test or t!")
			}
		}
		node.html.Write(_var)
	}
	node.html.WriteString("}}")
}

// Processing Components
func (t *Transformater) transformComponent(node *Node, componentName []byte, elementName []byte,
	attrs map[string][]byte) error {

	// lookup component and get StructInfo
	lookupurl := strings.Replace(string(componentName), ".", "/", -1)
	lr, err := register.Components.Lookup(lookupurl)
	if err == nil && (lr.Segment == nil || lr.Segment.Proton == nil) {
		err = errors.New(fmt.Sprintf("Can't find component for %v", string(componentName)))
	}
	if err != nil {
		return err
	}

	// create cache.StructInfo
	sc := cache.StructCache
	si := sc.GetCreate(reflect.TypeOf(lr.Segment.Proton), core.COMPONENT)

	// fmt.Println("\n\n----------------------------------------------------------------------------")
	// fmt.Printf("Find Component %s , parameters: \n", string(componentName))
	// for idx, v := range attrs {
	// 	fmt.Printf("\t%s := %v\n", idx, string(v))
	// }

	/*
		For ComponentID， 使用 tid 这种方式指定ID的，ID不能重复。
		其他情况下使用Component的名字来命名。这种情况下允许重复，ID直接累加。
	*/
	var (
		componentId     string
		hardSpecifiedId bool
	)

	if id, ok := attrs["tid"]; ok {
		componentId = string(id)
	} else if id, ok := attrs["t:id"]; ok {
		componentId = string(id)
	}

	//// loop version.
	// for key, val := range attrs {
	// 	if strings.ToLower(key) == "tid" || strings.ToLower(key) == "t:id" {
	// 		componentId = string(val)
	// 		break
	// 	}
	// }

	if componentId == "" {
		// occupy id if not specified.
		componentId = lr.Segment.StructInfo.StructName
		hardSpecifiedId = false
	} else {
		hardSpecifiedId = true
	}

	var (
		count  int
		ok     bool
		realId string = componentId
	)
	if count, ok = t.ComponentCount[componentId]; !ok {
		count = -1
	}

	// generate component id
	if count >= 0 {
		realId = fmt.Sprintf("%s_%d", componentId, count)
	}
	// fmt.Println("\n----------------------------------------------------")
	// fmt.Println("count is", count, "; componentId is ", realId)

	if _, ok := t.Components[realId]; !ok {
		t.Components[realId] = &ComponentInfo{
			Name:        string(componentName),
			Segment:     lr.Segment,
			ID:          realId,
			Index:       count,
			IDSpecified: hardSpecifiedId,
		}
	} else {
		// if find duplicated. return error.
		panic(fmt.Sprintf("ID Duplicated %s.", realId))
	}
	t.ComponentCount[componentId] = count + 1 // write back count

	// --------------------------------------------------------------------------------
	// write template back
	node.html.WriteString("{{t_")
	node.html.WriteString(strings.Replace(lookupurl, "/", "_", -1))
	node.html.WriteString(" $") // 1st param: it's container
	node.html.WriteString(" `")
	node.html.WriteString(realId) // 2nd param: unique id in component scope.
	node.html.WriteString("`")

	// elementName
	if elementName != nil {
		node.html.WriteString(" \"elementName\" `")
		node.html.Write(elementName)
		node.html.WriteString("`")
	}

	if attrs != nil {
		for key, val := range attrs {
			// write key. e.g.: "ParameterName"
			node.html.WriteString(" \"")
			// get which is cached.
			fi := si.FieldInfo(key)
			if fi != nil {
				node.html.WriteString(fi.Name)
			} else {
				node.html.WriteString(key)
			}
			node.html.WriteString("\" ")

			// write value, with autodetected transform.
			t.appendComponentParameter(&node.html, val)
		}
	}
	node.html.WriteString("}}")
	return nil
}

// Auto-detect literal or functional
// if value starts from . or $ , treate this as property. others as string
//
// Value transform: for name="_some_value_", we transform it into:
//   ~ before ~           ~ after ~             ~ note ~
//   ".Name"              .Name                // start form . or $
//   "literal:....."      "...."               // literal prefix
//	 "abcd"               "abcd"               // auto detect as plan text
//	 ".Name+'_'+.Id"      (print .Name '_' .Id)// special string join
//   "/xxx/{{.ID}}"       (print "/xxx/" .Id)  // special string format
//   "{refer .}"          refer .              // not string.
//
//
//   TODO support more prefix...
//
func (t *Transformater) appendComponentParameter(buffer *bytes.Buffer, val []byte) error {
	val = bytes.TrimSpace(val)
	switch {
	case len(val) > 0 && (val[0] == '.' || val[0] == '$' || val[0] == '('):
		buffer.Write(val)

	case len(val) > 5 && bytes.Equal(val[0:5], []byte("print")):
		buffer.WriteString("(")
		buffer.Write(val)
		buffer.WriteString(")")

	case len(val) > 8 && bytes.Equal(val[0:8], []byte("literal:")):
		buffer.WriteString(" \"")
		buffer.Write(bytes.Replace(val[8:], []byte{'"'}, []byte{'\\', '"'}, 0))
		buffer.WriteString("\"")

	case len(val) > 1 && (val[0] == '{' && val[1] != '{' && val[len(val)-1] == '}'):
		buffer.WriteString("(")
		buffer.Write(val[1 : len(val)-1])
		buffer.WriteString(")")

	// case len(val) > 0 && (val[0] == '[' && val[len(val)-1] == ']'):
	// TODO array of value.

	case printValueRegex.Match(val): // if is "/xxx/{{.ID}}"
		result := printValueRegex.FindSubmatch(val)
		// for _, r := range result {
		// 	fmt.Println(r)
		// }
		if len(result) == 3 { // translate to (print "/xxx/" .ID)
			buffer.WriteString(" (print \"")
			buffer.Write(result[1])
			buffer.WriteString("\" ")
			buffer.Write(result[2])
			buffer.WriteString(")")
		}
	default:
		buffer.WriteString(" \"")
		buffer.Write(bytes.Replace(val, []byte{'"'}, []byte{'\\', '"'}, 0))
		buffer.WriteString("\"")
	}
	return nil
}

// Render template to string
func (t *Transformater) RenderToString() string {
	return t.tree.Render()
}

// Render blocks to map.p
func (t *Transformater) RenderBlocks() map[string]string {
	if t.blocks != nil {
		returns := map[string]string{}
		for blockId, node := range t.blocks {
			returns[blockId] = node.Render()
		}
		return returns
	}
	return nil
}

// --------------------------------------------------------------------------------

// variables
var printValueRegex, _ = regexp.Compile("^(.*){{(.*)}}$")

// ---- utils --------------------------------------------------------------------------------

// TODO trim node function not finished.
func (t *Transformater) TrimTextNode(text []byte, node *Node) []byte {

	var debug = false
	if debug {
		fmt.Printf("++ space ++: [%s] --> ", strings.Replace(string(text), "\n", "\\n", -1))
	}

	var (
		firstValidCharacter int  = -1
		hasUsefulCharacters bool = false
		hasLeftSpace        bool = false
		hasLeftNewLine      bool = false
		hasRightNewLine     bool = false
		hasRightSpace       bool = false
		lastValidCharacter  int  = -1
	)

	// from left
	var conti = true
LOOP:
	for idx, b := range text {
		switch b {
		case ' ', '\t':
			if conti {
				hasLeftSpace = true
			}
		case '\r', '\n':
			conti = false
			hasLeftNewLine = true
		default: // has other c haracters
			hasUsefulCharacters = true
			firstValidCharacter = idx
			break LOOP
		}
	}
	if hasUsefulCharacters {
	LOOP2:
		for i := len(text) - 1; i >= 0; i-- {
			switch text[i] {
			case ' ', '\t':
				hasRightSpace = true
			case '\r', '\n':
				hasRightSpace = true
				hasRightNewLine = true
			default: // has other characters
				// fmt.Println(">>> ", i, text[i], string(text[i]), lastValidCharacter)
				lastValidCharacter = i
				break LOOP2
			}

		}
	}
	if hasRightNewLine {
		// fmt.Println("firstValidCharacter is ", firstValidCharacter, lastValidCharacter, hasRightNewLine)
	}
	// result
	var result bytes.Buffer
	if hasUsefulCharacters {
		if hasLeftNewLine {

		}
		if hasLeftSpace {
			// TODO \n
			result.WriteByte(' ')
		}

		// fmt.Printf("<<text[firstValidCharacter:lastValidCharacter]= text[%d:%d]=%s >> // %s",
		// 	firstValidCharacter, lastValidCharacter,
		// string(text[firstValidCharacter:lastValidCharacter]), string(text))

		result.Write(text[firstValidCharacter : lastValidCharacter+1])
		if hasRightSpace {
			result.WriteRune(' ')
		}
		if hasRightNewLine {
			result.WriteRune('\n')
		}
		if debug {
			fmt.Printf("--1: [%s]\n", strings.Replace(result.String(), "\n", "\\n", -1))
		}
		return result.Bytes()
	} else {
		if hasLeftSpace {
			result.WriteRune(' ')
		}
		if hasLeftNewLine {
			result.WriteRune('\n')
		}
		if debug {
			fmt.Printf("--2: [%s]\n", strings.Replace(result.String(), "\n", "\\n", -1))
		}
		return result.Bytes()
	}

	// return bytes.TrimSpace(text) //, " \r\n")
}
