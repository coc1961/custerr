package custerr

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/go-errors/errors"
)

var MaxStackDepth = 50

type Error struct {
	Err    error
	parent error
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
	var parent1 error
	if len(parent) > 0 {
		parent1 = parent[0]
	}
	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(2, stack[:])
	return &Error{
		Err:    err,
		stack:  stack[:length],
		parent: parent1,
	}
}

func Wrap(e interface{}) *Error {
	if e == nil {
		return nil
	}

	switch e := e.(type) {
	case *Error:
		return e
	default:
		err := New(e)
		err.stack = err.stack[1:]
		return err
	}
}

func Is(e error, original error) bool {
	for {
		if e == original {
			return true
		}
		if e, ok := e.(*Error); ok {
			return Is(e.Err, original)
		}

		if original, ok := original.(*Error); ok {
			return Is(e, original.Err)
		}
		if e = Unwrap(e); e == nil {
			return false
		}
	}
}

func Unwrap(err error) error {
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return nil
	}
	return u.Unwrap()
}

func (err *Error) Error() string {
	b := bytes.Buffer{}
	b.WriteString(err.ErrorStack())
	if err.parent != nil {
		b.WriteString(fmt.Sprintf("From:\n%v", err.parent))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String()
}

func (err *Error) Unwrap() error {
	if err.parent != nil {
		return err.parent
	}
	return nil
}

func (err *Error) Stack() []byte {
	buf := bytes.Buffer{}

	for _, frame := range err.StackFrames() {
		if strings.Contains(frame.String(), "/runtime/") ||
			strings.Contains(frame.String(), "/testing/") {
			continue
		}
		buf.WriteString(frame.String())
	}

	return buf.Bytes()
}

func (err *Error) Callers() []uintptr {
	return err.stack
}

func (err *Error) ErrorStack() string {
	return err.TypeName() + " " + err.Err.Error() + "\n" + string(err.Stack())
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
