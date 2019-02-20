package errlog

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

var (
	gopath = os.Getenv("GOPATH")
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

//Config holds the configuration for a logger
type Config struct {
	PrintFunc               func(format string, data ...interface{}) //Printer func (eg: fmt.Printf)
	LinesBefore             int                                      //How many lines to print *before* the error line when printing source code
	LinesAfter              int                                      //How many lines to print *after* the error line when printing source code
	PrintStack              bool                                     //Shall we print stack trace ? yes/no
	PrintSource             bool                                     //Shall we print source code along ? yes/no
	PrintError              bool                                     //Shall we print the error of Debug(err) ? yes/no
	ExitOnDebugSuccess      bool                                     //Shall we os.Exit(1) after Debug has finished logging everything ? (doesn't happen when err is nil)
	DisableStackIndentation bool                                     //Shall we print stack vertically instead of indented
}

type logger struct {
	config             *Config
	stackDepthOverload int
}

//NewLogger creates a new logger struct with given config
func NewLogger(cfg *Config) Logger {
	l := logger{
		config:             cfg,
		stackDepthOverload: 0,
	}

	l.Doctor()

	return &l
}

// Debug wraps up Logger debugging funcs related to an error
// If the given error is nil, it returns immediately
// It relies on Logger.Config to determine what will be printed or executed
func (l *logger) Debug(uErr error) bool {
	l.Doctor()
	if uErr == nil {
		return false
	}

	stages := getStackTrace(1 + l.stackDepthOverload)
	if len(stages) < 1 {
		l.Debug(errors.New("cannot read stack trace"))
		return true
	}

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

//DebugSource prints certain lines of source code of a file for debugging, using (*logger).config as configurations
func (l *logger) DebugSource(filepath string, debugLineNumber int) {
	filepathShort := filepath
	if gopath != "" {
		filepathShort = strings.Replace(filepath, gopath+"/src/", "", -1)
	}

	b, err := afero.ReadFile(fs, filepath)
	if err != nil {
		err = fmt.Errorf("cannot read file '%s': %s;", filepath, err)
		l.Debug(err)
	}
	lines := strings.Split(string(b), "\n")

	// set line range to print based on config values and debugLineNumber
	// and correct ouf of range values
	minLine := max(debugLineNumber-l.config.LinesBefore, 0)
	maxLine := min(debugLineNumber+l.config.LinesAfter, len(lines)-1)

	deleteBankLinesFromRange(lines, &minLine, &maxLine)

	//free some memory from unused values
	lines = lines[:maxLine+1]

	//find func line and adjust minLine if below
	funcLine := findFuncLine(lines, debugLineNumber)
	if funcLine > minLine {
		minLine = funcLine + 1
	}

	//try to find failing line if any
	failingLineIndex, columnStart, columnEnd := findFailingLine(lines, funcLine, debugLineNumber)

	if failingLineIndex != -1 {
		l.Printf("line %d of %s:%d", failingLineIndex+1, filepathShort, failingLineIndex+1)
	} else {
		l.Printf("error in %s (failing line not found, stack trace says func call is at line %d)", filepathShort, debugLineNumber+1)
	}

	l.PrintSource(lines, PrintSourceOptions{
		FuncLine: funcLine,
		Highlighted: map[int][]int{
			failingLineIndex: []int{columnStart, columnEnd},
		},
		StartLine: minLine,
		EndLine:   maxLine,
	})
}

// PrintSource prints source code based on opts
func (l *logger) PrintSource(lines []string, opts PrintSourceOptions) {
	//print func on first line
	if opts.FuncLine != -1 && opts.FuncLine < opts.StartLine {
		l.Printf("%s", color.RedString("%d: %s", opts.FuncLine+1, lines[opts.FuncLine]))
		if opts.FuncLine < opts.StartLine-1 { // append blank line if minLine is not next line
			l.Printf("%s", color.YellowString("..."))
		}
	}

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
			l.Printf("%d: %s", i+1, color.YellowString(lines[i]))
		} else { // line with highlightings
			logrus.Debugf("Next line should be highlighted from column %d to %d.", highlightStart, highlightEnd)
			l.Printf("%d: %s%s%s", i+1, color.YellowString(lines[i][:highlightStart]), color.RedString(lines[i][highlightStart:highlightEnd+1]), color.YellowString(lines[i][highlightEnd+1:]))
		}
	}
}

func (l *logger) Doctor() (neededDoctor bool) {
	neededDoctor = false
	if l.config.PrintFunc == nil {
		neededDoctor = true
		logrus.Debug("PrintFunc not set for this logger. Replacing with DefaultLoggerPrintFunc.")
		l.config.PrintFunc = DefaultLoggerPrintFunc
	}

	if l.config.LinesBefore < 0 {
		neededDoctor = true
		logrus.Debugf("LinesBefore is '%d' but should not be <0. Setting to 0.", l.config.LinesBefore)
		l.config.LinesBefore = 0
	}

	if l.config.LinesAfter < 0 {
		neededDoctor = true
		logrus.Debugf("LinesAfters is '%d' but should not be <0. Setting to 0.", l.config.LinesAfter)
		l.config.LinesAfter = 0
	}

	if neededDoctor && !debugMode {
		logrus.Warn("errlog: Doctor() has detected and fixed some problems on your logger configuration. It might have modified your configuration. Check logs by enabling debug. 'errlog.SetDebugMode(true)'.")
	}

	return
}

//Printf is the function used to log
func (l *logger) Printf(format string, data ...interface{}) {
	l.config.PrintFunc(format, data...)
}

//Overload adds depths to remove when parsing next stack trace
func (l *logger) Overload(amount int) {
	l.stackDepthOverload += amount
}

func (l *logger) SetConfig(cfg *Config) {
	l.config = cfg
	l.Doctor()
}

func (l *logger) Config() *Config {
	return l.config
}
