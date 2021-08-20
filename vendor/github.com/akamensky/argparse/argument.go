package argparse

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type arg struct {
	result   interface{} // Pointer to the resulting value
	opts     *Options    // Options
	sname    string      // Short name (in Parser will start with "-"
	lname    string      // Long name (in Parser will start with "--"
	size     int         // Size defines how many args after match will need to be consumed
	unique   bool        // Specifies whether flag should be present only ones
	parsed   bool        // Specifies whether flag has been parsed already
	fileFlag int         // File mode to open file with
	filePerm os.FileMode // File permissions to set a file
	selector *[]string   // Used in Selector type to allow to choose only one from list of options
	parent   *Command    // Used to get access to specific Command
	eqChar   bool        // This is used if the command is passed in with an equals char as a seperator
}

// Arg interface provides exporting of arg structure, while exposing it
type Arg interface {
	GetOpts() *Options
	GetSname() string
	GetLname() string
	GetResult() interface{}
}

func (o arg) GetOpts() *Options {
	return o.opts
}

func (o arg) GetSname() string {
	return o.sname
}

func (o arg) GetLname() string {
	return o.lname
}

// getResult returns the interface{} to the *(type) containing the argument's result value
// Will contain the empty/default value if argument value was not given
func (o arg) GetResult() interface{} {
	return o.result
}

type help struct{}

// checkLongName if long argumet present.
// checkLongName - returns the argumet's long name number of occurrences and error.
// For long name return value is 0 or 1.
func (o *arg) checkLongName(argument string) int {
	// Check for long name only if not empty
	if o.lname != "" {
		// If argument begins with "--" and next is not "-" then it is a long name
		if len(argument) > 2 && strings.HasPrefix(argument, "--") && argument[2] != '-' {
			if argument[2:] == o.lname {
				return 1
			}
		}
	}

	return 0
}

// checkShortName if argumet present.
// checkShortName - returns the argumet's short name number of occurrences and error.
// For shorthand argument - 0 if there is no occurrences, or count of occurrences.
// Shorthand argument with parametr, mast be the only or last in the argument string.
func (o *arg) checkShortName(argument string) (int, error) {
	// Check for short name only if not empty
	if o.sname != "" {

		// If argument begins with "-" and next is not "-" then it is a short name
		if len(argument) > 1 && strings.HasPrefix(argument, "-") && argument[1] != '-' {
			count := strings.Count(argument[1:], o.sname)
			switch {
			// For args with size 1 (Flag,FlagCounter) multiple shorthand in one argument are allowed
			case o.size == 1:
				return count, nil
			// For args with o.size > 1, shorthand argument is allowed only to complete the sequence of arguments combined into one
			case o.size > 1:
				if count > 1 {
					return count, fmt.Errorf("[%s] argument: The parameter must follow", o.name())
				}
				if strings.HasSuffix(argument[1:], o.sname) {
					return count, nil
				}
			//if o.size < 1 - it is an error
			default:
				return 0, fmt.Errorf("Argument's size < 1 is not allowed")
			}
		}
	}

	return 0, nil
}

// check if argumet present.
// check - returns the argumet's number of occurrences and error.
// For long name return value is 0 or 1.
// For shorthand argument - 0 if there is no occurrences, or count of occurrences.
// Shorthand argument with parametr, mast be the only or last in the argument string.
func (o *arg) check(argument string) (int, error) {
	rez := o.checkLongName(argument)
	if rez > 0 {
		return rez, nil
	}

	return o.checkShortName(argument)
}

func (o *arg) reduceLongName(position int, args *[]string) {
	argument := (*args)[position]
	// Check for long name only if not empty
	if o.lname != "" {
		// If argument begins with "--" and next is not "-" then it is a long name
		if len(argument) > 2 && strings.HasPrefix(argument, "--") && argument[2] != '-' {
			if o.eqChar {
				splitInd := strings.LastIndex(argument, "=")
				equalArg := []string{argument[:splitInd], argument[splitInd+1:]}
				argument = equalArg[0]
			}
			if argument[2:] == o.lname {
				for i := position; i < position+o.size; i++ {
					(*args)[i] = ""
				}
			}
		}
	}
}

