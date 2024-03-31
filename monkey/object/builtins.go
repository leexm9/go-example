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
				default:
					return NewError("argument to `len` not support, got %s", arg.Type())
				}
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
