// nolint
package custerr

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"

	errors1 "github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
)

func TestNew1(t *testing.T) {
	type args struct {
		e      interface{}
		parent error
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
		{
			name: "New From string with parent",
			args: args{
				e:      "test error",
				parent: errors.New("parent"),
			},
		},
		{
			name: "New From Error with parent",
			args: args{
				e:      errors.New("test error"),
				parent: errors.New("parent"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.e, tt.args.parent)
			assert.NotNil(t, got)
			/*
				fmt.Println(tt.name, "===================================")
				fmt.Println(New("Other Error", got))
			*/
		})
	}
}

func TestWrap(t *testing.T) {
	err := errors.New("test")
	err1 := New("test")

	if e1 := Wrap(err); e1 == err {
		t.Errorf("TestWrap error")
	}
	if e1 := Wrap(err1); e1 != err1 {
		t.Errorf("TestWrap error")
	}
	if e1 := Wrap(nil); e1 != nil {
		t.Errorf("TestWrap error")
	}
}

func TestIs(t *testing.T) {
	baseError := errors.New("base")
	otherError := errors.New("base")
	base1Error := fmt.Errorf("error1 %w", baseError)

	err1 := New(baseError)
	err2 := New(err1)
	err3 := New(err2)

	if !Is(err2, baseError) {
		t.Error("TestIs error")
	}
	if !Is(base1Error, baseError) {
		t.Error("TestIs error")
	}
	if !Is(err2, err1) {
		t.Error("TestIs error")
	}
	if !Is(err2, err3) {
		t.Error("TestIs error")
	}
	if Is(err2, otherError) {
		t.Error("TestIs error")
	}
}

func TestError_Unwrap(t *testing.T) {
	baseError := errors.New("base")
	err := New("new error", baseError)
	err1 := New("new error")
	if err.Unwrap() != baseError {
		t.Error("TestError_Unwrap error")
	}
	if err1.Unwrap() != nil {
		t.Error("TestError_Unwrap error")
	}
}

func TestError_Error(t *testing.T) {
	baseError := errors.New("base")
	err := New("new error", baseError)

	e := err.Error()

	if !strings.Contains(e, "/custerr/custerr_test.go:") {
		t.Error("TestError_Error error")
	}
	if !strings.Contains(e, "From:") {
		t.Error("TestError_Error error")
	}
	if !strings.Contains(e, "base") {
		t.Error("TestError_Error error")
	}
	if !strings.Contains(e, "err := New(\"new error\", baseError)") {
		t.Error("TestError_Error error")
	}
}

func TestError_Callers(t *testing.T) {
	c1, c2 := New("new error").Callers(), callers()

	arr1, arr2 := make([]int, 0), make([]int, 0)
	for _, pc := range c1 {
		arr1 = append(arr1, errors1.NewStackFrame(pc).LineNumber)
	}
	for _, pc := range c2 {
		arr2 = append(arr2, errors1.NewStackFrame(pc).LineNumber)
	}

	if err := compareStacks(arr1, arr2); err != nil {
		t.Errorf("TestError_Callers error %v", err)
	}
}

func compareStacks(actual, expected []int) error {
	if len(actual) != len(expected) {
		return errors.New(fmt.Sprintf("Stacks does not have equal length %v %v", actual, expected))
	}
	for i := range actual {
		if actual[i] != expected[i] {
			return errors.New(fmt.Sprintf("element %d differ %v %v", i, actual[i], expected[i]))
		}
	}
	return nil
}

func callers() []uintptr {
	return callersSkip(1)
}
func callersSkip(skip int) []uintptr {
	callers := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(skip+2, callers[:])
	return callers[:length]
}

func TestError_HasTag(t *testing.T) {
	err1 := New("error1").AddTags(Tag("database_error"))
	err2 := New("error2", err1).AddTags("service_error")
	if !err1.HasTag(Tag("database_error")) {
		t.Error("TestError_HasTag error")
	}
	if !err2.HasTag(Tag("database_error")) {
		t.Error("TestError_HasTag error")
	}
	if err1.HasTag(Tag("error_tag")) {
		t.Error("TestError_HasTag error")
	}
	if err2.HasTag(Tag("error_tag")) {
		t.Error("TestError_HasTag error")
	}
	if err1.HasTag(Tag("service_error")) {
		t.Error("TestError_HasTag error")
	}
	if !err2.HasTag(Tag("service_error")) {
		t.Error("TestError_HasTag error")
	}
}

func TestError_Tags(t *testing.T) {
	err1 := New("error1").AddTags(Tag("database_error"))
	err2 := New("error2", err1).AddTags("service_error")
	err3 := fmt.Errorf("error3 %w", err2)

	fmt.Println(Wrap(err3).Tags())
}
