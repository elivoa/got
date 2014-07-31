package got

import (
	"fmt"
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/debug"
	"html/template"
)

type Error500 struct {
	A *string
	core.Page
	C     []int
	Error interface{} // should this be type error
}

func (p *Error500) SetupRender() {
	fmt.Println("\n\nPage Error 500.")
}

func (p *Error500) Stack() template.HTML {
	var str template.HTML
	if e, ok := p.Error.(error); ok {
		str = template.HTML(debug.StackString(e))
	} else if s, ok := p.Error.(string); ok {
		str = template.HTML(debug.StackString(fmt.Errorf(s)))
	} else {
		str = template.HTML(fmt.Sprint(p.Error))
	}
	return str
	// buf := make([]byte, 1<<16)
	// length := runtime.Stack(buf, false)
	// // fmt.Printf("%s", buf)
	// str := string(buf[:length])
	// str = strings.Replace(str, "\n", "<br>", -1)
	// str = strings.Replace(str, "\t", "&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;", -1)
	// return template.HTML(str)
}
