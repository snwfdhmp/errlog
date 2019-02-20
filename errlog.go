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
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

//Logger interface allows to log an error, or to print source code lines. Check out NewLogger function to learn more about Logger objects and Config.
type Logger interface {
	// Debug wraps up Logger debugging funcs related to an error
	// If the given error is nil, it returns immediately
	// It relies on Logger.Config to determine what will be printed or executed
	// It returns whether err != nil
	Debug(err error) bool
	//PrintSource prints lines based on given opts (see PrintSourceOptions type definition)
	PrintSource(lines []string, opts PrintSourceOptions)
	//DebugSource debugs a source file
	DebugSource(filename string, lineNumber int)
	//SetConfig replaces current config with the given one
	SetConfig(cfg *Config)
	//Config returns current config
	Config() *Config
}

//SetDebugMode sets debug mode to On if toggle==true or Off if toggle==false. It changes log level an so displays more logs about whats happening. Useful for debugging.
func SetDebugMode(toggle bool) {
	if toggle {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}

type logger struct {
	config             *Config
	stackDepthOverload int
}

//Printf is the function used to log
func (l *logger) Printf(format string, data ...interface{}) {
	l.config.PrintFunc(format, data...)
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
	PrintFunc          func(format string, data ...interface{}) //Printer func (eg: fmt.Printf)
	LinesBefore        int                                      //How many lines to print *before* the error line when printing source code
	LinesAfter         int                                      //How many lines to print *after* the error line when printing source code
	PrintStack         bool                                     //Shall we print stack trace ? yes/no
	PrintSource        bool                                     //Shall we print source code along ? yes/no
	PrintError         bool                                     //Shall we print the error of Debug(err) ? yes/no
	ExitOnDebugSuccess bool                                     //Shall we os.Exit(1) after Debug has finished logging everything ? (doesn't happen when err is nil)
}

var (
	/*
		Note for contributors/users : these regexp have been made by me, taking my own source code as example for finding the right one to use.
		I use gofmt for source code formatting, that means this will work on most cases.
		Unfortunately, I didn't check against other code formatting tools, so it may require some evolution.
		Feel free to create an issue or send a PR.
	*/
	regexpParseStack                 = regexp.MustCompile(`((((([a-zA-Z._-]+)[/])*)(([(*a-zA-Z0-9)])*(\.))+[a-zA-Z0-9]+[(](.*)[)])[\s]+[/a-zA-Z0-9\.]+[:][0-9]+)`)
	regexpCodeReference              = regexp.MustCompile(`[/a-zA-Z0-9\.]+[:][0-9]+`)
	regexpCallArgs                   = regexp.MustCompile(`((([a-zA-Z._-]+)[/])*)(([(*a-zA-Z0-9)])*(\.))+[a-zA-Z0-9]+[(](.*)[)]`)
	regexpCallingObject              = regexp.MustCompile(`((([a-zA-Z._-]+)[/])*)(([(*a-zA-Z0-9)])*(\.))+[a-zA-Z0-9]+`)
	regexpFuncLine                   = regexp.MustCompile(`^func[\s][a-zA-Z0-9]+[(](.*)[)][\s]*{`)
	regexpParseDebugLineFindFunc     = regexp.MustCompile(`[\.]Debug[\(](.*)[/)]`)
	regexpParseDebugLineParseVarName = regexp.MustCompile(`[\.]Debug[\(](.*)[/)]`)
	regexpFindVarDefinition          = func(varName string) *regexp.Regexp {
		return regexp.MustCompile(fmt.Sprintf(`%s[\s\:]*={1}([\s]*[a-zA-Z0-9\._]+)`, varName))
	}

	//DefaultLoggerPrintFunc is fmt.Printf without return values
	DefaultLoggerPrintFunc = func(format string, data ...interface{}) {
		fmt.Printf(format+"\n", data...)
	}

	//DefaultLogger logger implements default configuration for a logger
	DefaultLogger = &logger{
		config: &Config{
			PrintFunc:          DefaultLoggerPrintFunc,
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

func (l *logger) Doctor() (neededDoctor bool) {
	neededDoctor = false
	if l.config.PrintFunc == nil {
		neededDoctor = true
		logrus.Debug("PrintFunc not set for this logger. Replacing with DefaultLoggerPrintFunc.")
		l.config.PrintFunc = DefaultLoggerPrintFunc
	}

	return
}

// Debug wraps up Logger debugging funcs related to an error
// If the given error is nil, it returns immediately
// It relies on Logger.Config to determine what will be printed or executed
func (l *logger) Debug(uErr error) bool {
	if l.Doctor() {
		logrus.Warn("Doctor() has detected and fixed some problems. It might have modified your configuration. Check logs by enabling debug. 'errlog.SetDebugMode(true)'.")
	}
	if uErr == nil {
		return false
	}
	stages := getStackTrace(1 + l.stackDepthOverload)
	if len(stages) < 1 {
		l.Debug(errors.New("cannot read stack trace"))
		return true
	}
	l.stackDepthOverload = 0
	if l.config.PrintError {
		l.Printf("Error in %s: %s", regexpCallArgs.FindString(stages[0]), color.YellowString(uErr.Error()))
	}

	if l.config.PrintSource {
		filepath, lineNumber := parseRef(stages[0])
		l.DebugSource(filepath, lineNumber)
	}

	if l.config.PrintStack {
		l.Printf("Stack trace:")
		l.printStack(stages)
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

// PrintSource prints source code
// Order :
// 1. FuncLine
// 2.
func (l *logger) PrintSource(lines []string, opts PrintSourceOptions) {
	//print func on first line
	if opts.FuncLine != -1 && opts.FuncLine < opts.StartLine {
		l.Printf("%s", color.RedString("%d: %s", opts.FuncLine+1, lines[opts.FuncLine]))
		if opts.FuncLine < opts.StartLine-1 { // append blank line if minLine is not next line
			l.Printf("%s", color.YellowString("..."))
		}
	}

	// === Free memory part saved from former PrintSource func, needs a bit of refactoring of new func to avoid breaking lines count
	//
	// //free some memory
	// lines = lines[minLine:]
	// //adjust variables with new slice specs
	// maxLine -= minLine
	// startLine := minLine // save this number still
	// minLine = 0

	// lines = lines[:maxLine+1]

	for i := opts.StartLine; i < opts.EndLine; i++ {
		highlightStart := -1
		highlightEnd := -1
		if _, ok := opts.Highlighted[i]; ok {
			if len(opts.Highlighted[i]) == 2 { //if hightlight slice is in the right format
				highlightStart = opts.Highlighted[i][0]
				highlightEnd = opts.Highlighted[i][1]
				if highlightEnd > len(lines[i])-1 {
					highlightEnd = len(lines[i]) - 1
				}
			} else {
				logrus.Debug("len(opts.Highlighted[i]) != 2; skipping highlight")
			}
		}

		if highlightStart == -1 { //simple line
			l.Printf("%d: %s", i+opts.StartLine+1, color.YellowString(lines[i]))
		} else { // line with highlightings
			logrus.Debugf("Next line should be highlighted from column %d to %d.", highlightStart, highlightEnd)
			l.Printf("%d: %s%s%s", i+opts.StartLine+1, color.YellowString(lines[i][:highlightStart]), color.RedString(lines[i][highlightStart:highlightEnd+1]), color.YellowString(lines[i][highlightEnd+1:]))
		}
	}
}

//DebugSource prints certain lines of source code of a file for debugging, using (*logger).config as configurations
func (l *logger) DebugSource(filepath string, debugLineNumber int) {
	l.Printf("line %d of %s:%d", debugLineNumber, filepath, debugLineNumber)

	b, err := afero.ReadFile(fs, filepath)
	if err != nil {
		panic(fmt.Sprintf("cannot read file '%s': %s;", filepath, err))
	}
	lines := strings.Split(string(b), "\n")

	// set line range to print based on config values and debugLineNumber
	minLine := debugLineNumber - l.config.LinesBefore
	maxLine := debugLineNumber + l.config.LinesAfter

	// correct ouf of range values
	if minLine < 0 {
		minLine = 0
	}
	if maxLine > len(lines)-1 {
		maxLine = len(lines) - 1
	}

	//clean leading blank lines
	for minLine <= maxLine {
		if strings.Trim(lines[minLine], " \n\t") != "" {
			break
		}
		minLine++
	}

	//clean trailing blank lines
	for maxLine >= minLine {
		if strings.Trim(lines[maxLine], " \n\t") != "" {
			break
		}
		maxLine--
	}

	//free some memory from unused values
	lines = lines[:maxLine+1]

	//find func line and adjust minLine if below
	funcLine := findFuncLine(lines, debugLineNumber)
	if funcLine > minLine {
		minLine = funcLine + 1
	}

	//try to find failing line if any
	failingLineIndex, columnStart, columnEnd := findFailingLine(lines, funcLine, debugLineNumber)

	l.PrintSource(lines, PrintSourceOptions{
		FuncLine: funcLine,
		Highlighted: map[int][]int{
			failingLineIndex: []int{columnStart, columnEnd},
		},
		StartLine: minLine,
		EndLine:   maxLine,
	})
}

func (l *logger) printStack(stages []string) {
	for i := range stages {
		for j := -1; j < i; j++ {
			l.Printf("  ")
		}
		l.Printf("%s:%s\n", regexpCallArgs.FindString(stages[i]), strings.Split(regexpCodeReference.FindString(stages[i]), ":")[1])
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
	DefaultLogger.printStack(getStackTrace(1))
}

//PrintStackMinus prints the current stack trace minus the amount of depth in parameter
func PrintStackMinus(depthToRemove int) {
	DefaultLogger.printStack(getStackTrace(1 + depthToRemove))
}

func findFuncLine(lines []string, lineNumber int) int {
	for i := lineNumber; i > 0; i-- {
		if regexpFuncLine.Match([]byte(lines[i])) {
			return i
		}
	}

	return -1
}

func findFailingLine(lines []string, funcLine int, debugLine int) (failingLineIndex, columnStart, columnEnd int) {
	failingLineIndex = -1 //init error flag
	reMatches := regexpParseDebugLineParseVarName.FindStringSubmatch(lines[debugLine-1])
	if len(reMatches) < 2 {
		return
	}
	varName := reMatches[1]
	reFindVar := regexpFindVarDefinition(varName)
	for i := debugLine; i >= funcLine && i > 0; i-- {
		logrus.Debugf("%d: %s", i, lines[i])
		if strings.Trim(lines[i], " \n\t") == "" {
			logrus.Debugf(color.BlueString("%d: ignoring blank line", i))
			continue
		} else if len(lines[i]) >= 2 && lines[i][:2] == "//" {
			logrus.Debugf(color.BlueString("%d: ignoring comment line", i))
			continue
		}
		index := reFindVar.FindStringSubmatchIndex(lines[i])
		if index == nil {
			logrus.Debugf(color.BlueString("%d: var definition not found for '%s' (regexp no match).", i, varName))
			continue
		}

		failingLineIndex = i
		columnStart = index[0]
		openedBrackets, closedBrackets := 0, 0
		for j := index[1]; j < len(lines[i]); j++ {
			if lines[i][j] == '(' {
				openedBrackets++
			} else if lines[i][j] == ')' {
				closedBrackets++
			}
			if openedBrackets == closedBrackets {
				columnEnd = j
				return
			}
		}

		if columnEnd == 0 {
			logrus.Debugf("Correcting columnEnd [0]. We failed to find err definition.")
			columnEnd = len(lines[i]) - 1
		}
		return
	}

	return
}
