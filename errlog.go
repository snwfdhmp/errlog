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
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/afero"
)

//Logger interface allows to log an error, or to print source code lines. Check out NewLogger function to learn more about Logger objects and Config.
type Logger interface {
	// Debug wraps up Logger debugging funcs related to an error
	// If the given error is nil, it returns immediately
	// It relies on Logger.Config to determine what will be printed or executed
	// It returns whether err != nil
	Debug(err error) bool
	//PrintSource prints certain lines of source code of a file, using (*logger).Config as configurations
	PrintSource(filename string, lineNumber int)
	//SetConfig replaces current config with the given one
	SetConfig(cfg *Config)
	//Config returns current config
	Config() *Config
}

type logger struct {
	config             *Config
	stackDepthOverload int
}

//NewLogger creates a new logger struct with given config
func NewLogger(cfg *Config) Logger {
	return &logger{
		config:             cfg,
		stackDepthOverload: 0,
	}
}

func (l *logger) SetConfig(cfg *Config) {
	l.config = cfg
}

func (l *logger) Config() *Config {
	return l.config
}

//Config holds the configuration for a logger
type Config struct {
	LinesBefore        int  //How many lines to print *before* the error line when printing source code
	LinesAfter         int  //How many lines to print *after* the error line when printing source code
	PrintStack         bool //Shall we print stack trace ? yes/no
	PrintSource        bool //Shall we print source code along ? yes/no
	PrintError         bool //Shall we print the error of Debug(err) ? yes/no
	ExitOnDebugSuccess bool //Shall we os.Exit(1) after Debug has finished logging everything ? (doesn't happen when err is nil)
}

var (
	/*
		Note for contributors/users : these regexp have been made by me, taking my own source code as example for finding the right one to use.
		I use gofmt for source code formatting, that means this will work on most cases.
		Unfortunately, I didn't check against other code formatting tools, so it may require some evolution.
		Feel free to create an issue or send a PR.
	*/
	regexpParseStack    = regexp.MustCompile(`((((([a-zA-Z._-]+)[/])*)(([(*a-zA-Z0-9)])*(\.))+[a-zA-Z0-9]+[(](.*)[)])[\s]+[/a-zA-Z0-9\.]+[:][0-9]+)`)
	regexpCodeReference = regexp.MustCompile(`[/a-zA-Z0-9\.]+[:][0-9]+`)
	regexpCallArgs      = regexp.MustCompile(`((([a-zA-Z._-]+)[/])*)(([(*a-zA-Z0-9)])*(\.))+[a-zA-Z0-9]+[(](.*)[)]`)
	regexpCallingObject = regexp.MustCompile(`((([a-zA-Z._-]+)[/])*)(([(*a-zA-Z0-9)])*(\.))+[a-zA-Z0-9]+`)
	regexpFuncLine      = regexp.MustCompile(`^func[\s][a-zA-Z0-9]+[(](.*)[)][\s]*{`)

	//DefaultLogger logger implements default configuration for a logger
	DefaultLogger = &logger{
		config: &Config{
			LinesBefore:        4,
			LinesAfter:         2,
			PrintStack:         false,
			PrintSource:        true,
			PrintError:         true,
			ExitOnDebugSuccess: false,
		},
	}

	linesBefore = 4
	linesAfter  = 2

	fs = afero.NewOsFs() //fs is at package level because I think it needn't be scoped to loggers
)

// Debug wraps up Logger debugging funcs related to an error
// If the given error is nil, it returns immediately
// It relies on Logger.Config to determine what will be printed or executed
func (l *logger) Debug(uErr error) bool {
	if uErr == nil {
		return false
	}
	stages := getStackTrace(1 + l.stackDepthOverload)
	l.stackDepthOverload = 0
	if l.config.PrintError {
		fmt.Printf("\nError in %s: %s\n", regexpCallArgs.FindString(stages[0]), color.YellowString(uErr.Error()))
	}

	if l.config.PrintSource {
		filepath, lineNumber := parseRef(stages[0])
		l.PrintSource(filepath, lineNumber)
	}

	if l.config.PrintStack {
		fmt.Println("Stack trace:")
		printStack(stages)
	}

	if l.config.ExitOnDebugSuccess {
		os.Exit(1)
	}

	return true
}

//Overload adds depths to remove when parsing next stack trace
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

//PrintSource prints certain lines of source code of a file, using (*logger).config as configurations
func (l *logger) PrintSource(filepath string, lineNumber int) {
	fmt.Printf("line %d of %s:%d\n", lineNumber, filepath, lineNumber)

	b, err := afero.ReadFile(fs, filepath)
	if err != nil {
		panic(fmt.Sprintf("cannot read file '%s': %s;", filepath, err))
	}
	lines := strings.Split(string(b), "\n")

	// set lines range
	minLine := lineNumber - l.config.LinesBefore
	maxLine := lineNumber + l.config.LinesAfter
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

	//clean blank lines at the start
	for minLine <= maxLine {
		if strings.Trim(lines[minLine], " \n\t") != "" {
			break
		}
		minLine++
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
		if i+2 == lineNumber-startLine {
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

//Debug is a shortcut for DefaultLogger.Debug.
func Debug(uErr error) bool {
	DefaultLogger.Overload(1) // Prevents from adding this func to the stack trace
	return DefaultLogger.Debug(uErr)
}

//getStackTrace parses stack trace from runtime/debug.Stack() and returns it (minus 2 depths for (i) runtime/debug.Stack (ii) itself)
func getStackTrace(deltaDepth int) []string {
	return regexpParseStack.FindAllString(string(debug.Stack()), -1)[2+deltaDepth:]
}

//PrintStack prints the current stack trace
func PrintStack() {
	printStack(getStackTrace(1))
}

//PrintStackMinus prints the current stack trace minus the amount of depth in parameter
func PrintStackMinus(depthToRemove int) {
	printStack(getStackTrace(1 + depthToRemove))
}
