# A simple package to enhance Go source code debugging

## Get started

1. Install with

```
go get github.com/snwfdhmp/errlog
```

2. Import with

```golang
import (
    "github.com/snwfdhmp/errlog"
)
```

2. Use with

```golang
err := someFunc()
if errlog.Debug(err) {
	return
}
```

## Usage

```golang
func someFunc() {
    //...
    if errlog.Debug(err) { // will debug & pass if err != nil, will ignore if err == nil
        return
    }
}
```

## Example

We are going to use this sample program :

```golang
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
	someSmallFunction()

	if err := someNastyFunction(); errlog.Debug(err) {
		return
	}

	someSmallFunction()
}

func someSmallFunction() {
	fmt.Println("I do things !")
}

func someNastyFunction() error {
	return errors.New("I'm failing for no reason")
}
```

Output :

```
$ go run main.go
Start of the program
I do things !

error in main.someBigFunction: I'm failing for no reason
line 26 of /Users/snwfdhmp/go/src/github.com/snwfdhmp/sandbox/testerr.go:26
22: func someBigFunction() {
23: 	someSmallFunction()
24: 
25: 	if err := someNastyFunction(); errlog.Debug(err) {
26: 		return
27: 	}
28: 
29: 	someSmallFunction()
30: }
31: 
Stack trace:
  main.someBigFunction():26
    main.wrapingFunc():19
      main.main():13
exit status 1
```

## Configure like you need

You can configure your own logger with these options :

```golang
type Config struct {
	PrintFunc          func(format string, data ...interface{}) //Printer func (eg: fmt.Printf)
	LinesBefore        int  					//How many lines to print *before* the error line when printing source code
	LinesAfter         int 						//How many lines to print *after* the error line when printing source code
	PrintStack         bool 					//Shall we print stack trace ? yes/no
	PrintSource        bool 					//Shall we print source code along ? yes/no
	PrintError         bool 					//Shall we print the error of Debug(err) ? yes/no
	ExitOnDebugSuccess bool 					//Shall we os.Exit(1) after Debug has finished logging everything ? (doesn't happen when err is nil)
}
```

Example :

```golang
debug := errlog.NewLogger(&errlog.Config{
	PrintFunc:          logrus.Printf,
	LinesBefore:        2,
	LinesAfter:         1,
	PrintError:         true,
	PrintSource:        true,
	PrintStack:         false,
	ExitOnDebugSuccess: true,
})
````

Outputs :

![Console Output](https://i.ibb.co/yPcq4kJ/output-logrus.jpg)

## Feedback

Feel free to open an issue for any feedback.

If you report bugs I fix them asap.

## Contributions

PR are accepted as soon as they follow Golang common standards.
For more information: https://golang.org/doc/effective_go.html
