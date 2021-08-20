// Package argparse provides users with more flexible and configurable option for command line arguments parsing.
package argparse

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// DisableDescription can be assigned as a command or arguments description to hide it from the Usage output
const DisableDescription = "DISABLEDDESCRIPTIONWILLNOTSHOWUP"

//disable help can be invoked from the parse and then needs to be propogated to subcommands
var disableHelp = false

// Command is a basic type for this package. It represents top level Parser as well as any commands and sub-commands
// Command MUST NOT ever be created manually. Instead one should call NewCommand method of Parser or Command,
// which will setup appropriate fields and call methods that have to be called when creating new command.
type Command struct {
	name        string
	description string
	args        []*arg
	commands    []*Command
	parsed      bool
	happened    bool
	parent      *Command
	HelpFunc    func(c *Command, msg interface{}) string
	exitOnHelp  bool
}

// GetName exposes Command's name field
func (o Command) GetName() string {
	return o.name
}

// GetDescription exposes Command's description field
func (o Command) GetDescription() string {
	return o.description
}

// GetArgs exposes Command's args field
func (o Command) GetArgs() (args []Arg) {
	for _, arg := range o.args {
		args = append(args, arg)
	}
	return
}

// GetCommands exposes Command's commands field
func (o Command) GetCommands() []*Command {
	return o.commands
}

// GetParent exposes Command's parent field
func (o Command) GetParent() *Command {
	return o.parent
}

// Help calls the overriddable Command.HelpFunc on itself,
// called when the help argument strings are passed via CLI
func (o *Command) Help(msg interface{}) string {
	tempC := o
	for tempC.HelpFunc == nil {
		if tempC.parent == nil {
			return ""
		}
		tempC = tempC.parent
	}
	return tempC.HelpFunc(o, msg)
}

// Parser is a top level object of argparse. It MUST NOT ever be created manually. Instead one should use
// argparse.NewParser() method that will create new parser, propagate necessary private fields and call needed
// functions.
type Parser struct {
	Command
}

// Options are specific options for every argument. They can be provided if necessary.
// Possible fields are:
//
// Options.Required - tells Parser that this argument is required to be provided.
// useful when specific Command requires some data provided.
//
// Options.Validate - is a validation function. Using this field anyone can implement a custom validation for argument.
// If provided and argument is present, then function is called. If argument also consumes any following values
// (e.g. as String does), then these are provided as args to function. If validation fails the error must be returned,
// which will be the output of `Parser.Parse` method.
//
// Options.Help - A help message to be displayed in Usage output. Can be of any length as the message will be
// formatted to fit max screen width of 100 characters.
//
// Options.Default - A default value for an argument. This value will be assigned to the argument at the end of parsing
// in case if this argument was not supplied on command line. File default value is a string which it will be open with
// provided options. In case if provided value type does not match expected, the error will be returned on run-time.
type Options struct {
	Required bool
	Validate func(args []string) error
	Help     string
	Default  interface{}
}

// NewParser creates new Parser object that will allow to add arguments for parsing
// It takes program name and description which will be used as part of Usage output
// Returns pointer to Parser object
func NewParser(name string, description string) *Parser {
	p := &Parser{}

	p.name = name
	p.description = description

	p.args = make([]*arg, 0)
	p.commands = make([]*Command, 0)

	p.help("h", "help")
	p.exitOnHelp = true
	p.HelpFunc = (*Command).Usage

	return p
}

// NewCommand will create a sub-command and propagate all necessary fields.
// All commands are always at the beginning of the arguments.
// Parser can have commands and those commands can have sub-commands,
// which allows for very flexible workflow.
// All commands are considered as required and all commands can have their own argument set.
// Commands are processed Parser -> Command -> sub-Command.
// Arguments will be processed in order of sub-Command -> Command -> Parser.
func (o *Command) NewCommand(name string, description string) *Command {
	c := new(Command)
	c.name = name
	c.description = description
	c.parsed = false
	c.parent = o
	if !disableHelp {
		c.help("h", "help")
		c.exitOnHelp = true
		c.HelpFunc = (*Command).Usage
	}

	if o.commands == nil {
		o.commands = make([]*Command, 0)
	}

	o.commands = append(o.commands, c)

	return c
}

