package util

import (
	"path"
	"runtime"
	"strings"
)

func GetCurrentPath(level int) string {
	_, file, _, ok := runtime.Caller(level)
	if !ok {
		panic("Can't get current path!")
	}
	basePath := path.Join(path.Dir(file), "../../..")
	// fmt.Println("|||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||")
	// fmt.Printf("basepath for level %d is %s.\n", level, basePath)
	// fmt.Println("|||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||")
	return basePath
}

// c/text ==> c__text
func EncodeContext(s string) string {
	return strings.Replace(s, "/", "__", -1)
}

func DecodeContext(s string) string {
	return strings.Replace(s, "__", "/", -1)
}
