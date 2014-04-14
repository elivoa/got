package got

import (
	"fmt"
	"github.com/elivoa/got/config"
	"github.com/elivoa/got/util"
)

func init() {
	fmt.Println("got system started")
	fmt.Println("Imports: ", config.LCC_OBJECT_KEY)
	fmt.Println("", util.GetCurrentPath(0))
}