// DisableHelp removes any help arguments from the commands list of arguments
// This prevents prevents help from being parsed or invoked from the argument list
func (o *Parser) DisableHelp() {
	disableHelp = true
	for i, arg := range o.args {
		if _, ok := arg.result.(*help); ok {
			o.args = append(o.args[:i], o.args[i+1:]...)
		}
	}
	for _, com := range o.commands {
		for i, comArg := range com.args {
			if _, ok := comArg.result.(*help); ok {
				com.args = append(com.args[:i], com.args[i+1:]...)
			}
		}
	}
}

// ExitOnHelp sets the exitOnHelp variable of Parser
func (o *Command) ExitOnHelp(b bool) {
	o.exitOnHelp = b
	for _, c := range o.commands {
		c.ExitOnHelp(b)
	}
}

// SetHelp removes the previous help argument, and creates a new one with the desired sname/lname
func (o *Parser) SetHelp(sname, lname string) {
	o.DisableHelp()
	o.help(sname, lname)
}

// Flag Creates new flag type of argument, which is boolean value showing if argument was provided or not.
// Takes short name, long name and pointer to options (optional).
// Short name must be single character, but can be omitted by giving empty string.
// Long name is required.
// Returns pointer to boolean with starting value `false`. If Parser finds the flag
// provided on Command line arguments, then the value is changed to true.
// Set of Flag and FlagCounter shorthand arguments can be combined together such as `tar -cvaf foo.tar foo`
func (o *Command) Flag(short string, long string, opts *Options) *bool {
	var result bool

	a := &arg{
		result: &result,
		sname:  short,
		lname:  long,
		size:   1,
		opts:   opts,
		unique: true,
	}

	if err := o.addArg(a); err != nil {
		panic(fmt.Errorf("unable to add Flag: %s", err.Error()))
	}

	return &result
}

// FlagCounter Creates new flagCounter type of argument, which is integer value showing the number of times the argument has been provided.
// Takes short name, long name and pointer to options (optional).
// Short name must be single character, but can be omitted by giving empty string.
// Long name is required.
// Returns pointer to integer with starting value `0`. Each time Parser finds the flag
// provided on Command line arguments, the value is incremented by 1.
// Set of FlagCounter and Flag shorthand arguments can be combined together such as `tar -cvaf foo.tar foo`
func (o *Command) FlagCounter(short string, long string, opts *Options) *int {
	var result int

	a := &arg{
		result: &result,
		sname:  short,
		lname:  long,
		size:   1,
		opts:   opts,
		unique: false,
	}

	if err := o.addArg(a); err != nil {
		panic(fmt.Errorf("unable to add FlagCounter: %s", err.Error()))
	}

	return &result
}

// String creates new string argument, which will return whatever follows the argument on CLI.
// Takes as arguments short name (must be single character or an empty string)
// long name and (optional) options
func (o *Command) String(short string, long string, opts *Options) *string {
	var result string

	a := &arg{
		result: &result,
		sname:  short,
		lname:  long,
		size:   2,
		opts:   opts,
		unique: true,
	}

	if err := o.addArg(a); err != nil {
		panic(fmt.Errorf("unable to add String: %s", err.Error()))
	}

	return &result
}

// Int creates new int argument, which will attempt to parse following argument as int.
// Takes as arguments short name (must be single character or an empty string)
// long name and (optional) options.
// If parsing fails parser.Parse() will return an error.
func (o *Command) Int(short string, long string, opts *Options) *int {
	var result int

	a := &arg{
		result: &result,
		sname:  short,
		lname:  long,
		size:   2,
		opts:   opts,
		unique: true,
	}

	if err := o.addArg(a); err != nil {
		panic(fmt.Errorf("unable to add Int: %s", err.Error()))
	}

	return &result
}

// Float creates new float argument, which will attempt to parse following argument as float64.
// Takes as arguments short name (must be single character or an empty string)
// long name and (optional) options.
// If parsing fails parser.Parse() will return an error.
func (o *Command) Float(short string, long string, opts *Options) *float64 {
	var result float64

	a := &arg{
		result: &result,
		sname:  short,
		lname:  long,
		size:   2,
		opts:   opts,
		unique: true,
	}

	if err := o.addArg(a); err != nil {
		panic(fmt.Errorf("unable to add Float: %s", err.Error()))
	}

	return &result
}

