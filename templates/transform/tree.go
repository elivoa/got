package transform

import (
	"bytes"
	"fmt"
)

// Struct Tree Node
// TODO: MEM performance issue, here stores 4 copies of template files.
type Node struct {
	tagName string
	attrs   map[string][]byte
	raw     []byte
	html    bytes.Buffer

	level    int
	parent   *Node
	children []*Node
	closed   bool
}

func newNode() *Node {
	return &Node{level: 0, closed: true}
}

func (n *Node) AddChild(node *Node) {
	node.parent = n
	node.level = n.level + 1
	if n.children == nil {
		n.children = make([]*Node, 0, 2)
	}
	n.children = append(n.children, node)
}

// Detach the current node from parent and return it;
func (n *Node) Detach() *Node {
	p := n.parent
	n.parent = nil
	if p == nil {
		return nil
	}
	for i := len(p.children) - 1; i >= 0; i-- {
		// for _, node := range p.children {
		node := p.children[i]
		// fmt.Println("find * ", node)
		if node == n {
			// fmt.Println("matched")
			p.children[i] = nil
			// p.children = append(p.children[:i], p.children[i+1:]...)
			return node
		}
	}
	return nil
}

func (n *Node) String() string {
	cn := 0
	if n.children != nil {
		cn = len(n.children)
	}
	return fmt.Sprintf("c(%v):%v", cn, n.html.String())
}

func (n *Node) Render() string {
	var html bytes.Buffer
	render(&html, n)
	return html.String()
}

func render(html *bytes.Buffer, n *Node) {
	if n == nil {
		return
	}
	html.Write(n.html.Bytes())
	if n.children != nil {
		for _, node := range n.children {
			render(html, node)
		}
	}
}

func (n *Node) GetAttrSafe(attr string) string {
	if nil != n.attrs {
		if attrvalue, ok := n.attrs[attr]; ok {
			return string(attrvalue)
		}

	}
	return ""
}
