package got

import (
	"fmt"
	"got/core"
)

type Error500 struct {
	A *string
	core.Page
	C []int
}

func (p *Error500) SetupRender() {
	fmt.Println("\n\nPage Error 500.")
}