// File creates new file argument, which is when provided will check if file exists or attempt to create it
// depending on provided flags (same as for os.OpenFile).
// It takes same as all other arguments short and long names, additionally it takes flags that specify
// in which mode the file should be open (see os.OpenFile for details on that), file permissions that
// will be applied to a file and argument options.
// Returns a pointer to os.File which will be set to opened file on success. On error the Parser.Parse
// will return error and the pointer might be nil.
func (o *Command) File(short string, long string, flag int, perm os.FileMode, opts *Options) *os.File {
	var result os.File

	a := &arg{
		result:   &result,
		sname:    short,
		lname:    long,
		size:     2,
		opts:     opts,
		unique:   true,
		fileFlag: flag,
		filePerm: perm,
	}

	if err := o.addArg(a); err != nil {
		panic(fmt.Errorf("unable to add File: %s", err.Error()))
	}

	return &result
}

// List creates new list argument. This is the argument that is allowed to be present multiple times on CLI.
// All appearances of this argument on CLI will be collected into the list of default type values which is strings.
// If no argument provided, then the list is empty.
// Takes same parameters as String.
// Returns a pointer the list of strings.
func (o *Command) List(short string, long string, opts *Options) *[]string {
	return o.StringList(short, long, opts)
}

// StringList creates new string list argument. This is the argument that is allowed to be present multiple times on CLI.
// All appearances of this argument on CLI will be collected into the list of strings. If no argument
// provided, then the list is empty. Takes same parameters as String
// Returns a pointer the list of strings.
func (o *Command) StringList(short string, long string, opts *Options) *[]string {
	result := make([]string, 0)

	a := &arg{
		result: &result,
		sname:  short,
		lname:  long,
		size:   2,
		opts:   opts,
		unique: false,
	}

	if err := o.addArg(a); err != nil {
		panic(fmt.Errorf("unable to add StringList: %s", err.Error()))
	}

	return &result
}

// IntList creates new integer list argument. This is the argument that is allowed to be present multiple times on CLI.
// All appearances of this argument on CLI will be collected into the list of integers. If no argument
// provided, then the list is empty. Takes same parameters as Int
// Returns a pointer the list of integers.
func (o *Command) IntList(short string, long string, opts *Options) *[]int {
	result := make([]int, 0)

	a := &arg{
		result: &result,
		sname:  short,
		lname:  long,
		size:   2,
		opts:   opts,
		unique: false,
	}

	if err := o.addArg(a); err != nil {
		panic(fmt.Errorf("unable to add IntList: %s", err.Error()))
	}

	return &result
}

// FloatList creates new float list argument. This is the argument that is allowed to be present multiple times on CLI.
// All appearances of this argument on CLI will be collected into the list of float64 values. If no argument
// provided, then the list is empty. Takes same parameters as Float
// Returns a pointer the list of float64 values.
func (o *Command) FloatList(short string, long string, opts *Options) *[]float64 {
	result := make([]float64, 0)

	a := &arg{
		result: &result,
		sname:  short,
		lname:  long,
		size:   2,
		opts:   opts,
		unique: false,
	}

	if err := o.addArg(a); err != nil {
		panic(fmt.Errorf("unable to add FloatList: %s", err.Error()))
	}

	return &result
}

// FileList creates new file list argument. This is the argument that is allowed to be present multiple times on CLI.
// All appearances of this argument on CLI will be collected into the list of os.File values. If no argument
// provided, then the list is empty. Takes same parameters as File
// Returns a pointer the list of os.File values.
func (o *Command) FileList(short string, long string, flag int, perm os.FileMode, opts *Options) *[]os.File {
	result := make([]os.File, 0)

	a := &arg{
		result:   &result,
		sname:    short,
		lname:    long,
		size:     2,
		opts:     opts,
		unique:   false,
		fileFlag: flag,
		filePerm: perm,
	}

	if err := o.addArg(a); err != nil {
		panic(fmt.Errorf("unable to add FileList: %s", err.Error()))
	}

	return &result
}

