package main

import (
	"errors"
	"fmt"

	"github.com/snwfdhmp/errlog"
)

func main() {
	fmt.Println("Example start")

	wrappingFunc()

	fmt.Println("Example end")
}

func wrappingFunc() {
	someBigFunction()
}

func someBigFunction() {
	someDumbFunction()

	someSmallFunction()

	someDumbFunction()

	if err := someNastyFunction(); errlog.Debug(err) {
		return
	}

	someSmallFunction()

	someDumbFunction()
}

func someSmallFunction() {
	_ = fmt.Sprintf("I do things !")
}

func someNastyFunction() error {
	return errors.New("I'm failing for some reason")
}

func someDumbFunction() bool {
	return false
}
