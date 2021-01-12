// nolint
package custerr

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew1(t *testing.T) {
	type args struct {
		e interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "New From string",
			args: args{
				e: "test error",
			},
		},
		{
			name: "New From Error",
			args: args{
				e: errors.New("test error"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.e)
			assert.NotNil(t, got)
			// fmt.Println(New("Other Error", got))
			//os.Exit(0)
		})
	}
}
func TestStackFormat(t *testing.T) {

	defer func() {

		err := recover()

		if err != 'a' {

			t.Fatal(err)

		}

		e, expected := Errorf("hi"), callers()

		bs := [][]uintptr{e.stack, expected}

		if err := compareStacks(bs[0][2:], bs[1][3:]); err != nil {

			t.Errorf("Stack didn't match %v %v", bs[0], bs[1])

			t.Errorf(err.Error())

		}

		stack := string(e.Stack())

		if !strings.Contains(stack, "a: b(5)") {

			t.Errorf("Stack trace does not contain source line: 'a: b(5)'")

			t.Errorf(stack)

		}

		if !strings.Contains(stack, "custerr_test.go:") {

			t.Errorf("Stack trace does not contain file name: 'custerr_test.go:'")

			t.Errorf(stack)

		}

	}()

	a()

}

func TestSkipWorks(t *testing.T) {

	defer func() {

		err := recover()

		if err != 'a' {

			t.Fatal(err)

		}

		bs := [][]uintptr{Wrap("hi", 2).stack, callersSkip(2)}

		if err := compareStacks(bs[0], bs[1]); err != nil {

			t.Errorf("Stack didn't match")

			t.Errorf(err.Error())

		}

	}()

	a()

}

func TestIs(t *testing.T) {

	if Is(nil, io.EOF) {

		t.Errorf("nil is an error")

	}

	if !Is(io.EOF, io.EOF) {

		t.Errorf("io.EOF is not io.EOF")

	}

	if !Is(io.EOF, New(io.EOF)) {

		t.Errorf("io.EOF is not New(io.EOF)")

	}

	if !Is(New(io.EOF), New(io.EOF)) {

		t.Errorf("New(io.EOF) is not New(io.EOF)")

	}

	if Is(io.EOF, fmt.Errorf("io.EOF")) {

		t.Errorf("io.EOF is fmt.Errorf")

	}

}

func TestWrapError(t *testing.T) {

	e := func() error {

		return Wrap("hi", 1)

	}()

	if e.(*Error).Err.Error() != "hi" {

		t.Errorf("Constructor with a string failed")

	}

	if Wrap(fmt.Errorf("yo"), 0).Err.Error() != "yo" {

		t.Errorf("Constructor with an error failed")

	}

	if Wrap(e, 0) != e {

		t.Errorf("Constructor with an Error failed")

	}

	if Wrap(nil, 0) != nil {

		t.Errorf("Constructor with nil failed")

	}

}

func ExampleErrorf(x int) (int, error) {

	if x%2 == 1 {

		return 0, Errorf("can only halve even numbers, got %d", x)

	}

	return x / 2, nil

}

func ExampleWrapError() (error, error) {

	// Wrap io.EOF with the current stack-trace and return it

	return nil, Wrap(io.EOF, 0)

}

func ExampleWrapError_skip() {

	defer func() {

		if err := recover(); err != nil {

			// skip 1 frame (the deferred function) and then return the wrapped err

			err = Wrap(err, 1)

		}

	}()

}

func ExampleIs(reader io.Reader, buff []byte) {

	_, err := reader.Read(buff)

	if Is(err, io.EOF) {

		return

	}

}

func ExampleNew(UnexpectedEOF error) error {

	// calling New attaches the current stacktrace to the existing UnexpectedEOF error

	return New(UnexpectedEOF)

}

func ExampleWrap() error {

	if err := recover(); err != nil {

		return Wrap(err, 1)

	}

	return a()

}

func ExampleError_Error(err error) {

	fmt.Println(err.Error())

}

func ExampleError_ErrorStack(err error) {

	fmt.Println(err.(*Error).ErrorStack())

}

func ExampleError_Stack(err *Error) {

	fmt.Println(err.Stack())

}

func ExampleError_TypeName(err *Error) {

	fmt.Println(err.TypeName(), err.Error())

}

func ExampleError_StackFrames(err *Error) {

	for _, frame := range err.StackFrames() {

		fmt.Println(frame.File, frame.LineNumber, frame.Package, frame.Name)

	}

}

func a() error {

	b(5)

	return nil

}

func b(i int) {

	c()

}

func c() {

	panic('a')

}

// compareStacks will compare a stack created using the errors package (actual)

// with a reference stack created with the callers function (expected). The

// first entry is compared inexact since the actual and expected stacks cannot

// be created at the exact same program counter position so the first entry

// will always differ somewhat. Returns nil if the stacks are equal enough and

// an error containing a detailed error message otherwise.

func compareStacks(actual, expected []uintptr) error {

	if len(actual) != len(expected) {

		return stackCompareError("Stacks does not have equal length", actual, expected)

	}

	for i, pc := range actual {

		if i == 0 {

			firstEntryDiff := (int)(expected[i]) - (int)(pc)

			if firstEntryDiff < -27 || firstEntryDiff > 27 {

				return stackCompareError(fmt.Sprintf("First entry PC diff to large (%d)", firstEntryDiff), actual, expected)

			}

		} else if pc != expected[i] {

			return stackCompareError(fmt.Sprintf("Stacks does not match entry %d (and maybe others)", i), actual, expected)

		}

	}

	return nil

}

func stackCompareError(msg string, actual, expected []uintptr) error {

	return fmt.Errorf("%s\nActual stack trace:\n%s\nExpected stack trace:\n%s", msg, readableStackTrace(actual), readableStackTrace(expected))

}

func callers() []uintptr {

	return callersSkip(1)

}

func callersSkip(skip int) []uintptr {

	callers := make([]uintptr, MaxStackDepth)

	length := runtime.Callers(skip+2, callers[:])

	return callers[:length]

}

func readableStackTrace(callers []uintptr) string {

	var result bytes.Buffer

	frames := callersToFrames(callers)

	for _, frame := range frames {

		result.WriteString(fmt.Sprintf("%s:%d (%#x)\n\t%s\n", frame.File, frame.Line, frame.PC, frame.Function))

	}

	return result.String()

}

func callersToFrames(callers []uintptr) []runtime.Frame {

	frames := make([]runtime.Frame, 0, len(callers))

	framesPtr := runtime.CallersFrames(callers)

	for {

		frame, more := framesPtr.Next()

		frames = append(frames, frame)

		if !more {

			return frames

		}

	}

}
