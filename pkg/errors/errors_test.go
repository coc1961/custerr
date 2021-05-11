// nolint
package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		e      string
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
			name: "New From string with parent",
			args: args{
				e:      "test error",
				parent: errors.New("parent"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewWithError(tt.args.e, tt.args.parent)
			if got == nil {
				t.Error("Error TestNew")
			}
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

	err1 := NewWithError("baseError", baseError).AddTags(Tag("test_tag"))
	err2 := NewWithError("err1", err1)
	err3 := NewWithError("err2", err2)

	if !Is(base1Error, baseError) {
		t.Error("TestIs error")
	}

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
	if !Is(err3, Tag("test_tag")) {
		t.Error("TestIs error")
	}

	err4 := NewWithError("error4", err1)
	if !Is(err4, Tag("test_tag")) {
		t.Error("TestIs error")
	}

	if !Is(err4, err1) {
		t.Error("TestIs error")
	}

	if Is(base1Error, nil) {
		t.Error("TestIs error")
	}
	if Is(nil, nil) {
		t.Error("TestIs error")
	}

}

func TestError_Unwrap(t *testing.T) {
	baseError := errors.New("base")
	err := NewWithError("new error", baseError)
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
	err, c := NewWithError("new error", baseError).AddTags("test_tag").AddTags("test_tag_1"), callers()
	err1 := Wrap(fmt.Errorf("other test %w", err))

	e := ErrorStack(err1).Error()
	frame := NewStackFrame(c[0])

	fmt.Println(e)

	if !strings.Contains(e, frame.File) ||
		!strings.Contains(e, fmt.Sprintf("%d", frame.LineNumber)) {
		t.Error("TestError_Error error")
	}
	if !strings.Contains(e, "base") {
		t.Error("TestError_Error error")
	}
	if !strings.Contains(e, `NewWithError(\"new error\", baseError)`) {
		t.Error("TestError_Error error")
	}
}

func TestError_Callers(t *testing.T) {
	c1, c2 := New("new error").Callers(), callers()

	arr1, arr2 := make([]int, 0), make([]int, 0)
	for _, pc := range c1 {
		arr1 = append(arr1, NewStackFrame(pc).LineNumber)
	}
	for _, pc := range c2 {
		arr2 = append(arr2, NewStackFrame(pc).LineNumber)
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
	err2 := NewWithError("error2", err1).AddTags("service_error")
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

	err3 := New("error1").AddTags(Tag("database_error"))
	err4 := NewWithError("err3", err3)
	if !err4.HasTag(Tag("database_error")) {
		t.Error("TestError_HasTag error")
	}

}

func TestError_Tags(t *testing.T) {
	err1 := New("error1").AddTags(Tag("database_error"))
	err2 := NewWithError("error2", err1).AddTags("service_error")
	err3 := fmt.Errorf("error3 %w", err2)

	if len(Wrap(err3).Tags()) != 2 {
		t.Error("TestError_Tags")
	}
}

func TestHasTag(t *testing.T) {
	err1 := New("error1").AddTags(Tag("database_error"))
	err2 := NewWithError("error2", err1).AddTags("service_error")
	if !HasTag(err1, Tag("database_error")) {
		t.Error("TestError_HasTag error")
	}
	if !HasTag(err2, Tag("database_error")) {
		t.Error("TestError_HasTag error")
	}
	if HasTag(err1, Tag("error_tag")) {
		t.Error("TestError_HasTag error")
	}
	if HasTag(err2, Tag("error_tag")) {
		t.Error("TestError_HasTag error")
	}
	if HasTag(err1, Tag("service_error")) {
		t.Error("TestError_HasTag error")
	}
	if !HasTag(err2, Tag("service_error")) {
		t.Error("TestError_HasTag error")
	}
}

func TestTags(t *testing.T) {
	err1 := New("error1").AddTags(Tag("database_error"))
	err2 := NewWithError("error2", err1).AddTags("service_error")
	err3 := fmt.Errorf("error3 %w", err2)

	if len(Tags(err3)) != 2 {
		t.Error("TestError_Tags")
	}
}

func Test_goThroughErrors(t *testing.T) {
	err1 := New("error1")
	err2 := NewWithError("error2", err1)
	err3 := fmt.Errorf("error3 %w", err2)

	mp := make([]error, 0)
	goThroughErrors(err3, func(e error) bool {
		mp = append(mp, e)
		return true
	})
	if len(mp) != 3 {
		t.Error("Test_travelErrors error")
	}

	mp = make([]error, 0)
	goThroughErrors(nil, func(e error) bool {
		mp = append(mp, e)
		return true
	})
	if len(mp) != 0 {
		t.Error("Test_travelErrors error")
	}
}

func TestTag(t *testing.T) {
	tag := Tag("test")

	switch tag {
	case Tag("test"):
		// Ok
	case Tag("test1"):
		t.Error("TestTag")
	default:
		t.Error("TestTag")
	}
}
