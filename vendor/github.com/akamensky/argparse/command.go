package argparse

import (
	"fmt"
	"strings"
)

func (o *Command) help(sname, lname string) {
	result := &help{}

	if lname == "" {
		sname, lname = "h", "help"
	}

	a := &arg{
		result: result,
		sname:  sname,
		lname:  lname,
		size:   1,
		opts:   &Options{Help: "Print help information"},
		unique: true,
	}

	o.addArg(a)
}

func (o *Command) addArg(a *arg) error {
	// long name should be provided
	if a.lname == "" {
		return fmt.Errorf("long name should be provided")
	}
	// short name could be provided and must not exceed 1 character
	if len(a.sname) > 1 {
		return fmt.Errorf("short name must not exceed 1 character")
	}
	// Search parents for overlapping commands and fail if any
	current := o
	for current != nil {
		if current.args != nil {
			for _, v := range current.args {
				if a.lname != "help" || a.sname != "h" {
					if a.sname != "" && a.sname == v.sname {
						return fmt.Errorf("short name %s occurs more than once", a.sname)
					}
					if a.lname == v.lname {
						return fmt.Errorf("long name %s occurs more than once", a.lname)
					}
				}
			}
		}
		current = current.parent
	}
	a.parent = o
	o.args = append(o.args, a)
	return nil
}

//parseSubCommands - Parses subcommands if any
func (o *Command) parseSubCommands(args *[]string) error {
	if o.commands != nil && len(o.commands) > 0 {
		// If we have subcommands and 0 args left
		// that is an error of SubCommandError type
		if len(*args) < 1 {
			return newSubCommandError(o)
		}
		for _, v := range o.commands {
			err := v.parse(args)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//parseArguments - Parses arguments
func (o *Command) parseArguments(args *[]string) error {
	// Iterate over the args
	for i := 0; i < len(o.args); i++ {
		oarg := o.args[i]
		for j := 0; j < len(*args); j++ {
			arg := (*args)[j]
			if arg == "" {
				continue
			}
			if strings.Contains(arg, "=") {
				splitInd := strings.LastIndex(arg, "=")
				equalArg := []string{arg[:splitInd], arg[splitInd+1:]}
				if cnt, err := oarg.check(equalArg[0]); err != nil {
					return err
				} else if cnt > 0 {
					if equalArg[1] == "" {
						return fmt.Errorf("not enough arguments for %s", oarg.name())
					}
					oarg.eqChar = true
					oarg.size = 1
					currArg := []string{equalArg[1]}
					err := oarg.parse(currArg, cnt)
					if err != nil {
						return err
					}
					oarg.reduce(j, args)
					continue
				}
			}
			if cnt, err := oarg.check(arg); err != nil {
				return err
			} else if cnt > 0 {
				if len(*args) < j+oarg.size {
					return fmt.Errorf("not enough arguments for %s", oarg.name())
				}
				err := oarg.parse((*args)[j+1:j+oarg.size], cnt)
				if err != nil {
					return err
				}
				oarg.reduce(j, args)
				continue
			}
		}

		// Check if arg is required and not provided
		if oarg.opts != nil && oarg.opts.Required && !oarg.parsed {
			return fmt.Errorf("[%s] is required", oarg.name())
		}

		// Check for argument default value and if provided try to type cast and assign
		if oarg.opts != nil && oarg.opts.Default != nil && !oarg.parsed {
			err := oarg.setDefault()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Will parse provided list of arguments
// common usage would be to pass directly os.Args
func (o *Command) parse(args *[]string) error {
	// If we already been parsed do nothing
	if o.parsed {
		return nil
	}

	// If no arguments left to parse do nothing
	if len(*args) < 1 {
		return nil
	}

	// Parse only matching commands
	// But we always have to parse top level
	if o.name == "" {
		o.name = (*args)[0]
	} else {
		if o.name != (*args)[0] && o.parent != nil {
			return nil
		}
	}

	// Set happened status to true when command happend
	o.happened = true

	// Reduce arguments by removing Command name
	*args = (*args)[1:]

	// Parse subcommands if any
	if err := o.parseSubCommands(args); err != nil {
		return err
	}

	// Parse arguments if any
	if err := o.parseArguments(args); err != nil {
		return err
	}

	// Set parsed status to true and return quietly
	o.parsed = true
	return nil
}
