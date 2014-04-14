// 这是由于route包的循环引用，所以将一些帮助方法移动到此处。
package exit

import ()

// Exit to you.
type Exit struct {
	/*
	 * redirect::/order/list             // redirect to another URL
	 * template::person-list             // specific template location, not the default one.
	 * error::error message.             // TODO panic
	 * forward::/some/page               // TODO forward
	 * text::directly_render_some_text   // render text output
	 * json::directly_render_some_json   // render json output
	 */
	ExitType string
	Value    interface{}
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

func TrueExit() *Exit                   { return &Exit{"bool", true} }
func FalseExit() *Exit                  { return &Exit{"bool", false} }
func Bool(b bool) *Exit                 { return &Exit{"bool", b} }
func Redirect(url interface{}) *Exit    { return &Exit{"redirect", url} }
func Forward(url interface{}) *Exit     { return &Exit{"forward", url} }
func Template(tpl interface{}) *Exit    { return &Exit{"template", tpl} }
func RenderText(text interface{}) *Exit { return &Exit{"text", text} }
func RenderJson(json interface{}) *Exit { return &Exit{"json", json} }
func Error(err interface{}) *Exit       { return &Exit{"error", err} }
