package got

import (
	"fmt"
	"github.com/elivoa/got/core"
)

type Errors struct {
	A *string
	core.Page
	C []int
}

func (p *Errors) SetupRender() {
	fmt.Println("\n\nPage Error page")
}


