package errors

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"
)

type StackFrame struct {
	File           string
	LineNumber     int
	Name           string
	Package        string
	ProgramCounter uintptr
}

func NewStackFrame(pc uintptr) (frame StackFrame) {

	frame = StackFrame{ProgramCounter: pc}
	if frame.Func() == nil {
		return
	}
	frame.Package, frame.Name = packageAndName(frame.Func())
	frame.File, frame.LineNumber = frame.Func().FileLine(pc - 1)
	return

}

func (frame *StackFrame) Func() *runtime.Func {
	if frame.ProgramCounter == 0 {
		return nil
	}
	return runtime.FuncForPC(frame.ProgramCounter)
}

func (frame *StackFrame) String() string {
	str := fmt.Sprintf("\t%s:%d (0x%x)\n", frame.File, frame.LineNumber, frame.ProgramCounter)

	source, err := frame.SourceLine()
	if err != nil {
		return str
	}

	return str + fmt.Sprintf("\t\tCode := %s -> %s\n", frame.Name, source)
}

func (frame *StackFrame) SourceLine() (string, error) {
	data, err := ioutil.ReadFile(frame.File)

	if err != nil {
		return "", New(err)
	}

	lines := bytes.Split(data, []byte{'\n'})
	if frame.LineNumber <= 0 || frame.LineNumber >= len(lines) {
		return "???", nil
	}
	return string(bytes.Trim(lines[frame.LineNumber-1], " \t")), nil
}

func packageAndName(fn *runtime.Func) (string, string) {
	name := fn.Name()
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
