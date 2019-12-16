# Errlog: reduce debugging time while programming [![Go Report Card](https://goreportcard.com/badge/github.com/snwfdhmp/errlog)](https://goreportcard.com/report/github.com/snwfdhmp/errlog) [![Awesome](https://cdn.rawgit.com/sindresorhus/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg)](https://github.com/sindresorhus/awesome) [![Documentation](https://godoc.org/github.com/snwfdhmp/errlog?status.svg)](http://godoc.org/github.com/snwfdhmp/errlog) [![GitHub issues](https://img.shields.io/github/issues/snwfdhmp/errlog.svg)](https://github.com/snwfdhmp/errlog/issues) [![license](https://img.shields.io/github/license/snwfdhmp/errlog.svg?maxAge=6000)](https://github.com/snwfdhmp/errlog/LICENSE)

![Example](https://i.imgur.com/Ulf1RGw.png)

## Introduction

Use errlog to improve error logging and **speed up  debugging while you create amazing code** :

- Highlight source code
- **Detect and point out** which func call is causing the fail
- Pretty stack trace
- **No-op mode** for production
- Easy implementation, adaptable logger
- Plug to any current project without changing you or your teammates habits
- Plug to **your current logging system**

|Go to|
|---|
|[Get started](#get-started)|
|[Documentation](#documentation)|
|[Examples](#example)|
|[Tweaking](#tweak-as-you-need)|
|[Feedbacks](#feedbacks)|
|[Contributions](#contributions)|
|[License](#license-information)|
|[Contributors](#contributors)|

## Get started

### Install

```shell
go get github.com/snwfdhmp/errlog
```

### Usage

Replace your `if err != nil` with `if errlog.Debug(err)` to add debugging informations.

```golang
func someFunc() {
    //...
    if errlog.Debug(err) { // will debug & pass if err != nil, will ignore if err == nil
        return
    }
}
```

In production, call `errlog.DefaultLogger.Disable(true)` to enable no-op (equivalent to `if err != nil`)

## Tweak as you need

You can configure your own logger with the following options :

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

> As we don't yet update automatically this README immediately when we add new features, this definition may be outdated. (Last update: 2019/08/07)
> [See the struct definition in godoc.org](https://godoc.org/github.com/snwfdhmp/errlog#Config) for the up to date definition


## Example

### Try yourself

| Name and link | Description |
| --- | --- |
| [Basic](examples/basic/basic.go) | standard usage, quick setup
| [Custom](examples/custom/custom.go) | guided configuration for fulfilling your needs |
| [Disabled](examples/disabled/disabled.go) | how to disable the logging & debugging (eg: for production use) |
| [Failing line far away](examples/failingLineFar/failingLineFar.go) | example of finding the func call that caused the error while it is lines away from the errlog.Debug call |
| [Pretty stack trace](examples/stackTrace/stackTrace.go) | pretty stack trace printing instead of debugging. |

### Just read

#### Basic example

> Note that in the example, you will see some unuseful func. Those are made to generate additional stack trace levels for the sake of example

We're going to use this sample program :

```golang
func main() {
    fmt.Println("Program start")

    wrapingFunc() //call to our important function

    fmt.Println("Program end")
}

func wrapingFunc() {
    someBigFunction() // call some func 
}

func someBigFunction() {
    someDumbFunction() // just random calls
    someSmallFunction() // just random calls
    someDumbFunction() // just random calls

    // Here it can fail, so instead of `if err  != nil` we use `errlog.Debug(err)`
    if err := someNastyFunction(); errlog.Debug(err) {
        return
    }

    someSmallFunction() // just random calls
    someDumbFunction() // just random calls
}

func someSmallFunction() {
    _ = fmt.Sprintf("I do things !")
}

func someNastyFunction() error {
    return errors.New("I'm failing for some reason") // simulate an error
}

func someDumbFunction() bool {
    return false // just random things
}
```


#### Output

![Console Output examples/basic.go](https://i.imgur.com/tOkDgwP.png)


We are able to **detect and point out which line is causing the error**.

### Custom Configuration Example

Let's see what we can do with a **custom configuration.**

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

> Please note: This definition may be outdated. (Last update: 2019/08/07)
> [See the struct definition in godoc.org](https://godoc.org/github.com/snwfdhmp/errlog#Config) for the up to date definition

#### Output

![Console Output examples/custom.go](https://i.imgur.com/vh2iEnS.png)


### When the failing func call is a few lines away

Even when the func call is a few lines away, there is no problem for finding it.

#### Output

![Source Example: error earlier in the code](https://i.imgur.com/wPBrYqs.png)

## Documentation

Documentation can be found here : [![Documentation](https://godoc.org/github.com/snwfdhmp/errlog?status.svg)](http://godoc.org/github.com/snwfdhmp/errlog)

## Feedbacks

Feel free to open an issue for any feedback or suggestion.

I fix process issues quickly.

## Contributions

We are happy to collaborate with you :

- Ask for a new feature: [Open an issue](https://github.com/snwfdhmp/errlog/issues/new)
- Add your feature: [Open a PR](https://github.com/snwfdhmp/errlog/compare)

When submitting a PR, please apply Effective Go best practices. For more information: https://golang.org/doc/effective_go.html

## License information

Click the following badge to open LICENSE information.

[![license](https://img.shields.io/github/license/snwfdhmp/errlog.svg?maxAge=60000)](https://github.com/snwfdhmp/errlog/LICENSE)

## Contributors

### Major

- [snwfdhmp](https://github.com/snwfdhmp): Author and maintainer
- [chemidy](https://github.com/chemidy): Added important badges

### Minor fixes

- [orisano](https://github.com/orisano)
- [programmingman](https://github.com/programmingman)
