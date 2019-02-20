package errlog

import "fmt"

var (
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
)
