package errlog

import (
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
	//Disable is used to disable Logger (every call to this Logger will perform NO-OP (no operation)) and return instantly
	//Use Disable(true) to disable and Disable(false) to enable again
	Disable(bool)
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
	Mode                    int
}

// PrintSourceOptions represents config for (*logger).PrintSource func
type PrintSourceOptions struct {
	FuncLine    int
	StartLine   int
	EndLine     int
	Highlighted map[int][]int //map[lineIndex][columnstart, columnEnd] of chars to highlight
}

//logger holds logger object, implementing Logger interface
type logger struct {
	config             *Config //config for the logger
	stackDepthOverload int     //stack depth to ignore when reading stack
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
	if l.config.Mode == ModeDisabled {
		return uErr != nil
	}
	l.Doctor()
	if uErr == nil {
		return false
	}

	stLines := parseStackTrace(1 + l.stackDepthOverload)
	if stLines == nil || len(stLines) < 1 {
		l.Printf("Error: %s", uErr)
		l.Printf("Errlog tried to debug the error but the stack trace seems empty. If you think this is an error, please open an issue at https://github.com/snwfdhmp/errlog/issues/new and provide us logs to investigate.")
		return true
	}

	if l.config.PrintError {
		l.Printf("Error in %s: %s", stLines[0].CallingObject, color.YellowString(uErr.Error()))
	}

	if l.config.PrintSource {
		l.DebugSource(stLines[0].SourcePathRef, stLines[0].SourceLineRef)
	}

	if l.config.PrintStack {
		l.Printf("Stack trace:")
		l.printStack(stLines)
	}

	if l.config.ExitOnDebugSuccess {
		os.Exit(1)
	}

	l.stackDepthOverload = 0

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
		l.Printf("errlog: cannot read file '%s': %s. If sources are not reachable in this environment, you should set PrintSource=false in logger config.", filepath, err)
		return
		// l.Debug(err)
	}
	lines := strings.Split(string(b), "\n")

	// set line range to print based on config values and debugLineNumber
	minLine := debugLineNumber - l.config.LinesBefore
	maxLine := debugLineNumber + l.config.LinesAfter

	//delete blank lines from range and clean range if out of lines range
	deleteBlankLinesFromRange(lines, &minLine, &maxLine)

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
		l.Printf("error in %s (failing line not found, stack trace says func call is at line %d)", filepathShort, debugLineNumber)
	}

	l.PrintSource(lines, PrintSourceOptions{
		FuncLine: funcLine,
		Highlighted: map[int][]int{
			failingLineIndex: {columnStart, columnEnd},
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
		if _, ok := opts.Highlighted[i]; !ok || len(opts.Highlighted[i]) != 2 {
			l.Printf("%d: %s", i+1, color.YellowString(lines[i]))
			continue
		}

		hlStart := max(opts.Highlighted[i][0], 0)          //highlight column start
		hlEnd := min(opts.Highlighted[i][1], len(lines)-1) //highlight column end
		l.Printf("%d: %s%s%s", i+1, color.YellowString(lines[i][:hlStart]), color.RedString(lines[i][hlStart:hlEnd+1]), color.YellowString(lines[i][hlEnd+1:]))
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

func (l *logger) printStack(stLines []StackTraceItem) {
	for i := len(stLines) - 1; i >= 0; i-- {
		padding := ""
		if !l.config.DisableStackIndentation {
			for j := 0; j < len(stLines)-1-i; j++ {
				padding += "  "
			}
		}
		l.Printf("%s (%s:%d)", stLines[i].CallingObject, stLines[i].SourcePathRef, stLines[i].SourceLineRef)
	}
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

func (l *logger) SetMode(mode int) bool {
	if !isIntInSlice(mode, enabledModes) {
		return false
	}
	l.Config().Mode = mode
	return true
}

func (l *logger) Disable(shouldDisable bool) {
	if shouldDisable {
		l.Config().Mode = ModeDisabled
	} else {
		l.Config().Mode = ModeEnabled
	}
}

const (
	// ModeDisabled represents the disabled mode (NO-OP)
	ModeDisabled = iota + 1
	// ModeEnabled represents the enabled mode (Print)
	ModeEnabled
)

var (
	enabledModes = []int{ModeDisabled, ModeEnabled}
)
