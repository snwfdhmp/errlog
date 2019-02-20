# A simple object to enhance Go source code debugging

## Get started

Install via

```
go get github.com/snwfdhmp/errlog
```

Import via

```golang
import (
    "github.com/snwfdhmp/errlog"
)
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
	LinesBefore        int
	LinesAfter         int
	PrintStack         bool
	PrintSource        bool
	PrintError         bool
	ExitOnDebugSuccess bool
}
```

Example :

```golang
debug := errlog.NewLogger(&errlog.Config{
	LinesBefore:        2,
	LinesAfter:         1,
	PrintError:         true,
	PrintSource:        true,
	PrintStack:         false,
	ExitOnDebugSuccess: true,
})
````

Outputs :

```
Error in main.someBigFunction(): I'm failing for no reason
line 41 of /Users/snwfdhmp/go/src/github.com/snwfdhmp/sandbox/testerr.go:41
33: func someBigFunction() {
...
40:     if err := someNastyFunction(); debug.Debug(err) {
41:             return
42:     }
exit status 1
```

## Feedback

Feel free to open an issue for any feedback.

If you report bugs I fix them asap.

## Contributions

PR are accepted as soon as they follow Golang common standards.
For more information: https://golang.org/doc/effective_go.html
