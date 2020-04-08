/*
GOT Module: builtin package
*/

package builtin

import (
	"net/http"

	"github.com/elivoa/got/builtin/pages/got/fileupload"
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/utils"
)

var BuiltinModule = &core.Module{
	Name:        "github.com/elivoa/got/builtin",
	VarName:     "BuiltinModule",
	BasePath:    utils.CurrentBasePath(), // filepath.Join(workPath, "../got"),
	PackageName: "github.com/elivoa/got/builtin",
	// PackagePath: "builtin",
	// SourcePath:  "builtin",
	Description: "GOT Framework Built-in pages and components etc.",
	// some special configuration.
	Register: func() {
		// *** very special:: file upload *** TODO make this beautiful.
		// Special mapping, all file upload maps here
		//
		http.HandleFunc("/got/fileupload/", fileupload.FU)
	},
}
