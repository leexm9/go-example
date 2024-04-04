package evaluator

import (
	"go-example/monkey/ast"
	"go-example/monkey/object"
	"go-example/monkey/token"
	"strconv"
)

func quote(node ast.Node, env *object.Environment) object.Object {
	node = evalUnquoteCells(node, env)
	return &object.Quote{Node: node}
}

func evalUnquoteCells(quoted ast.Node, env *object.Environment) ast.Node {
	return ast.Modify(quoted, func(node ast.Node) ast.Node {
		if !isUnquoteCells(node) {
			return node
		}

		call, ok := node.(*ast.CallExpression)
		if !ok {
			return node
		}

		if len(call.Arguments) != 1 {
			return node
		}

		unquote := Eval(call.Arguments[0], env)
		return convertObjectToASTNode(unquote)
	})
}

func isUnquoteCells(node ast.Node) bool {
	callExpression, ok := node.(*ast.CallExpression)
	if !ok {
		return false
	}
	return callExpression.Function.TokenLiteral() == "unquote"
}

func convertObjectToASTNode(obj object.Object) ast.Node {
	switch obj := obj.(type) {
	case *object.Integer:
		t := token.Token{
			Type:    token.INT,
			Literal: strconv.FormatInt(obj.Value, 10),
		}
		return &ast.IntegerLiteral{Token: t, Value: obj.Value}
	case *object.Boolean:
		var t token.Token
		if obj == object.TRUE {
			t = token.Token{Type: token.TRUE, Literal: "true"}
		} else {
			t = token.Token{Type: token.FALSE, Literal: "false"}
		}
		return &ast.Boolean{Token: t, Value: obj.Value}
	case *object.Quote:
		return obj.Node
	default:
		return nil
	}
}
