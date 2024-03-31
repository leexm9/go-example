package object

import "fmt"

func NewError(format string, a ...any) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func IsError(obj Object) bool {
	if obj != nil {
		return obj.Type() == ERROR_OBJ
	}
	return false
}
