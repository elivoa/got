package logs

import (
	"fmt"
	"log"
	"os"
)

const (
	Error = 1 << iota
	Warning
	Info
	Debug
	Trace
)

// TODO: override methods to support middle prefix-

type Logger struct {
	*log.Logger
	loglevel int
}

func (l *Logger) Error() bool   { return l.loglevel >= Error }
func (l *Logger) Warning() bool { return l.loglevel >= Warning }
func (l *Logger) Info() bool    { return l.loglevel >= Info }
func (l *Logger) Debug() bool   { return l.loglevel >= Debug }
func (l *Logger) Trace() bool   { return l.loglevel >= Trace }

// ------

var logmap map[string]*Logger

func init() {
	logmap = make(map[string]*Logger, 0)
}

var flag = log.Lmicroseconds | log.Lshortfile

func InitLogger(name string, prefix string, logLevel int) *Logger {
	if logger, ok := logmap[name]; ok {
		panic(fmt.Sprintf("Logger '%s' already exists.", name))
	} else {
		logger = &Logger{
			Logger:   log.New(os.Stdout, prefix, flag),
			loglevel: logLevel,
		}
		logmap[name] = logger
		return logger
	}
}

// Logger get logger by name, create one if not exist
func Get(name string) *Logger {
	if logger, ok := logmap[name]; ok {
		return logger
	} else {
		logger = &Logger{
			Logger:   log.New(os.Stdout, "", flag),
			loglevel: Trace,
		}
		logmap[name] = logger
		return logger
	}
}

func init() {
	InitLogger("IOC:Inject", "", Trace)
	InitLogger("GOT:PageFlow", "", Trace)
	InitLogger("GOT:EventCall", "", Trace)
	InitLogger("URL Lookup", "", Trace)
	InitLogger("ComponentFLow", "", Trace)
	InitLogger("Route", "", Trace)
	InitLogger("Return", "", Trace)

}