// Selector creates a selector argument. Selector argument works in the same way as String argument, with
// the difference that the string value must be from the list of options provided by the program.
// Takes short and long names, argument options and a slice of strings which are allowed values
// for CLI argument.
// Returns a pointer to a string. If argument is not required (as in argparse.Options.Required),
// and argument was not provided, then the string is empty.
func (o *Command) Selector(short string, long string, options []string, opts *Options) *string {
	var result string

	a := &arg{
		result:   &result,
		sname:    short,
		lname:    long,
		size:     2,
		opts:     opts,
		unique:   true,
		selector: &options,
	}

	if err := o.addArg(a); err != nil {
		panic(fmt.Errorf("unable to add Selector: %s", err.Error()))
	}

	return &result
}

// message2String puts msg in result string
// done boolean indicates if result is ready to be returned
// Accepts an interface that can be error, string or fmt.Stringer that will be prepended to a message.
// All other interface types will be ignored
func message2String(msg interface{}) (string, bool) {
	var result string
	if msg != nil {
		switch msg.(type) {
		case subCommandError:
			result = fmt.Sprintf("%s\n", msg.(error).Error())
			if msg.(subCommandError).cmd != nil {
				result += msg.(subCommandError).cmd.Usage(nil)
			}
			return result, true
		case error:
			result = fmt.Sprintf("%s\n", msg.(error).Error())
		case string:
			result = fmt.Sprintf("%s\n", msg.(string))
		case fmt.Stringer:
			result = fmt.Sprintf("%s\n", msg.(fmt.Stringer).String())
		}
	}
	return result, false
}

// getPrecedingCommands - collects info on command chain from root to current (o *Command) and all arguments in this chain
func (o *Command) getPrecedingCommands(chain *[]string, arguments *[]*arg) {
	current := o
	// Also add arguments
	// Get line of commands until root
	for current != nil {
		*chain = append(*chain, current.name)
		if current.args != nil {
			*arguments = append(*arguments, current.args...)
		}
		current = current.parent
	}

	// Reverse the slice
	last := len(*chain) - 1
	for i := 0; i < len(*chain)/2; i++ {
		(*chain)[i], (*chain)[last-i] = (*chain)[last-i], (*chain)[i]
	}
}

// getSubCommands - collects info on subcommands of current command
func (o *Command) getSubCommands(chain *[]string) []Command {
	commands := make([]Command, 0)
	if o.commands != nil && len(o.commands) > 0 {
		*chain = append(*chain, "<Command>")
		for _, v := range o.commands {
			// Skip hidden commands
			if v.description == DisableDescription {
				continue
			}
			commands = append(commands, *v)
		}
	}
	return commands
}

// precedingCommands2Result - puts info about command chain from root to current (o *Command) into result string buffer
func (o *Command) precedingCommands2Result(result string, chain []string, arguments []*arg, maxWidth int) string {
	usedHelp := false
	leftPadding := len("usage: " + chain[0] + "")
	// Add preceding commands
	for _, v := range chain {
		result = addToLastLine(result, v, maxWidth, leftPadding, true)
	}
	// Add arguments from this and all preceding commands
	for _, v := range arguments {
		// Skip arguments that are hidden
		if v.opts.Help == DisableDescription {
			continue
		}
		if v.lname == "help" && usedHelp {
		} else {
			result = addToLastLine(result, v.usage(), maxWidth, leftPadding, true)
		}
		if v.lname == "help" || v.sname == "h" {
			usedHelp = true
		}
	}
	// Add program/Command description to the result
	result = result + "\n\n" + strings.Repeat(" ", leftPadding)
	result = addToLastLine(result, o.description, maxWidth, leftPadding, true)
	result = result + "\n\n"

	return result
}

// subCommands2Result - puts info about subcommands of current command into result string buffer
func subCommands2Result(result string, commands []Command, maxWidth int) string {
	// Add list of sub-commands to the result
	if len(commands) > 0 {
		cmdContent := "Commands:\n\n"
		// Get biggest padding
		var cmdPadding int
		for _, com := range commands {
			if com.description == DisableDescription {
				continue
			}
			if len("  "+com.name+"  ") > cmdPadding {
				cmdPadding = len("  " + com.name + "  ")
			}
		}
		// Now add commands with known padding
		for _, com := range commands {
			if com.description == DisableDescription {
				continue
			}
			cmd := "  " + com.name
			cmd = cmd + strings.Repeat(" ", cmdPadding-len(cmd)-1)
			cmd = addToLastLine(cmd, com.description, maxWidth, cmdPadding, true)
			cmdContent = cmdContent + cmd + "\n"
		}
		result = result + cmdContent + "\n"
	}
	return result
}

