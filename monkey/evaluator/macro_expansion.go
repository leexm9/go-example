package evaluator

import (
	"go-example/monkey/ast"
	"go-example/monkey/object"
)

func DefineMacros(program *ast.Program, env *object.Environment) {
	var definitions []int

	for i, stmt := range program.Statements {
		if isMacroDefinition(stmt) {
			addMacro(stmt, env)
			definitions = append(definitions, i)
		}
	}

	for i := len(definitions) - 1; i >= 0; i-- {
		definitionIndex := definitions[i]
		program.Statements = append(
			program.Statements[:definitionIndex],
			program.Statements[definitionIndex+1:]...,
		)
	}
}

func ExpandMacros(program *ast.Program, env *object.Environment) ast.Node {
	return ast.Modify(program, func(node ast.Node) ast.Node {
		callExpr, ok := node.(*ast.CallExpression)
		if !ok {
			return node
		}

		macro, ok := isMacroCall(callExpr, env)
		if !ok {
			return node
		}

		args := quoteArgs(callExpr)
		evalEnv := extendMacroEnv(macro, args)

		evaluated := Eval(macro.Body, evalEnv)
		quote, ok := evaluated.(*object.Quote)
		if !ok {
			panic("we only support returning AST-node from macros")
		}

		return quote.Node
	})
}

func isMacroDefinition(stmt ast.Statement) bool {
	letStmt, ok := stmt.(*ast.LetStatement)
	if !ok {
		return false
	}

	_, ok = letStmt.Value.(*ast.MacroLiteral)
	if !ok {
		return false
	}

	return true
}

func addMacro(stmt ast.Statement, env *object.Environment) {
	letStmt, _ := stmt.(*ast.LetStatement)
	macroLit, _ := letStmt.Value.(*ast.MacroLiteral)

	macro := &object.Macro{
		Parameters: macroLit.Parameters,
		Body:       macroLit.Body,
		Env:        env,
	}

	env.Set(letStmt.Name.Value, macro)
}

func isMacroCall(expr *ast.CallExpression, env *object.Environment) (*object.Macro, bool) {
	ident, ok := expr.Function.(*ast.Identifier)
	if !ok {
		return nil, false
	}

	obj, ok := env.Get(ident.Value)
	if !ok {
		return nil, false
	}

	macro, ok := obj.(*object.Macro)
	if !ok {
		return nil, false
	}

	return macro, true
}

func quoteArgs(expr *ast.CallExpression) []*object.Quote {
	var args []*object.Quote
	for _, argument := range expr.Arguments {
		args = append(args, &object.Quote{Node: argument})
	}
	return args
}

func extendMacroEnv(macro *object.Macro, args []*object.Quote) *object.Environment {
	extended := object.NewEnclosedEnvironment(macro.Env)
	for i, parameter := range macro.Parameters {
		extended.Set(parameter.Value, args[i])
	}
	return extended
}
