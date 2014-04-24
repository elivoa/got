package got

import (
	"fmt"
	"got/core"
)

type PermissionDenied struct {
	A *string
	core.Page
	C     []int
	Error interface{} // should this be type error
}

func (p *PermissionDenied) SetupRender() {
	fmt.Println("\n\n Permission Denied Error!")
}
