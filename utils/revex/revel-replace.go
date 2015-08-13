package revex

import (
	"io/ioutil"
	"log"
	"os"
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
