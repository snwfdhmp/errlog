# Simple error logging for Go programs

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
    if err != nil {
        errlog.Debug(err)
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

	if err := someNastyFunction(); err != nil {
		errlog.Debug(err)
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
25: 	if err := someNastyFunction(); err != nil {
26: 		errlog.Debug(err)
27: 		return
28: 	}
29: 
30: 	someSmallFunction()
31: }
32: 
Stack trace:
  main.someBigFunction():26
    main.wrapingFunc():19
      main.main():13
exit status 1
```

## Feedback

Feel free to open an issue for any feedback.

If you report bugs I fix them asap.

## Contributions

PR are accepted as soon as they follow Golang common standards.
For more information: https://golang.org/doc/effective_go.html