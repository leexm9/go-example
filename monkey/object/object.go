package object

import (
	"bytes"
	"fmt"
	"go-example/monkey/ast"
	"strconv"
	"strings"
)

type ObjectType string

const (
	INTEGER_OBJ      = "integer"
	BOOLEAN_OBJ      = "boolean"
	NULL_OBJ         = "null"
	RETURN_VALUE_OBJ = "return_value"
	ERROR_OBJ        = "error"
	FUNCTION_OBJ     = "function"
	STRING_OBJ       = "string"
	BUILTIN_OBJ      = "builtin"
)

var (
	TRUE  = &Boolean{Value: true}
	FALSE = &Boolean{Value: false}
	NULL  = &Null{}
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string { return strconv.FormatInt(i.Value, 10) }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return strconv.FormatBool(b.Value) }

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

type String struct {
	Value string
}

func (s String) Type() ObjectType { return STRING_OBJ }
func (s String) Inspect() string  { return s.Value }

type ReturnValue struct {
	Value Object
}

func (rt *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rt *ReturnValue) Inspect() string  { return rt.Value.Inspect() }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return fmt.Sprintf("Message: ", e.Message) }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (fn *Function) Type() ObjectType { return FUNCTION_OBJ }
func (fn *Function) Inspect() string {
	var out bytes.Buffer

	var params []string
	for _, parameter := range fn.Parameters {
		params = append(params, parameter.String())
	}

	out.WriteString("fn(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(fn.Body.String())
	out.WriteString("\n}")

	return out.String()
}

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }
