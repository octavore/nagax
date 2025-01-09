package httperror

import (
	"fmt"
	"testing"

	"github.com/shoenig/test"
)

type CodeTestError struct{}

func (e *CodeTestError) Error() string {
	return "200"
}

func (e *CodeTestError) GetCode() int {
	return 200
}

func TestCodeFromErr(t *testing.T) {
	testCases := []struct {
		desc string
		err  error
		code int
	}{
		{desc: "basic", err: &CodeTestError{}, code: 200},
		{desc: "wrapped", err: fmt.Errorf("%w", &CodeTestError{}), code: 200},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.err), func(t *testing.T) {
			code, ok := CodeFromErr(tc.err)
			test.True(t, ok)
			test.Eq(t, code, 200)
		})
	}
}