// arguments2Result - puts info about all arguments of current command into result string buffer
func arguments2Result(result string, arguments []*arg, maxWidth int) string {
	usedHelp := false
	if len(arguments) > 0 {
		argContent := "Arguments:\n\n"
		// Get biggest padding
		var argPadding int
		// Find biggest padding
		for _, argument := range arguments {
			if argument.opts.Help == DisableDescription {
				continue
			}
			if len(argument.lname)+9 > argPadding {
				argPadding = len(argument.lname) + 9
			}
		}
		// Now add args with padding
		for _, argument := range arguments {
			if argument.opts.Help == DisableDescription {
				continue
			}
			if argument.lname == "help" && usedHelp {
			} else {
				arg := "  "
				if argument.sname != "" {
					arg = arg + "-" + argument.sname + "  "
				} else {
					arg = arg + "    "
				}
				arg = arg + "--" + argument.lname
				arg = arg + strings.Repeat(" ", argPadding-len(arg))
				if argument.opts != nil && argument.opts.Help != "" {
					arg = addToLastLine(arg, argument.getHelpMessage(), maxWidth, argPadding, true)
				}
				argContent = argContent + arg + "\n"
			}
			if argument.lname == "help" || argument.sname == "h" {
				usedHelp = true
			}
		}
		result = result + argContent + "\n"
	}
	return result
}

// Happened shows whether Command was specified on CLI arguments or not. If Command did not "happen", then
// all its descendant commands and arguments are not parsed. Returns a boolean value.
func (o *Command) Happened() bool {
	return o.happened
}

// Usage returns a multiline string that is the same as a help message for this Parser or Command.
// Since Parser is a Command as well, they work in exactly same way. Meaning that usage string
// can be retrieved for any level of commands. It will only include information about this Command,
// its sub-commands, current Command arguments and arguments of all preceding commands (if any)
//
// Accepts an interface that can be error, string or fmt.Stringer that will be prepended to a message.
// All other interface types will be ignored
func (o *Command) Usage(msg interface{}) string {
	for _, cmd := range o.commands {
		if cmd.Happened() {
			return cmd.Usage(msg)
		}
	}

	// Stay classy
	maxWidth := 80
	// List of arguments from all preceding commands
	arguments := make([]*arg, 0)
	// Line of commands until root
	var chain []string

	// Put message in result
	result, done := message2String(msg)
	if done {
		return result
	}

	//collect info about Preceding Commands into chain and arguments
	o.getPrecedingCommands(&chain, &arguments)
	// If this Command has sub-commands we need their list
	commands := o.getSubCommands(&chain)

	// Build usage description from description of preceding commands chain and each of subcommands
	result += "usage:"
	result = o.precedingCommands2Result(result, chain, arguments, maxWidth)
	result = subCommands2Result(result, commands, maxWidth)
	// Add list of arguments to the result
	result = arguments2Result(result, arguments, maxWidth)

	return result
}

// Parse method can be applied only on Parser. It takes a slice of strings (as in os.Args)
// and it will process this slice as arguments of CLI (the original slice is not modified).
// Returns error on any failure. In case of failure recommended course of action is to
// print received error alongside with usage information (might want to check which Command
// was active when error happened and print that specific Command usage).
// In case no error returned all arguments should be safe to use. Safety of using arguments
// before Parse operation is complete is not guaranteed.
func (o *Parser) Parse(args []string) error {
	subargs := make([]string, len(args))
	copy(subargs, args)

	result := o.parse(&subargs)
	unparsed := make([]string, 0)
	for _, v := range subargs {
		if v != "" {
			unparsed = append(unparsed, v)
		}
	}
	if result == nil && len(unparsed) > 0 {
		return errors.New("unknown arguments " + strings.Join(unparsed, " "))
	}

	return result
}