func (o *arg) reduceShortName(position int, args *[]string) {
	argument := (*args)[position]
	// Check for short name only if not empty
	if o.sname != "" {
		// If argument begins with "-" and next is not "-" then it is a short name
		if len(argument) > 1 && strings.HasPrefix(argument, "-") && argument[1] != '-' {
			// For args with size 1 (Flag,FlagCounter) we allow multiple shorthand in one
			if o.size == 1 {
				if strings.Contains(argument[1:], o.sname) {
					(*args)[position] = strings.Replace(argument, o.sname, "", -1)
					if (*args)[position] == "-" {
						(*args)[position] = ""
					}
					if o.eqChar {
						(*args)[position] = ""
					}
				}
				// For all other types it must be separate argument
			} else {
				if argument[1:] == o.sname {
					for i := position; i < position+o.size; i++ {
						(*args)[i] = ""
					}
				}
			}
		}
	}
}

// clear out already used argument from args at position
func (o *arg) reduce(position int, args *[]string) {
	o.reduceLongName(position, args)
	o.reduceShortName(position, args)
}

func (o *arg) parseInt(args []string, argCount int) error {
	//data of integer type is for
	switch {
	//FlagCounter argument
	case len(args) < 1:
		if o.size > 1 {
			return fmt.Errorf("[%s] must be followed by an integer", o.name())
		}
		*o.result.(*int) += argCount
	case len(args) > 1:
		return fmt.Errorf("[%s] followed by too many arguments", o.name())
		//or Int argument with one integer parameter
	default:
		val, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("[%s] bad integer value [%s]", o.name(), args[0])
		}
		*o.result.(*int) = val
	}
	o.parsed = true
	return nil
}

func (o *arg) parseBool(args []string) error {
	//data of bool type is for Flag argument
	*o.result.(*bool) = true
	o.parsed = true
	return nil
}

func (o *arg) parseFloat(args []string) error {
	//data of float64 type is for Float argument with one float parameter
	if len(args) < 1 {
		return fmt.Errorf("[%s] must be followed by a floating point number", o.name())
	}
	if len(args) > 1 {
		return fmt.Errorf("[%s] followed by too many arguments", o.name())
	}

	val, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("[%s] bad floating point value [%s]", o.name(), args[0])
	}

	*o.result.(*float64) = val
	o.parsed = true
	return nil
}

func (o *arg) parseString(args []string) error {
	//data of string type is for String argument with one string parameter
	if len(args) < 1 {
		return fmt.Errorf("[%s] must be followed by a string", o.name())
	}
	if len(args) > 1 {
		return fmt.Errorf("[%s] followed by too many arguments", o.name())
	}

	// Selector case
	if o.selector != nil {
		match := false
		for _, v := range *o.selector {
			if args[0] == v {
				match = true
			}
		}
		if !match {
			return fmt.Errorf("bad value for [%s]. Allowed values are %v", o.name(), *o.selector)
		}
	}

	*o.result.(*string) = args[0]
	o.parsed = true
	return nil
}

func (o *arg) parseFile(args []string) error {
	//data of os.File type is for File argument with one file name parameter
	if len(args) < 1 {
		return fmt.Errorf("[%s] must be followed by a path to file", o.name())
	}
	if len(args) > 1 {
		return fmt.Errorf("[%s] followed by too many arguments", o.name())
	}

	f, err := os.OpenFile(args[0], o.fileFlag, o.filePerm)
	if err != nil {
		return err
	}

	*o.result.(*os.File) = *f
	o.parsed = true
	return nil
}

