package ferr

import (
	"fmt"
	"testing"
)

func Test_assembleStack(t *testing.T) {
	err := returnErr3()
	if err != nil {
		fmt.Println(Summarize(err))
	}
}

func returnErr3() error {
	err := returnErr2()
	if err != nil {
		return Wrap(err)
	}

	return nil
}


func returnErr2() error {
	err := returnErr1()
	if err != nil {
		return Wrapf(err, "Failed to call returnErr1")
	}

	return nil
}

func returnErr1() error {
	err := returnErr()
	if err != nil {
		return Wrap(err)
	}

	return nil
}

func returnErr() error {
	return Wrap(fmt.Errorf("this is an error"))
}