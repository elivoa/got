package utils

import (
	"database/sql"
	"fmt"
	"go/build"
	"log"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func CurrentPackagePath() string {
	// get base path
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		panic("Can't get current path!")
	}
	basePath := path.Join(path.Dir(file))
	return PackagePath(basePath)
	// for _, gopath := range filepath.SplitList(build.Default.GOPATH) {
	// 	srcPath := filepath.Join(gopath, "src")
	// 	if strings.HasPrefix(basePath, srcPath) {
	// 		return filepath.ToSlash(basePath[len(srcPath)+1:])
	// 	}
	// }

	// srcPath := filepath.Join(build.Default.GOROOT, "src", "pkg")
	// if strings.HasPrefix(basePath, srcPath) {
	// 	log.Fatalf("Code path should be in GOPATH, but is in GOROOT: %v", basePath)
	// 	return filepath.ToSlash(basePath[len(srcPath)+1:])
	// }

	// log.Fatalln("Unexpected! Code path is not in GOPATH:", basePath)
	// return ""
}

func PackagePath(basePath string) string {
	for _, gopath := range filepath.SplitList(build.Default.GOPATH) {
		srcPath := filepath.Join(gopath, "src")
		if strings.HasPrefix(basePath, srcPath) {
			return filepath.ToSlash(basePath[len(srcPath)+1:])
		}
	}

	srcPath := filepath.Join(build.Default.GOROOT, "src", "pkg")
	if strings.HasPrefix(basePath, srcPath) {
		log.Fatalf("Code path should be in GOPATH, but is in GOROOT: %v", basePath)
		return filepath.ToSlash(basePath[len(srcPath)+1:])
	}

	log.Fatalln("Unexpected! Code path is not in GOPATH:", basePath)
	return ""
}

func CurrentBasePath() string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		panic("Can't get current path!")
	}
	currentPath := path.Join(path.Dir(file))
	fmt.Println("currentpath is : ", currentPath)
	return BasePath(currentPath)
	// for _, gopath := range filepath.SplitList(build.Default.GOPATH) {
	// 	srcPath := filepath.Join(gopath, "src")
	// 	if strings.HasPrefix(currentPath, srcPath) {
	// 		return srcPath
	// 		// return filepath.ToSlash(currentPath[len(srcPath)+1:])
	// 	}
	// }

	// srcPath := filepath.Join(build.Default.GOROOT, "src", "pkg")
	// if strings.HasPrefix(currentPath, srcPath) {
	// 	log.Fatalf("Code path should be in GOPATH, but is in GOROOT: %v", currentPath)
	// 	return srcPath
	// 	// return filepath.ToSlash(currentPath[len(srcPath)+1:])
	// }

	// log.Fatalln("Unexpected! Code path is not in GOPATH:", currentPath)
	// return ""
}

func BasePath(currentPath string) string {
	for _, gopath := range filepath.SplitList(build.Default.GOPATH) {
		srcPath := filepath.Join(gopath, "src")
		if strings.HasPrefix(currentPath, srcPath) {
			return srcPath
			// return filepath.ToSlash(currentPath[len(srcPath)+1:])
		}
	}

	srcPath := filepath.Join(build.Default.GOROOT, "src", "pkg")
	if strings.HasPrefix(currentPath, srcPath) {
		log.Fatalf("Code path should be in GOPATH, but is in GOROOT: %v", currentPath)
		return srcPath
		// return filepath.ToSlash(currentPath[len(srcPath)+1:])
	}

	log.Fatalln("Unexpected! Code path is not in GOPATH:", currentPath)
	return ""
}

func IsCapitalized(s string) bool {
	if len(s) > 0 {
		firstLetter := s[0]
		if 65 <= firstLetter && firstLetter <= 90 {
			return true
		}
	}
	return false
}

func Capitalize(s string) string {
	if len(s) > 0 {
		firstLetter := s[0]
		return strings.ToUpper(strconv.Itoa(int(firstLetter))) + s[1:]
	}
	return ""
}

var valid_earliest_time time.Time

func ValidTime(t time.Time) bool {
	fmt.Println(t)
	return t.After(valid_earliest_time)
}

// convert utils
func ToNullInt64Array(int64Array []int64) []sql.NullInt64 {
	var nullarray = []sql.NullInt64{}
	for _, v := range int64Array {
		nullarray = append(nullarray, sql.NullInt64{Int64: v, Valid: true})
	}
	return nullarray
}