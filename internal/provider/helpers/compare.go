package helpers

import "github.com/go-test/deep"

// ObjectsEqual checks to see if two objects are equivalent, accounting for
// differences in the order of the contents.
func ObjectsEqual(obj1, obj2 interface{}) (bool, []string) {
	differences := deep.Equal(obj1, obj2)
	if len(differences) != 0 {
		return false, differences
	}

	return true, nil
}
