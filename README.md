# A simple package to enhance Go source code debugging

![Example](https://i.imgur.com/wPBrYqs.png)

## Introduction

Use errlog to enhance your error logging with :

- Code source highlight
- Failing func recognition
- Readable stack trace



## Get started

### Install

```
go get github.com/snwfdhmp/errlog
```

### Import

```golang
import "github.com/snwfdhmp/errlog"
```

### Usage

```golang
func someFunc() {
    //...
    if errlog.Debug(err) { // will debug & pass if err != nil, will ignore if err == nil
        return
    }
}
```

## Example

We will use this sample program :

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

### Output

![Console Output examples/basic.go](https://i.imgur.com/tOkDgwP.png)

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

### Example

> In this example, logrus is used, but any other logger can be used. PrintFunc is of type `func (format string, data ...interface{})`, so you can easily implement your own logger func. Beware that you should add '\n' at the end of format string when printing.

Now using a custom configuration.

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
```

### Output

![Console Output examples/custom.go](https://i.imgur.com/vh2iEnS.png)


### Another Example

Errlog finds the exact line where the error is defined.

### Output

![Source Example: error earlier in the code](https://i.imgur.com/wPBrYqs.png)

## Feedback

Feel free to open an issue for any feedback or suggestion.

I fix bugs quickly.

## Contributions

PR are accepted as soon as they follow Golang common standards.
For more information: https://golang.org/doc/effective_go.html
