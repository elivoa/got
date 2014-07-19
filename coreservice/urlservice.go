// package locate under route and lcc, above all the others.
package coreservice

import (
	"bytes"
	"fmt"
	"github.com/elivoa/got/config"
	"github.com/elivoa/got/register"
	"github.com/elivoa/got/utils"
	"github.com/elivoa/got/core"
	"strings"
)

var Url = new(UrlService)

type UrlService struct {
}

// Generate url by Page value, including:
// 1. base url of page.
// 2. TODO contexts.
// 3. URL parameters
func (s *UrlService) GenerateUrlByPage(page core.Pager) string {
	// 1. base url of page.
	t := utils.GetRootType(page)
	seg, ok := register.PageTypeMap[t]
	if !ok {
		panic(fmt.Sprintf("Can't find page config(seg) for %s when generating url.", t))
	}

	// gather items
	reversedItems := []string{}
	for cursor := seg; cursor != nil; {
		reversedItems = append(reversedItems, cursor.Name)
		cursor = cursor.Parent
	}

	// TODO: try to remove XXXIndex
	// len:= len(reversedItems)
	// if len> 0 {
	// 	last := reversedItems[len(len)-1]
	// 	if index := strings.Index(last, "Index"); index > 0 {

	// 		strings.ToLower(last[0:index]) == reversedItems[len(reversedItems)]
	// 		fmt.Println("last is: ", last[0:index])
	// 		if
	// 	}
	// }

	var url bytes.Buffer
	for i := len(reversedItems) - 2; i >= 0; i-- { // remove the root /
		url.WriteRune('/')
		url.WriteString(reversedItems[i])
	}
	fmt.Println("done!", reversedItems)
	fmt.Println("generated url is !", url.String())
	// 2. TODO contexts.
	// 3. URL parameters
	return url.String()
}

func (s *UrlService) AppendVerificationCode(url string, code string) string {
	var appendSign = "?"
	if strings.Index(url, "?") > 0 {
		appendSign = "&"
	}
	return strings.Join([]string{url, appendSign, config.VERIFICATION_CODE_KEY, "=", code}, "")
}
