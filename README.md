[![Go Report Card](https://goreportcard.com/badge/github.com/snwfdhmp/errlog)](https://goreportcard.com/report/github.com/snwfdhmp/errlog) [![Documentation](https://godoc.org/github.com/snwfdhmp/errlog?status.svg)](http://godoc.org/github.com/snwfdhmp/errlog) [![GitHub issues](https://img.shields.io/github/issues/snwfdhmp/errlog.svg)](https://github.com/snwfdhmp/errlog/issues) [![license](https://img.shields.io/github/license/snwfdhmp/errlog.svg?maxAge=6000)](https://github.com/snwfdhmp/errlog/LICENSE) 

# A simple package to enhance Go source code debugging

![Example](https://i.imgur.com/Ulf1RGw.png)

## Introduction

Use errlog to enhance your error logging with :

- Code source highlight
- Failing func recognition
- Readable stack trace

## Get started

### Install

```shell
go get github.com/snwfdhmp/errlog
```

### Import

```golang
import "github.com/snwfdhmp/errlog"
```

### Usage

Now, replace some `if err != nil` with `if errlog.Debug(err)` to add debugging informations.

```golang
func someFunc() {
    //...
    if errlog.Debug(err) { // will debug & pass if err != nil, will ignore if err == nil
        return
    }
}
```

## Configure like you need

You can configure your own logger with these options :

```golang
type Config struct {
    PrintFunc          func(format string, data ...interface{}) //Printer func (eg: fmt.Printf)
    LinesBefore        int  //How many lines to print *before* the error line when printing source code
    LinesAfter         int  //How many lines to print *after* the error line when printing source code
    PrintStack         bool //Shall we print stack trace ? yes/no
    PrintSource        bool //Shall we print source code along ? yes/no
    PrintError         bool //Shall we print the error of Debug(err) ? yes/no
    ExitOnDebugSuccess bool //Shall we os.Exit(1) after Debug has finished logging everything ? (doesn't happen when err is nil). Will soon be replaced by ExitFunc to enable panic-ing the current goroutine. (if you need this quick, please open an issue)
}
```

> This definition may be outdated, visit the [Config struct definition in godoc.org](https://godoc.org/github.com/snwfdhmp/errlog#Config) for the up to date definition


## Example

We will use this sample program :

```golang
//someSmallFunc represents any func
func someSmallFunc() {
    fmt.Println("I do things !")
}

//someBigFunc represents any func having to handle errors from other funcs
func someBigFunc() {
    someSmallFunc()

    if err := someNastyFunc(); errlog.Debug(err) { //here, he want to catch an error
        return
    }

    someSmallFunc()
}

//someNastyFunc represents any failing func
func someNastyFunc() error {
    return errors.New("I'm failing for no reason")
}

func main() {
    fmt.Println("Start of the program")
    wrappingFunc()
    fmt.Println("End of the program")
}

func wrappingFunc() {
    someBigFunc()
}
```

### Output

![Console Output examples/basic.go](https://i.imgur.com/tOkDgwP.png)

### Example

Now let's see what we can do with a custom configuration.

```golang
debug := errlog.NewLogger(&errlog.Config{
    // PrintFunc is of type `func (format string, data ...interface{})`
    // so you can easily implement your own logger func.
    // In this example, logrus is used, but any other logger can be used.
    // Beware that you should add '\n' at the end of format string when printing.
    PrintFunc:          logrus.Printf,
    PrintSource:        true, //Print the failing source code
    LinesBefore:        2, //Print 2 lines before failing line
    LinesAfter:         1, //Print 1 line after failing line
    PrintError:         true, //Print the error
    PrintStack:         false, //Don't print the stack trace
    ExitOnDebugSuccess: true, //Exit if err
})
```

> This definition may be outdated, visit the [Config struct definition in godoc.org](https://godoc.org/github.com/snwfdhmp/errlog#Config) for the up to date definition

### Output

![Console Output examples/custom.go](https://i.imgur.com/vh2iEnS.png)


### Another Example

Errlog finds the exact line where the error is defined.

### Output

![Source Example: error earlier in the code](https://i.imgur.com/wPBrYqs.png)

## Documentation

Documentation can be found here : [![Documentation](https://godoc.org/github.com/snwfdhmp/errlog?status.svg)](http://godoc.org/github.com/snwfdhmp/errlog)

## Feedback

Feel free to open an issue for any feedback or suggestion.

I fix process issues quickly.

## Contributions

PR are accepted as soon as they follow Golang common standards.
For more information: https://golang.org/doc/effective_go.html

## License information

[![license](https://img.shields.io/github/license/snwfdhmp/errlog.svg?maxAge=60000)](https://github.com/snwfdhmp/errlog/LICENSE)