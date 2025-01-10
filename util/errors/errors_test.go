package errors

import (
	"testing"

	"github.com/shoenig/test"
)

type TestError struct{}

func (t *TestError) Error() string {
	return "error"
}

func TestTypeName(t *testing.T) {
	err := &TestError{}
	test.Eq(t, "*errors.TestError", TypeName(err))

	wrappedErr := Wrap(err)
	test.Eq(t, "*errors.TestError", TypeName(wrappedErr))
}

func TestIsType(t *testing.T) {
	err := &TestError{}
	wrappedErr := Wrap(err)
	test.True(t, IsType(err, wrappedErr))
	test.True(t, IsType(wrappedErr, err))
}

func TestWrapNil(t *testing.T) {
	test.Nil(t, Wrap(nil))
}

func TestDoubleWrap(t *testing.T) {
	err := &TestError{}
	wrappedErr := Wrap(err)
	wrappedErr2 := Wrap(wrappedErr)
	test.Eq(t, wrappedErr, wrappedErr2)
}
