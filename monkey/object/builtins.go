package object

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		"len",
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return NewError("wrong number of arguments. got=%d, want=1", len(args))
				}
				switch arg := args[0].(type) {
				case *String:
					return &Integer{Value: int64(len(arg.Value))}
				case *Array:
					return &Integer{Value: int64(len(arg.Elements))}
				default:
					return NewError("argument to `len` not supported, got %s", arg.Type())
				}
			},
		},
	},
	{
		"push",
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 2 {
					return NewError("wrong number of arguments. got=%d, want=2", len(args))
				}
				if args[0].Type() != ARRAY_OBJ {
					return NewError("argument to `push` must be Array, got %s", args[0].Type())
				}
				array := args[0].(*Array)
				length := len(array.Elements)

				newElems := make([]Object, length+1)
				copy(newElems, array.Elements)
				newElems[length] = args[1]
				return &Array{Elements: newElems}
			},
		},
	},
}

func GetBuiltingByName(name string) *Builtin {
	for _, item := range Builtins {
		if item.Name == name {
			return item.Builtin
		}
	}
	return nil
}
