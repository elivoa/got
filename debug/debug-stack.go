// Package debug are copied from standard go libary "debug/stack", add some
// function to return stack trace by string.
package debug

import (
	"bytes"
	"fmt"
	"github.com/elivoa/got/core"
	"io/ioutil"
	"reflect"
	"runtime"
	"runtime/debug"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

func StackString(err error) string {
	var buf bytes.Buffer
	// write the first error.
	buf.WriteString(">> Error Stack Trace: ")

	// write inner error if any.
	e := err
	var depth = 1
	for e != nil { // infinty loop
		buf.WriteString(fmt.Sprintf("%v: %s\n", reflect.TypeOf(e), e.Error()))
		buf.WriteString("----------------------------------------------------")
		buf.WriteString("----------------------------------------------------\n")
		if coreerr, ok := e.(core.CoreError); ok {
			if coreerr.Stack() != nil {
				buf.Write(coreerr.Stack())
			} else {
				buf.WriteString("(Error Stack not available. Set ProductionMode to true to see Stacks.)\n")
			}
			e = coreerr.InnerError()
			if e != nil {
				buf.WriteString("\nInner Error is: ")
			}
		} else {
			e = nil
			if depth == 1 {
				buf.Write(debug.Stack())
			}
		}
		depth++
	}
	return buf.String()
}

func Stack() []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := 2; ; i++ { // Caller we care about is the user, 2 frames up
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		line-- // in stack trace, lines are 1-indexed but our array is 0-indexed
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.Trim(lines[n], " \t")
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Since the package path might contains dots (e.g. code.google.com/...),
	// we first remove the path prefix if there is one.
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}