func (o *arg) parseStringList(args []string) error {
	//data of []string type is for List and StringList argument with set of string parameters
	if len(args) < 1 {
		return fmt.Errorf("[%s] must be followed by a string", o.name())
	}
	if len(args) > 1 {
		return fmt.Errorf("[%s] followed by too many arguments", o.name())
	}

	*o.result.(*[]string) = append(*o.result.(*[]string), args[0])
	o.parsed = true
	return nil
}

func (o *arg) parseIntList(args []string) error {
	//data of []int type is for IntList argument with set of int parameters
	switch {
	case len(args) < 1:
		return fmt.Errorf("[%s] must be followed by an integer", o.name())
	case len(args) > 1:
		return fmt.Errorf("[%s] followed by too many arguments", o.name())
	}

	val, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("[%s] bad integer value [%s]", o.name(), args[0])
	}
	*o.result.(*[]int) = append(*o.result.(*[]int), val)
	o.parsed = true
	return nil
}

func (o *arg) parseFloatList(args []string) error {
	//data of []float64 type is for FloatList argument with set of int parameters
	switch {
	case len(args) < 1:
		return fmt.Errorf("[%s] must be followed by a floating point number", o.name())
	case len(args) > 1:
		return fmt.Errorf("[%s] followed by too many arguments", o.name())
	}

	val, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("[%s] bad floating point value [%s]", o.name(), args[0])
	}
	*o.result.(*[]float64) = append(*o.result.(*[]float64), val)
	o.parsed = true
	return nil
}

func (o *arg) parseFileList(args []string) error {
	//data of []os.File type is for FileList argument with set of int parameters
	switch {
	case len(args) < 1:
		return fmt.Errorf("[%s] must be followed by a path to file", o.name())
	case len(args) > 1:
		return fmt.Errorf("[%s] followed by too many arguments", o.name())
	}
	f, err := os.OpenFile(args[0], o.fileFlag, o.filePerm)
	if err != nil {
		//if one of FileList's file opening have been failed, close all other in this list
		errs := make([]string, 0, len(*o.result.(*[]os.File)))
		for _, f := range *o.result.(*[]os.File) {
			if err := f.Close(); err != nil {
				//almost unreal, but what if another process closed this file
				errs = append(errs, err.Error())
			}
		}
		if len(errs) > 0 {
			err = fmt.Errorf("while handling error: %v, other errors occured: %#v", err.Error(), errs)
		}
		*o.result.(*[]os.File) = []os.File{}
		return err
	}
	*o.result.(*[]os.File) = append(*o.result.(*[]os.File), *f)
	o.parsed = true
	return nil
}

// To overwrite while testing
// Possibly extend to allow user overriding
var exit func(int) = os.Exit
var print func(...interface{}) (int, error) = fmt.Println

func (o *arg) parseSomeType(args []string, argCount int) error {
	var err error
	switch o.result.(type) {
	case *help:
		print(o.parent.Help(nil))
		if o.parent.exitOnHelp {
			exit(0)
		}
	//data of bool type is for Flag argument
	case *bool:
		err = o.parseBool(args)
	case *int:
		err = o.parseInt(args, argCount)
	case *float64:
		err = o.parseFloat(args)
	case *string:
		err = o.parseString(args)
	case *os.File:
		err = o.parseFile(args)
	case *[]string:
		err = o.parseStringList(args)
	case *[]int:
		err = o.parseIntList(args)
	case *[]float64:
		err = o.parseFloatList(args)
	case *[]os.File:
		err = o.parseFileList(args)
	default:
		err = fmt.Errorf("unsupported type [%t]", o.result)
	}
	return err
}

func (o *arg) parse(args []string, argCount int) error {
	// If unique do not allow more than one time
	if o.unique && (o.parsed || argCount > 1) {
		return fmt.Errorf("[%s] can only be present once", o.name())
	}

	// If validation function provided -- execute, on error return it immediately
	if o.opts != nil && o.opts.Validate != nil {
		err := o.opts.Validate(args)
		if err != nil {
			return err
		}
	}
	return o.parseSomeType(args, argCount)
}

