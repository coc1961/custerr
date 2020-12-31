package custerr

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
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
			fmt.Println(got)
		})
	}
}
