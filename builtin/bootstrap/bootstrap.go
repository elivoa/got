package bootstrap

import (
	"github.com/elivoa/got/config"
	"github.com/gorilla/context"
	"html/template"
	"net/http"
)

func NewPageBootstraps() *PageBootstraps {
	return &PageBootstraps{
		HTMLs: map[string]*PageBootstrapHtmlItem{},
	}
}

type PageBootstraps struct {
	// 存储页面列表
	HTMLs map[string]*PageBootstrapHtmlItem
}

type PageBootstrapHtmlItem struct {
	Key   string
	HTML  template.HTML
	Index int
}

func GetBootstraps(req *http.Request) *PageBootstraps {
	if bts, ok := context.GetOk(req, config.PAGE_FINAL_BOOTSTRAP_CONTEXT_KEY); ok {
		return bts.(*PageBootstraps)
	}
	return nil
}

func AddHtml(req *http.Request, uniquekey string, html template.HTML) {
	var bootstraps *PageBootstraps
	bts := context.Get(req, config.PAGE_FINAL_BOOTSTRAP_CONTEXT_KEY)
	if bts == nil {
		bootstraps = NewPageBootstraps()
		context.Set(req, config.PAGE_FINAL_BOOTSTRAP_CONTEXT_KEY, bootstraps)
	} else {
		bootstraps = bts.(*PageBootstraps)
	}

	bootstraps.HTMLs[uniquekey] = &PageBootstrapHtmlItem{
		Key:   uniquekey,
		HTML:  html,
		Index: 0,
	}
}

func Has(req *http.Request, uniquekey string) bool {
	_, ok := context.GetOk(req, config.PAGE_FINAL_BOOTSTRAP_CONTEXT_KEY)
	return ok
}
