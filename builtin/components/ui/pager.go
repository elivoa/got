/*
Got UI package -- provide basic UI components.
Time-stamp: <[pager.go] Elivoa @ Saturday, 2014-07-19 18:12:56>
*/
package ui

import (
	"bytes"
	"fmt"
	"github.com/elivoa/got/core"
	"strconv"
	"strings"
)

// decorader -- TODO make this configurable.
var (
	d_pagenumber1 = ""    // wrapper between normal page_numbers
	d_pagenumber2 = ""    //
	d_curr1       = "<"   // wrapper between current page_numbers
	d_curr2       = ">"   //
	d_spliter     = " | " // wrapper spliter between page_numbers
	d_prev_label  = "Previous"
	d_next_label  = "Next"
	d_first_label = "First" // not used
	d_last_label  = "Last"  // not used
)

var (
	i18n_index int
	// 0 en, 1 cn
	i18n = map[string][]string{
		"第":         []string{"第", ""},      // no use
		"条":         []string{"条", "items"}, // no use
		"共":         []string{"共", "Total"}, // no use
		"firstpage": []string{"首页", "First Page"},
		"lastpage":  []string{"末页", "Last Page"},
	}
)

type Pager struct {
	core.Component

	// parameters
	Total       int    // parameter Total -- total items available.
	Current     int    // parameter Current -- current item index.
	PageItems   int    // parameter PageItems -- total items per page.
	URLTemplate string // i.e. /order/list/{{CurrentPage}}/{{Total}}/{{PageItems}}
	Lang        string // cn (default) | en

	// outputs
	/*
	   This line will throw an Error: because of the type. TODO how to fix this?
	   2014/04/03 15:35:38 parser.go:497: Failed to find import for arg of type: ui.PageNumber
	*/
	PageNumbers []PageNumber // page numbers used to generate webpages. It's slice * is not needed.

	// old, to delete
	// HTML string // HTML -- generated pager html.
}

// page number items
type PageNumber struct {
	// -10000 the first page link
	// -01000 previous page link
	// -00001 the last page link
	// -00010 next page link
	PageNumber int
}

func (p *Pager) Setup() {
	p.GeneratePageNumbers()
	if p.Lang == "en" {
		i18n_index = 1
	}
}

// generate page numbers used later in rendering html.
func (p *Pager) GeneratePageNumbers() (string, error) {
	p.FixData() // fix data before.

	p.PageNumbers = make([]PageNumber, 0) // make a slice

	// var buffer bytes.Buffer
	left := p.Total
	i := 1
	for left > 0 {
		// main page nubmer
		if p.PageItems*(i-1) < p.Current && p.Current <= p.PageItems*i {
			// current page nubmer
			p.PageNumbers = append(p.PageNumbers, PageNumber{i})
		} else {
			// normal page number
			p.PageNumbers = append(p.PageNumbers, PageNumber{i})
		}
		// prepare next loop
		i += 1
		left -= p.PageItems
	}
	return "", nil
}

func (p *Pager) IsCurrentPage(pn PageNumber) bool {
	// print("---------------------------\n")
	// print(pn.PageNumber, "  --  ", p.PageItems, "  ==  ", p.Current, "\n")
	// print(">> ", pn.PageNumber*p.PageItems, " <= ", p.Current, " < \n")
	if (pn.PageNumber-1)*p.PageItems <= p.Current && p.Current < pn.PageNumber*p.PageItems {
		return true
	}
	return false
}

func (p *Pager) CreatePagerLink(pn PageNumber) string {
	url := strings.Replace(p.URLTemplate, "{{Start}}", strconv.Itoa(p.PageItems*(pn.PageNumber-1)), 1)
	url = strings.Replace(url, "{{PageItems}}", strconv.Itoa(p.PageItems), 1)
	return url
}

func (p *Pager) CreateFirstPagerLink() string {
	url := strings.Replace(p.URLTemplate, "{{Start}}", "0", 1)
	url = strings.Replace(url, "{{PageItems}}", strconv.Itoa(p.PageItems), 1)
	return url
}

func (p *Pager) CreateLastPagerLink() string {
	// print("========================================  ")
	// print(p.Total / p.PageItems)
	url := strings.Replace(p.URLTemplate, "{{Start}}", strconv.Itoa(p.Total/p.PageItems*p.PageItems), 1) // is this correct?
	url = strings.Replace(url, "{{PageItems}}", strconv.Itoa(p.PageItems), 1)
	return url
}

func (p *Pager) PageCursorMessage() string {
	if p.Lang == "en" {
		return fmt.Sprintf("%d - %d，Total %d items.", p.Current, p.Current+p.PageItems, p.Total)
	} else {
		return fmt.Sprintf("第%d - %d条，共%d条", p.Current, p.Current+p.PageItems, p.Total)
	}
}

// deprecated, try to use GeneratePageNumbers instead
// TODO move to Pager file.
func (p *Pager) GeneratePagerHtml() (string, error) {
	p.FixData()

	var buffer bytes.Buffer
	left := p.Total
	i := 1
	for left > 0 {
		// generate spliter between page_number.
		if i > 1 {
			buffer.WriteString(d_spliter)
		}

		// main page nubmer
		if p.PageItems*(i-1) < p.Current && p.Current <= p.PageItems*i {
			// current page nubmer
			buffer.WriteString("<a href='")
			buffer.WriteString("#")
			buffer.WriteString("'>")
			buffer.WriteString(d_curr1)
			buffer.WriteString(strconv.Itoa(i))
			buffer.WriteString(d_curr2)
			buffer.WriteString("</a>")
		} else {
			// normal page number
			buffer.WriteString(d_pagenumber1)
			buffer.WriteString(strconv.Itoa(i))
			buffer.WriteString(d_pagenumber2)
		}
		// prepare next loop
		i += 1
		left -= p.PageItems
	}

	return buffer.String(), nil
}

// FixData fix invalid data. return true if has errors and fixed.
func (p *Pager) FixData() bool {
	var hasError = false
	if p.Total < 0 {
		p.Total = 0
		hasError = true
	}
	if p.Current < 0 {
		p.Current = 0
		hasError = true
	}
	if p.PageItems <= 0 {
		p.PageItems = 10
		hasError = true
	}
	return hasError
}

// Check checks if all the values is right. TODO not used
func (p *Pager) Check() (bool, error) {
	return false, nil
}

func (p *Pager) Msg(key string) string {
	if msgs, ok := i18n[key]; ok {
		return msgs[i18n_index]
	} else {
		return key
	}
}
