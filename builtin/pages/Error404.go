package got

import (
	"fmt"
	"got/core"
)

type Error404 struct {
	A *string
	core.Page
	C []int
}

func (p *Error404) SetupRender() {
	fmt.Println("\n\nPage Error 404.")
}
