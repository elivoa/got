package got

import (
	"fmt"
	"got/core"
)

type Error404 struct {
	A *string
	core.Page
	C     []int
	Error interface{} // should this be type error

}

func (p *Error404) SetupRender() {
	fmt.Println("\n\nPage Error 404.")
}
