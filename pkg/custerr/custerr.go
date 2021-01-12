package custerr

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"

	"github.com/go-errors/errors"
)

var MaxStackDepth = 50

type Error struct {
	Err    error
	parent []error
	stack  []uintptr
	frames []errors.StackFrame
}

func New(e interface{}, parent ...error) *Error {
	var err error

	switch e := e.(type) {
	case error:
		err = e
	default:
		err = fmt.Errorf("%v", e)
	}

	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(2, stack[:])
	return &Error{
		Err:    err,
		stack:  stack[:length],
		parent: parent,
	}
}

func Wrap(e interface{}, skip int) *Error {
	if e == nil {
		return nil
	}

	var err error

	switch e := e.(type) {
	case *Error:
		return e
	case error:
		err = e
	default:
		err = fmt.Errorf("%v", e)
	}

	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(2+skip, stack[:])
	return &Error{
		Err:   err,
		stack: stack[:length],
	}
}

func Is(e error, original error) bool {

	if e == original {
		return true
	}

	if e, ok := e.(*Error); ok {
		return Is(e.Err, original)
	}

	if original, ok := original.(*Error); ok {
		return Is(e, original.Err)
	}

	return false
}

func Errorf(format string, a ...interface{}) *Error {
	return Wrap(fmt.Errorf(format, a...), 2)
}

func (err *Error) Error() string {

	msg := err.Err.Error()

	msg = fmt.Sprintf("%s\n%v", msg, string(err.Stack()))

	if len(err.parent) > 0 {
		for _, e := range err.parent {
			msg = fmt.Sprintf("%sFrom:\n%v", msg, e)
			msg += "\n"
		}
	}
	msg += "\n"

	return msg
}

func (err *Error) Stack() []byte {
	buf := bytes.Buffer{}

	for _, frame := range err.StackFrames() {
		/*
			if strings.Contains(frame.String(), "/runtime/") ||
				strings.Contains(frame.String(), "/testing/") {
				continue
			}
		*/
		buf.WriteString(frame.String())
	}

	return buf.Bytes()
}

func (err *Error) Callers() []uintptr {
	return err.stack
}

func (err *Error) ErrorStack() string {
	return err.TypeName() + " " + err.Error() + "\n" + string(err.Stack())
}

func (err *Error) StackFrames() []errors.StackFrame {
	if err.frames == nil {
		err.frames = make([]errors.StackFrame, len(err.stack))

		for i, pc := range err.stack {
			err.frames[i] = errors.NewStackFrame(pc)
		}
	}

	return err.frames
}

func (err *Error) TypeName() string {
	return reflect.TypeOf(err.Err).String()
}
