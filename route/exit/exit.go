// 这是由于route包的循环引用，所以将一些帮助方法移动到此处。
/*
Can return to these targets
  redirect::/order/list             // redirect to another URL
  template::person-list             // specific template location, not the default one.
  error::error message.             // TODO panic
  forward::/some/page               // TODO forward
  forward::InjectedPage             // forward to injected page, value is page instance.
  text::directly_render_some_text   // render text output
  json::directly_render_some_json   // render json output

*/
package exit

import (
	"fmt"
	"github.com/elivoa/got/core/exception"
)

// Exit to you.
type Exit struct {
	ExitType string
	Value    interface{} // stores values
}

func (r *Exit) IsBreakExit() bool {
	// returns true or nothing
	if r.ExitType == "bool" && r.Value == true {
		return false
	}
	if r.ExitType == "template" {
		return false
	}
	return true
}

func (r *Exit) IsReturnsTrue() bool  { return r.ExitType == "bool" && r.Value == true }
func (r *Exit) IsReturnsFalse() bool { return r.ExitType == "bool" && r.Value == false }

// ---- Exit Helper 1 --------

var (
	_trueExit  = &Exit{"bool", true}
	_falseExit = &Exit{"bool", false}
)

func TrueExit() *Exit                   { return _trueExit }
func FalseExit() *Exit                  { return _falseExit }
func Bool(b bool) *Exit                 { return &Exit{"bool", b} }
func Redirect(url interface{}) *Exit    { return &Exit{"redirect", url} }
func Forward(url interface{}) *Exit     { return &Exit{"forward", url} }
func Template(tpl interface{}) *Exit    { return &Exit{"template", tpl} }
func RenderText(text interface{}) *Exit { return &Exit{"text", text} } // use Text
func RenderJson(json interface{}) *Exit { return &Exit{"json", json} } // use Json
func Text(text interface{}) *Exit       { return &Exit{"text", text} }
func Json(json interface{}) *Exit       { return &Exit{"json", json} }
func MarshalJson(inf interface{}) *Exit { return &Exit{"marshaljson", inf} } // add 2015-03-25
func Error(err interface{}) *Exit       { return &Exit{"error", err} }

func DownloadFile(mime string, filename string, data interface{}) *Exit {
	return &Exit{"download", []interface{}{mime, filename, data}}
}

// ----------- additional functions -------------
func RedirectFirstValid(targets ...interface{}) *Exit {
	if len(targets) <= 0 {
		panic(exception.NewCoreError(nil, "Not enough parameters in exit.RedirectFirstValid()"))
	}
	for _, target := range targets {
		switch target.(type) {
		case string:
			fmt.Println(">>. is string")
			if target.(string) != "" {
				return &Exit{"redirect", target}
			}
		case interface{}:
			if target != nil {
				return &Exit{"redirect", target}
			}
		}
	}
	return nil
}
