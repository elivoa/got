package service

import (
	"bytes"
	"fmt"
	"got/core"
	"got/core/lifecircle"
)

type LinkService struct {
	core.Service

	Life *lifecircle.Life `inject:"life"` // TODO: inject this.
}

// GeneratePageUrl generates page's url, if it's a component, generate the container page's url.
func (s *LinkService) GeneratePageUrl() string {
	return s.Life.GeneratePageUrl()
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
