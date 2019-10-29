package errors

import (
	"fmt"
	"github.com/pkg/errors"
	"testing"
)

func TestErrors(t *testing.T) {
	err := fmt.Errorf("test err1")
	fmt.Printf("err1: %s\n", err)
	fmt.Printf("err1: %+v\n", err)

	err2 := errors.Wrap(err, "wrap2") //not need :
	fmt.Printf("err2: %s\n", err2)
	fmt.Printf("err2: %+v\n", err2)

	err3 := errors.New("test new err1")
	fmt.Printf("err3: %s\n", err3)
	fmt.Printf("err3: %+v\n", err3)

	err4 := errors.WithStack(err)
	fmt.Printf("err4: %s\n", err4)
	fmt.Printf("err4: %+v\n", err4)

	err5 := errors.WithMessage(err, "msg test") //not need :, no stack
	fmt.Printf("err5: %s\n", err5)
	fmt.Printf("err5: %+v\n", err5)

	errors.Cause()
}
