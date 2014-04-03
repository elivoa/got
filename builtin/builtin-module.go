/*
GOT Module: builtin package
*/

package builtin

import (
	"github.com/elivoa/got/builtin/pages/got/fileupload"
	"got/register"
	"got/utils"
	"net/http"
)

var BuiltinModule = &register.Module{
	Name:        "github.com/elivoa/got/builtin",
	BasePath:    utils.CurrentBasePath(),
	PackagePath: "github.com/elivoa/got/builtin",
	Description: "GOT Framework Built-in pages and components etc.",
	// some special configuration.
	Register: func() {
		// *** very special:: file upload *** TODO make this beautiful.
		// Special mapping, all file upload maps here
		//
		http.HandleFunc("/got/fileupload/", fileupload.FU)
	},
}
