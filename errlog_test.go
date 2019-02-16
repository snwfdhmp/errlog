package errlog

import (
	"errors"
	"testing"
)

func TestDebug(t *testing.T) {
	err := errors.New("process failed due to something")
	Debug(err)
}
