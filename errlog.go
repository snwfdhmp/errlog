// Package errlog provides a simple object to enhance Go source code debugging
//
// Example result:
//
//
// 		$ go run myfailingapp.go
// 		Program starting
// 		error in main.main: something failed here
// 		line 13 of /Users/snwfdhmp/go/src/github.com/snwfdhmp/sandbox/testerr.go
// 		9: func main() {
// 		10:     fmt.Println("Program starting")
// 		11:     err := errors.New("something failed here")
// 		12:
// 		13:     errlog.Debug(err)
// 		14:
// 		15:     fmt.Println("End of the program")
// 		16: }
// 		exit status 1
//
//
// You can configure your own logger with these options :
//
//
// 		type Config struct {
// 			LinesBefore        int
// 			LinesAfter         int
// 			PrintStack         bool
// 			PrintSource        bool
// 			PrintError         bool
// 			ExitOnDebugSuccess bool
// 		}
//
//
// Example :
//
//
// 		debug := errlog.NewLogger(&errlog.Config{
// 			LinesBefore:        2,
// 			LinesAfter:         1,
// 			PrintError:         true,
// 			PrintSource:        true,
// 			PrintStack:         false,
// 			ExitOnDebugSuccess: true,
// 		})
//
// 		// ...
// 		if err != nil {
// 			debug.Debug(err)
// 			return
// 		}
//
// Outputs :
//
// 		Error in main.someBigFunction(): I'm failing for no reason
// 		line 41 of /Users/snwfdhmp/go/src/github.com/snwfdhmp/sandbox/testerr.go:41
// 		33: func someBigFunction() {
// 		...
// 		40:     if err := someNastyFunction(); err != nil {
// 		41:             debug.Debug(err)
// 		42:             return
// 		43:     }
// 		exit status 1
//
package errlog

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

var (
	debugMode = false
)

//SetDebugMode sets debug mode to On if toggle==true or Off if toggle==false. It changes log level an so displays more logs about whats happening. Useful for debugging.
func SetDebugMode(toggle bool) {
	if toggle {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	debugMode = toggle
}

var (
	fs = afero.NewOsFs() //fs is at package level because I think it needn't be scoped to loggers
)

func parseRef(refLine string) (string, int) {
	ref := strings.Split(regexpCodeReference.FindString(refLine), ":")
	if len(ref) != 2 {
		panic(fmt.Sprintf("len(ref) > 2;ref='%s';", ref))
	}

	lineNumber, err := strconv.Atoi(ref[1])
	if err != nil {
		panic(fmt.Sprintf("cannot parse line number '%s': %s", ref[1], err))
	}

	return ref[0], lineNumber
}

// PrintSourceOptions represents config for (*logger).PrintSource func
type PrintSourceOptions struct {
	FuncLine    int
	StartLine   int
	EndLine     int
	Highlighted map[int][]int //map[lineIndex][columnstart, columnEnd] of chars to highlight
}

func (l *logger) printStack(stages []string) {
	for i := len(stages) - 1; i >= 0; i-- {
		padding := ""
		if !l.config.DisableStackIndentation {
			for j := 0; j < len(stages)-1-i; j++ {
				padding += "  "
			}
		}
		l.Printf("  %s%s:%s", padding, regexpCallArgs.FindString(stages[i]), strings.Split(regexpCodeReference.FindString(stages[i]), ":")[1])
	}
}

//Debug is a shortcut for DefaultLogger.Debug.
func Debug(uErr error) bool {
	DefaultLogger.Overload(1) // Prevents from adding this func to the stack trace
	return DefaultLogger.Debug(uErr)
}

//PrintStack prints the current stack trace
func PrintStack() {
	DefaultLogger.printStack(getStackTrace(1))
}

//PrintStackMinus prints the current stack trace minus the amount of depth in parameter
func PrintStackMinus(depthToRemove int) {
	DefaultLogger.printStack(getStackTrace(1 + depthToRemove))
}
