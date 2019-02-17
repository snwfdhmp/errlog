// package errlog aims to simplify Golang program debugging.
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
// // ...
// if err != nil {
// 	debug.Debug(err)
// 	return
// }
// ```
//
// Outputs :
//
// ```
// Error in main.someBigFunction(): I'm failing for no reason
// line 41 of /Users/snwfdhmp/go/src/github.com/snwfdhmp/sandbox/testerr.go:41
// 33: func someBigFunction() {
// ...
// 40:     if err := someNastyFunction(); err != nil {
// 41:             debug.Debug(err)
// 42:             return
// 43:     }
// exit status 1
// ```
package errlog

import (
	"fmt"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/afero"
)

type Logger interface {
	Debug(err error)
}

type logger struct {
	Config             *Config
	stackDepthOverload int
}

func NewLogger(cfg *Config) *logger {
	return &logger{
		Config:             cfg,
		stackDepthOverload: 0,
	}
}

type Config struct {
	LinesBefore        int
	LinesAfter         int
	PrintStack         bool
	PrintSource        bool
	PrintError         bool
	ExitOnDebugSuccess bool
}

var (
	regexpParseStack    = regexp.MustCompile(`((((([a-zA-Z]+)[/])*)(([(*a-zA-Z0-9)])*(\.))+[a-zA-Z0-9]+[(](.*)[)])[\s]+[/a-zA-Z0-9\.]+[:][0-9]+)`)
	regexpCodeReference = regexp.MustCompile(`[/a-zA-Z0-9\.]+[:][0-9]+`)
	regexpCallArgs      = regexp.MustCompile(`((([a-zA-Z]+)[/])*)(([(*a-zA-Z0-9)])*(\.))+[a-zA-Z0-9]+[(](.*)[)]`)
	regexpCallingObject = regexp.MustCompile(`((([a-zA-Z]+)[/])*)(([(*a-zA-Z0-9)])*(\.))+[a-zA-Z0-9]+`)
	regexpFuncLine      = regexp.MustCompile(`^func[\s][a-zA-Z0-9]+[(](.*)[)][\s]*{`)

	DefaultLogger = logger{
		Config: &Config{
			LinesBefore:        4,
			LinesAfter:         2,
			PrintStack:         true,
			PrintSource:        true,
			PrintError:         true,
			ExitOnDebugSuccess: false,
		},
	}

	linesBefore = 4
	linesAfter  = 2

	fs = afero.NewOsFs()
)

// Debug prints useful informations for debug such as surrounding code, stack trace, ...
func (l *logger) Debug(uErr error) {
	if uErr == nil {
		return
	}
	stages := getStackTrace(1 + l.stackDepthOverload)
	l.stackDepthOverload = 0
	if l.Config.PrintError {
		fmt.Printf("\nError in %s: %s\n", regexpCallArgs.FindString(stages[0]), color.YellowString(uErr.Error()))
	}

	if l.Config.PrintSource {
		filepath, lineNumber := parseRef(stages[0])
		l.PrintLines(filepath, lineNumber)
	}

	if l.Config.PrintStack {
		fmt.Println("Stack trace:")
		printStack(stages)
	}

	if l.Config.ExitOnDebugSuccess {
		os.Exit(1)
	}
}

func (l *logger) Overload(amount int) {
	l.stackDepthOverload += amount
}

func findFuncLine(lines []string, lineNumber int) int {
	for i := lineNumber; i > 0; i-- {
		if regexpFuncLine.Match([]byte(lines[i])) {
			return i
		}
	}

	return -1
}

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

func (l *logger) PrintLines(filepath string, lineNumber int) {
	fmt.Printf("line %d of %s:%d\n", lineNumber, filepath, lineNumber)

	b, err := afero.ReadFile(fs, filepath)
	if err != nil {
		panic(fmt.Sprintf("cannot read file '%s': %s;", filepath, err))
	}
	lines := strings.Split(string(b), "\n")

	// set lines range
	minLine := lineNumber - l.Config.LinesBefore
	maxLine := lineNumber + l.Config.LinesAfter
	if minLine < 0 {
		minLine = 0
	}
	if maxLine > len(lines)-1 {
		maxLine = len(lines) - 1
	}

	lines = lines[:maxLine+1] //free some memory

	//find func line and correct minLine if necessary
	funcLine := findFuncLine(lines, lineNumber)
	if funcLine > minLine {
		minLine = funcLine + 1
	}

	//print func on first line
	if funcLine != -1 && funcLine < minLine {
		fmt.Println(color.RedString("%d: %s", funcLine+1, lines[funcLine]))
		if funcLine < minLine-1 {
			fmt.Println(color.YellowString("..."))
		}
	}

	//free some memory
	lines = lines[minLine:]
	maxLine -= minLine
	startLine := minLine
	minLine = 0

	//clean blank lines at the end
	for maxLine >= minLine {
		if strings.Trim(lines[maxLine], " \n\t") != "" {
			break
		}
		maxLine--
	}

	lines = lines[:maxLine+1]

	//print lines of code
	for i := minLine; i <= maxLine; i++ {
		if i+1 == lineNumber {
			fmt.Println(color.RedString("%d: %s", i+1+startLine, lines[i]))
			continue
		}
		fmt.Println(color.YellowString("%d: %s", i+1+startLine, lines[i]))
	}
}

func printStack(stages []string) {
	for i := range stages {
		for j := -1; j < i; j++ {
			fmt.Printf("  ")
		}
		fmt.Printf("%s:%s\n", regexpCallArgs.FindString(stages[i]), strings.Split(regexpCodeReference.FindString(stages[i]), ":")[1])
	}
}

func Debug(uErr error) {
	DefaultLogger.Overload(1)
	DefaultLogger.Debug(uErr)
}

func getStackTrace(deltaDepth int) []string {
	return regexpParseStack.FindAllString(string(debug.Stack()), -1)[2+deltaDepth:]
}

//PrintStack prints the stack
func PrintStack() {
	printStack(getStackTrace(1))
}

func PrintStackMinus(depthToRemove int) {
	printStack(getStackTrace(1 + depthToRemove))
}
