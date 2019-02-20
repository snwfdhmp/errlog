package main

import (
	"errors"
	"fmt"

	"github.com/snwfdhmp/errlog"
)

func main() {
	fmt.Println("Start of the program")

	wrapingFunc()

	fmt.Println("End of the program")
}

func wrapingFunc() {
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
	fmt.Println("I do things !")
}

func someNastyFunction() error {
	return errors.New("I'm failing for no reason")
}

func someDumbFunction() bool {
	return false
}
