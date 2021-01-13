package errors

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

var MaxStackDepth = 50

type Tag string

func (t Tag) String() string {
	return string(t)
}
func (t Tag) Is(tag interface{}) bool {
	return fmt.Sprint(tag) == t.String()
}

type Error struct {
	Err    string
	parent error
	stack  []uintptr
	frames []StackFrame
	tags   []Tag
}

func NewWithError(e string, parent error) *Error {
	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(2, stack[:])

	return &Error{
		Err:    e,
		stack:  stack[:length],
		parent: parent,
	}
}

func New(e string) *Error {
	err := NewWithError(e, nil)
	err.stack = err.stack[1:]
	return err
}

func Wrap(e error) *Error {
	if e == nil {
		return nil
	}

	switch e := e.(type) {
	case *Error:
		return e
	default:
		var err1 error
		if e, ok := e.(error); ok {
			err1 = Unwrap(e)
		}
		err := NewWithError(e.Error(), err1)
		err.stack = err.stack[1:]
		return err
	}
}

func Is(e error, original interface{}) bool {
	switch ori := original.(type) {
	case Tag:
		if e, ok := e.(*Error); ok {
			return e.HasTag(ori)
		}
	case error:
		found := !travelErrors(e, func(e error) bool {
			if e == ori {
				return false
			}

			if ori, ok := original.(*Error); ok {
				if ori.parent != nil {
					if Is(e, ori.parent) {
						return false
					}
				}
			}
			return true
		})

		return found
	}
	return false
}

func travelErrors(e error, fn func(e error) bool) bool {
	if !fn(e) {
		return false
	}
	if e := Unwrap(e); e != nil {
		if !travelErrors(e, fn) {
			return false
		}
	}
	return true
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

func Tags(err error) []Tag {
	tags := make(map[string]Tag)
	travelErrors(err, func(e error) bool {
		if e, ok := e.(*Error); ok {
			for _, t := range e.tags {
				tags[t.String()] = t
			}
		}
		return true
	})
	tagsArray := make([]Tag, 0)
	for _, v := range tags {
		tagsArray = append(tagsArray, v)
	}
	return tagsArray
}

type errorSack string

func (e errorSack) Error() string {
	return string(e)
}

func ErrorStack(err error) error {
	b := bytes.Buffer{}
	b.WriteString("--------------------------------\n")
	travelErrors(err, func(e error) bool {
		if er, ok := e.(*Error); ok {
			b.WriteString(er.ErrorStack())
		} else {
			b.WriteString(reflect.TypeOf(e).String() + " " + e.Error() + "\n")
		}
		b.WriteString("--------------------------------\n")
		return true
	})
	return errorSack(b.String())
}

func HasTag(err error, tag Tag) bool {
	for _, t := range Wrap(err).Tags() {
		if t.Is(tag) {
			return true
		}
	}
	return false
}

func (err *Error) AddTags(tags ...Tag) *Error {
	err.tags = append(err.tags, tags...)
	return err
}

func (err *Error) Tags() []Tag {
	return Tags(err)
}

func (err *Error) HasTag(tag Tag) bool {
	for _, t := range err.Tags() {
		if t.Is(tag) {
			return true
		}
	}
	return false
}

func (err *Error) Error() string {
	return err.Err
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
	return "type: " + err.TypeName() + " tags: " + fmt.Sprint(err.Tags()) + " error: " + err.Err + "\n" + string(err.Stack())
}

func (err *Error) StackFrames() []StackFrame {
	if err.frames == nil {
		err.frames = make([]StackFrame, len(err.stack))

		for i, pc := range err.stack {
			err.frames[i] = NewStackFrame(pc)
		}
	}

	return []StackFrame(err.frames)
}

func (err *Error) TypeName() string {
	return reflect.TypeOf(err).String()
}
