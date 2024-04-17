package object

import "fmt"

var (
	True  = &Boolean{Value: true}
	False = &Boolean{Value: false}
)

func NewError(format string, a ...any) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func IsError(obj Object) bool {
	if obj != nil {
		return obj.Type() == ERROR_OBJ
	}
	return false
}

func IsTruthy(obj Object) bool {
	switch obj := obj.(type) {
	case *Boolean:
		return obj.Value
	case *Null:
		return false
	default:
		return true
	}
}
