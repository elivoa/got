/*
GOT framework builtin services.

Time-stamp: <[LinkService.go] Elivoa @ Wednesday, 2015-04-22 16:31:53>
*/
package services

import (
	"bytes"
	"fmt"
	"github.com/elivoa/got/core"
	"github.com/elivoa/got/core/lifecircle"
	"github.com/elivoa/got/coreservice/coercion"
	"github.com/elivoa/got/utils"
	"time"
)

var Link = new(LinkService)

type LinkService struct {
	core.Service

	Life *lifecircle.Life `inject:"life"` // TODO: inject this.
}

// GeneratePageUrl generates page's url, if it's a component, generate the container page's url.
func (s *LinkService) GeneratePageUrl(page string) string {
	// TODO: extract PageSourceService to get page from page name.
	// here we directly get this page from page name.(index, order/edit)

	fmt.Println("TODO finish this function's design!")
	return "/" + page
	// return s.Life.GeneratePageUrl()
}

func (s *LinkService) GeneratePageUrlWithContext(page string, contexts ...interface{}) string {
	fmt.Println("TODO finish this function's design!")

	var buffer bytes.Buffer
	buffer.WriteString("/")
	buffer.WriteString(page)

	if nil != contexts {
		for _, context := range contexts {
			buffer.WriteString("/")
			buffer.WriteString(fmt.Sprint(context)) // TODO: not support none string contexts.
		}
	}
	return buffer.String()
}

func (s *LinkService) GeneratePageUrlWithContextAndQueryParameters(page string,
	parameters map[string]interface{}, contexts ...interface{}) string {

	fmt.Println("TODO finish this function's design!")

	url := s.GeneratePageUrlWithContext(page, contexts...)

	if nil != parameters && len(parameters) > 0 {
		var buffer bytes.Buffer
		var index = 0
		for key, value := range parameters {
			var (
				strValue string
				usethis  bool = true
			)
			switch value.(type) {
			case time.Time:
				t := value.(time.Time)
				if utils.IsValidTime(t) {
					strValue = coercion.DateTime(t)
				} else {
					strValue = ""
					usethis = false
				}
			default:
				if nil == value {
					usethis = false
				} else {
					strValue = fmt.Sprint(value)
				}
			}

			if usethis {
				if index > 0 {
					buffer.WriteRune('&')
				}
				buffer.WriteString(key)
				buffer.WriteRune('=')
				buffer.WriteString(strValue)
				index += 1
			}
		}
		url = url + "?" + buffer.String()
	}
	return url
}

// for common use.
func (s *LinkService) GenerateEventUrl(eventName string, contexts ...interface{}) string {
	return s.GenerateEventUrl(eventName, 0, contexts)
}

// Used by buildin components, can ignore the last N system compnents.
func (s *LinkService) GenerateEventUrlIgnoreComponent(eventName string, ignoreLastNComponents int,
	contexts ...interface{}) string {

	// for example /got/status.templatestatus:templatedetail/<template key>
	var pieces = []string{}
	var current = s.Life
	// if current is component or mixin, get id.
	for !current.Is(core.PAGE) {
		pieces = append(pieces, current.GetProton().ClientId())
		current = current.GetContainer()
	}

	// build url: /got/status.templatestatus
	var buffer bytes.Buffer
	buffer.WriteString(s.Life.GeneratePageUrl())
	for i := len(pieces) - 1; i >= ignoreLastNComponents; i-- {
		buffer.WriteString(".")
		buffer.WriteString(pieces[i])
	}

	// event call part: /got/status.templatestatus:templatedetail
	buffer.WriteString(":")
	buffer.WriteString(eventName)

	// build context part: /got/status.templatestatus:templatedetail/<template key>
	if nil != contexts {
		for _, context := range contexts {
			buffer.WriteString("/")
			buffer.WriteString(fmt.Sprint(context)) // TODO: not support none string contexts.
		}
	}

	return buffer.String()
}
