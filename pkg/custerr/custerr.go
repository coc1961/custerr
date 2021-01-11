package custerr

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"
)

type Error struct {
	e         interface{}
	prevError error
	stack     []string
}

func (e *Error) Error() string {
	b := bytes.Buffer{}
	_, _ = b.WriteString(fmt.Sprintf("%v\n", e.e))
	for _, s := range e.stack {
		_, _ = b.WriteString(fmt.Sprintf("\t%v\n", s))
	}
	if e.prevError != nil {
		b.WriteString(fmt.Sprintf("\n%v\n", e.prevError.Error()))
	}
	return b.String()
}

func NewWithParentError(e interface{}, prevError error) *Error {
	err := internalNew(e, 3)
	err.prevError = prevError
	return err
}

func New(e interface{}) *Error {
	return internalNew(e, 3)
	/*
		pc := make([]uintptr, 100)
		max := runtime.Callers(2, pc)
		pc1 := pc[0:max]

		stack := make([]string, 0)
		frames := runtime.CallersFrames(pc1)
		for {
			frame, more := frames.Next()
			if !more {
				break
			}
			if strings.Contains(frame.File, "runtime/") {
				continue
			}
			_, name := packageAndName(frame.Function)
			stack = append(stack, fmt.Sprintf("%v(%v) %v() %v", frame.File, frame.Line, name, sourceLine(frame)))
		}
		return &Error{
			e:     e,
			stack: stack,
		}
	*/
}

func internalNew(e interface{}, skip int) *Error {
	pc := make([]uintptr, 100)
	max := runtime.Callers(skip, pc)
	pc1 := pc[0:max]

	stack := make([]string, 0)
	frames := runtime.CallersFrames(pc1)
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		if strings.Contains(frame.File, "runtime/") {
			continue
		}
		_, name := packageAndName(frame.Function)
		stack = append(stack, fmt.Sprintf("%v(%v) %v() %v", frame.File, frame.Line, name, sourceLine(frame)))
	}
	return &Error{
		e:     e,
		stack: stack,
	}
}

func sourceLine(frame runtime.Frame) string {
	data, err := ioutil.ReadFile(frame.File)
	if err != nil {
		return ""
	}
	lines := bytes.Split(data, []byte{'\n'})
	if frame.Line <= 0 || frame.Line >= len(lines) {
		return ""
	}
	return string(bytes.Trim(lines[frame.Line-1], " \t"))
}

func packageAndName(name string) (string, string) {
	pkg := ""
	if lastslash := strings.LastIndex(name, "/"); lastslash >= 0 {
		pkg += name[:lastslash] + "/"
		name = name[lastslash+1:]
	}
	if period := strings.Index(name, "."); period >= 0 {
		pkg += name[:period]
		name = name[period+1:]
	}
	name = strings.Replace(name, "·", ".", -1)
	return pkg, name
}
