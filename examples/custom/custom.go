package main

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/snwfdhmp/errlog"
)

var (
	debug = errlog.NewLogger(&errlog.Config{
		PrintFunc:          logrus.Errorf,
		LinesBefore:        6,
		LinesAfter:         3,
		PrintError:         true,
		PrintSource:        true,
		PrintStack:         false,
		ExitOnDebugSuccess: true,
	})
)

func main() {
	logrus.Print("Start of the program")

	wrapingFunc()

	logrus.Print("End of the program")
}

func wrapingFunc() {
	someBigFunction()
}

func someBigFunction() {
	someDumbFunction()

	someSmallFunction()

	someDumbFunction()

	if err := someNastyFunction(); debug.Debug(err) {
		return
	}

	someSmallFunction()

	someDumbFunction()
}

func someSmallFunction() {
	logrus.Print("I do things !")
}

func someNastyFunction() error {
	return errors.New("I'm failing for no reason")
}

func someDumbFunction() bool {
	return false
}
