/*
errlog package aims to simplify Golang program debugging.

Example result:

```
$ go run myfailingapp.go
Program starting
error in main.main: something failed here
line 13 of /Users/snwfdhmp/go/src/github.com/snwfdhmp/sandbox/testerr.go
9: func main() {
10:     fmt.Println("Program starting")
11:     err := errors.New("something failed here")
12:
13:     errlog.Debug(err)
14:
15:     fmt.Println("End of the program")
16: }
exit status 1
```
*/
package errlog

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/afero"
)

var (
	regexpParseStack    = regexp.MustCompile(`((((([a-zA-Z]+)[/])*)(([(*a-zA-Z0-9)])*(\.))+[a-zA-Z0-9]+[(](.*)[)])[\s]+[/a-zA-Z0-9\.]+[:][0-9]+)`)
	regexpCodeReference = regexp.MustCompile(`[/a-zA-Z0-9\.]+[:][0-9]+`)
	regexpCallArgs      = regexp.MustCompile(`((([a-zA-Z]+)[/])*)(([(*a-zA-Z0-9)])*(\.))+[a-zA-Z0-9]+[(](.*)[)]`)
	regexpCallingObject = regexp.MustCompile(`((([a-zA-Z]+)[/])*)(([(*a-zA-Z0-9)])*(\.))+[a-zA-Z0-9]+`)
	regexpFuncLine      = regexp.MustCompile(`^func[\s][a-zA-Z0-9]+[(](.*)[)][\s]*{`)

	fs = afero.NewOsFs()
)

// Debug prints useful informations for debug such as surrounding code, stack trace, ...
func Debug(uErr error) {
	stages := regexpParseStack.FindAllString(string(debug.Stack()), -1)

	ref := strings.Split(regexpCodeReference.FindString(stages[2]), ":")
	if len(ref) != 2 {
		panic(fmt.Sprintf("len(ref) > 2;ref='%s';", ref))
	}
	filepath := ref[0]
	lineNumber, err := strconv.Atoi(ref[1])
	if err != nil {
		panic(fmt.Sprintf("cannot parse line number '%s': %s", ref[1], err))
	}

	fmt.Printf("\nerror in %s: %s\nline %d of %s:%d\n", regexpCallingObject.FindString(stages[2]), color.YellowString(uErr.Error()), lineNumber, filepath, lineNumber)

	printLines(filepath, lineNumber)
	fmt.Println("Stack trace:")
	printStack(stages[2:])
}

func findFuncLine(lines []string, lineNumber int) int {
	for i := lineNumber; i > 0; i-- {
		if regexpFuncLine.Match([]byte(lines[i])) {
			return i
		}
	}

	return -1
}

func printLines(filepath string, lineNumber int) {
	b, err := afero.ReadFile(fs, filepath)
	if err != nil {
		panic(fmt.Sprintf("cannot read file '%s': %s;", filepath, err))
	}

	lines := strings.Split(string(b), "\n")
	minLine := lineNumber - 10 //@todo to int
	maxLine := lineNumber + 6
	if minLine < 0 {
		minLine = 0
	}
	if maxLine > len(lines)-1 {
		maxLine = len(lines) - 1
	}

	funcLine := findFuncLine(lines, lineNumber)
	if funcLine != -1 && funcLine < minLine {
		fmt.Println(color.RedString("%d: %s", funcLine+1, lines[funcLine]))
		if funcLine < minLine-1 {
			fmt.Println(color.YellowString("..."))
		}
	}
	if funcLine > minLine {
		minLine = funcLine
	}
	for i := minLine; i < maxLine; i++ {
		if i+1 == lineNumber {
			fmt.Println(color.RedString("%d: %s", i+1, lines[i]))
			continue
		}
		fmt.Println(color.YellowString("%d: %s", i+1, lines[i]))
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
