package revex

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// Reve* Configs
var (
	ImportPath string
	RunMode    string

	TRACE = log.New(ioutil.Discard, "TRACE ", log.Ldate|log.Ltime|log.Lshortfile)
	INFO  = log.New(ioutil.Discard, "INFO  ", log.Ldate|log.Ltime|log.Lshortfile)
	WARN  = log.New(ioutil.Discard, "WARN  ", log.Ldate|log.Ltime|log.Lshortfile)
	ERROR = log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)
)

type ExecutableTemplate interface {
	Execute(io.Writer, interface{}) error
}

// Execute a template and returns the result as a string.
func ExecuteTemplate(tmpl ExecutableTemplate, data interface{}) string {
	var b bytes.Buffer
	tmpl.Execute(&b, data)
	return b.String()
}

// Reads the lines of the given file.  Panics in the case of error.
func MustReadLines(filename string) []string {
	r, err := ReadLines(filename)
	if err != nil {
		panic(err)
	}
	return r
}

// Reads the lines of the given file.  Panics in the case of error.
func ReadLines(filename string) ([]string, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(bytes), "\n"), nil
}

func ContainsString(list []string, target string) bool {
	for _, el := range list {
		if el == target {
			return true
		}
	}
	return false
}
