package argparse

import "strings"

func getLastLine(input string) string {
	slice := strings.Split(input, "\n")
	return slice[len(slice)-1]
}

func addToLastLine(base string, add string, width int, padding int, canSplit bool) string {
	// If last line has less than 10% space left, do not try to fill in by splitting else just try to split
	hasTen := (width - len(getLastLine(base))) > width/10
	if len(getLastLine(base)+" "+add) >= width {
		if hasTen && canSplit {
			adds := strings.Split(add, " ")
			for _, v := range adds {
				base = addToLastLine(base, v, width, padding, false)
			}
			return base
		}
		base = base + "\n" + strings.Repeat(" ", padding)
	}
	base = base + " " + add
	return base
}
