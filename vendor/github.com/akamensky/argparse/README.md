# Golang argparse

[![GoDoc](https://godoc.org/github.com/akamensky/argparse?status.svg)](https://godoc.org/github.com/akamensky/argparse) [![Go Report Card](https://goreportcard.com/badge/github.com/akamensky/argparse)](https://goreportcard.com/report/github.com/akamensky/argparse) [![Coverage Status](https://coveralls.io/repos/github/akamensky/argparse/badge.svg?branch=master)](https://coveralls.io/github/akamensky/argparse?branch=master) [![Build Status](https://travis-ci.org/akamensky/argparse.svg?branch=master)](https://travis-ci.org/akamensky/argparse)

Let's be honest -- Go's standard command line arguments parser `flag` terribly sucks. 
It cannot come anywhere close to the Python's `argparse` module. This is why this project exists.

The goal of this project is to bring ease of use and flexibility of `argparse` to Go. 
Which is where the name of this package comes from.

#### Installation

To install and start using argparse simply do:

```
$ go get -u -v github.com/akamensky/argparse
```

You are good to go to write your first command line tool!
See Usage and Examples sections for information how you can use it

#### Usage

To start using argparse in Go see above instructions on how to install.
From here on you can start writing your first program.
Please check out examples from `examples/` directory to see how to use it in various ways.

Here is basic example of print command (from `examples/print/` directory):
```go
package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"os"
)

func main() {
	// Create new parser object
	parser := argparse.NewParser("print", "Prints provided string to stdout")
	// Create string flag
	s := parser.String("s", "string", &argparse.Options{Required: true, Help: "String to print"})
	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
	}
	// Finally print the collected string
	fmt.Println(*s)
}
```

#### Basic options

Create your parser instance and pass it program name and program description.
Program name if empty will be taken from `os.Args[0]` (which is okay in most cases).
Description can be as long as you wish and will be used in `--help` output
```go
parser := argparse.NewParser("progname", "Description of my awesome program. It can be as long as I wish it to be")
```

String will allow you to get a string from arguments, such as `$ progname --string "String content"`
```go
var myString *string = parser.String("s", "string", ...)
```

Selector works same as a string, except that it will only allow specific values.
For example like this `$ progname --debug-level WARN`
```go
var mySelector *string = parser.Selector("d", "debug-level", []string{"INFO", "DEBUG", "WARN"}, ...)
```

StringList allows to collect multiple string values into the slice of strings by repeating same flag multiple times.
Such as `$ progname --string hostname1 --string hostname2 -s hostname3`
```go
var myStringList *[]string = parser.StringList("s", "string", ...)
```

List allows to collect multiple values into the slice of strings by repeating same flag multiple times 
(at fact - it is an Alias of StringList).
Such as `$ progname --host hostname1 --host hostname2 -H hostname3`
```go
var myList *[]string = parser.List("H", "hostname", ...)
```

Flag will tell you if a simple flag was set on command line (true is set, false is not).
For example `$ progname --force`
```go
var myFlag *bool = parser.Flag("f", "force", ...)
```

FlagCounter will tell you the number of times that  simple flag  was set on command line 
(integer greater than or equal to 1 or 0 if not set).
For example `$ progname -vv --verbose`
```go
var myFlagCounter *int = parser.FlagCounter("v", "verbose", ...)
```

Int will allow you to get a decimal integer from arguments, such as `$ progname --integer "42"`
```go
var myInteger *int = parser.Int("i", "integer", ...)
```

IntList allows to collect multiple decimal integer values into the slice of integers by repeating same flag multiple times.
Such as `$ progname --integer 42 --integer +51 -i -1`
```go
var myIntegerList *[]int = parser.IntList("i", "integer", ...)
```

Float will allow you to get a floating point number from arguments, such as `$ progname --float "37.2"`
```go
var myFloat *float64 = parser.Float("f", "float", ...)
```

FloatList allows to collect multiple floating point number values into the slice of floats by repeating same flag multiple times.
Such as `$ progname --float 42 --float +37.2 -f -1.0`
```go
var myFloatList *[]float64 = parser.FloatList("f", "float", ...)
```

File will validate that file exists and will attempt to open it with provided privileges.
To be used like this `$ progname --log-file /path/to/file.log`
```go
var myLogFile *os.File = parser.File("l", "log-file", os.O_RDWR, 0600, ...)
```

FileList allows to collect files into the slice of files by repeating same flag multiple times.
FileList will validate that files exists and will attempt to open them with provided privileges.
To be used like this `$ progname --log-file /path/to/file.log --log-file /path/to/file_cpy.log -l /another/path/to/file.log`
```go
var myLogFiles *[]os.File = parser.FileList("l", "log-file", os.O_RDWR, 0600, ...)
```

You can implement sub-commands in your CLI using `parser.NewCommand()` or go even deeper with `command.NewCommand()`.
Since parser inherits from command, every command supports exactly same options as parser itself,
thus allowing to add arguments specific to that command or more global arguments added on parser itself!

You can also dynamically retrieve argument values:
```
var myInteger *int = parser.Int("i", "integer", ...)
parser.Parse()
fmt.Printf("%d", *parser.GetArgs()[0].GetResult().(*int))
```

#### Basic Option Structure

The `Option` structure is declared at `argparse.go`:
```go
type Options struct {
	Required bool
	Validate func(args []string) error
	Help     string
	Default  interface{}
}
```

You can Set `Required` to let it know if it should ask for arguments.
Or you can set `Validate` as a lambda function to make it know while value is valid.
Or you can set `Help` for your beautiful help document.
Or you can set `Default` will set the default value if user does not provide a value.

Example:
```
dirpath := parser.String("d", "dirpath",
			 &argparse.Options{
			 	Require: false,
				Help: "the input files' folder path",
				Default: "input",
			})
```

#### Caveats

There are a few caveats (or more like design choices) to know about:
* Shorthand arguments MUST be a single character. Shorthand arguments are prepended with single dash `"-"`
* If not convenient shorthand argument can be completely skipped by passing empty string `""` as first argument
* Shorthand arguments ONLY for `parser.Flag()` and  `parser.FlagCounter()` can be combined into single argument same as `ps -aux`, `rm -rf` or `lspci -vvk` 
* Long arguments must be specified and cannot be empty. They are prepended with double dash `"--"`
* You cannot define two same arguments. Only first one will be used. For example doing `parser.Flag("t", "test", nil)` followed by `parser.String("t", "test2", nil)` will not work as second `String` argument will be ignored (note that both have `"t"` as shorthand argument). However since it is case-sensitive library, you can work arounf it by capitalizing one of the arguments
* There is a pre-defined argument for `-h|--help`, so from above attempting to define any argument using `h` as shorthand will fail
* `parser.Parse()` returns error in case of something going wrong, but it is not expected to cover ALL cases
* Any arguments that left un-parsed will be regarded as error


#### Contributing

Can you write in Go? Then this projects needs your help!

Take a look at open issues, specially the ones tagged as `help-wanted`.
If you have any improvements to offer, please open an issue first to ensure this improvement is discussed.

There are following tasks to be done:
* Add more examples
* Improve code quality (it is messy right now and could use a major revamp to improve gocyclo report)
* Add more argument options (such as numbers parsing)
* Improve test coverage
* Write a wiki for this project

However note that the logic outlined in method comments must be preserved 
as the the library must stick with backward compatibility promise!

#### Acknowledgments

Thanks to Python developers for making a great `argparse` which inspired this package to match for greatness of Go