func (o *arg) name() string {
	var name string
	if o.lname == "" {
		name = "-" + o.sname
	} else if o.sname == "" {
		name = "--" + o.lname
	} else {
		name = "-" + o.sname + "|" + "--" + o.lname
	}
	return name
}

func (o *arg) usage() string {
	var result string
	result = o.name()
	switch o.result.(type) {
	case *bool:
		break
	case *int:
		result = result + " <integer>"
	case *float64:
		result = result + " <float>"
	case *string:
		if o.selector != nil {
			result = result + " (" + strings.Join(*o.selector, "|") + ")"
		} else {
			result = result + " \"<value>\""
		}
	case *os.File:
		result = result + " <file>"
	case *[]string:
		result = result + " \"<value>\"" + " [" + result + " \"<value>\" ...]"
	default:
		break
	}
	if o.opts == nil || o.opts.Required == false {
		result = "[" + result + "]"
	}
	return result
}

func (o *arg) getHelpMessage() string {
	message := ""
	if len(o.opts.Help) > 0 {
		message += o.opts.Help
		if !o.opts.Required && o.opts.Default != nil {
			message += fmt.Sprintf(". Default: %v", o.opts.Default)
		}
	}
	return message
}

// setDefaultFile - gets default os.File object based on provided default filename string
func (o *arg) setDefaultFile() error {
	// In case of File we should get string as default value
	if v, ok := o.opts.Default.(string); ok {
		f, err := os.OpenFile(v, o.fileFlag, o.filePerm)
		if err != nil {
			return err
		}
		*o.result.(*os.File) = *f
	} else {
		return fmt.Errorf("cannot use default type [%T] as value of pointer with type [*string]", o.opts.Default)
	}
	return nil
}

// setDefaultFiles - gets list of default os.File objects based on provided list of default filenames strings
func (o *arg) setDefaultFiles() error {
	// In case of FileList we should get []string as default value
	var files []os.File
	if fileNames, ok := o.opts.Default.([]string); ok {
		files = make([]os.File, 0, len(fileNames))
		for _, v := range fileNames {
			f, err := os.OpenFile(v, o.fileFlag, o.filePerm)
			if err != nil {
				//if one of FileList's file opening have been failed, close all other in this list
				errs := make([]string, 0, len(*o.result.(*[]os.File)))
				for _, f := range *o.result.(*[]os.File) {
					if err := f.Close(); err != nil {
						//almost unreal, but what if another process closed this file
						errs = append(errs, err.Error())
					}
				}
				if len(errs) > 0 {
					err = fmt.Errorf("while handling error: %v, other errors occured: %#v", err.Error(), errs)
				}
				*o.result.(*[]os.File) = []os.File{}
				return err
			}
			files = append(files, *f)
		}
	} else {
		return fmt.Errorf("cannot use default type [%T] as value of pointer with type [*[]string]", o.opts.Default)
	}
	*o.result.(*[]os.File) = files
	return nil
}

// setDefault - if no value getted for specific argument, set default value, if provided
func (o *arg) setDefault() error {
	// Only set default if it was not parsed, and default value was defined
	if !o.parsed && o.opts != nil && o.opts.Default != nil {
		switch o.result.(type) {
		case *bool, *int, *float64, *string, *[]bool, *[]int, *[]float64, *[]string:
			if reflect.TypeOf(o.result) != reflect.PtrTo(reflect.TypeOf(o.opts.Default)) {
				return fmt.Errorf("cannot use default type [%T] as value of pointer with type [%T]", o.opts.Default, o.result)
			}
			reflect.ValueOf(o.result).Elem().Set(reflect.ValueOf(o.opts.Default))

		case *os.File:
			if err := o.setDefaultFile(); err != nil {
				return err
			}
		case *[]os.File:
			if err := o.setDefaultFiles(); err != nil {
				return err
			}
		}
	}

	return nil
}
