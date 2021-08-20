package argparse

import (
	"os"
	"reflect"
)

// IsNilFile allows to test whether returned `*os.File` has been initialized with data passed on CLI.
// Returns true if `fd == &{nil}`, which means `*os.File` was not initialized, false if `fd` is
// a fully initialized `*os.File` or if `fd == nil`.
func IsNilFile(fd *os.File) bool {
	return reflect.DeepEqual(fd, &os.File{})
}
