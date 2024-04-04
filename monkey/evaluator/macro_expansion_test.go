package evaluator

import (
	"go-example/monkey/ast"
	"go-example/monkey/lexer"
	"go-example/monkey/object"
	"go-example/monkey/parser"
	"testing"
)

func TestDefineMacros(t *testing.T) {
	input := ` 
				let number = 1; 
				let function = fn(x, y) { x + y }; 
				let mymacro = macro(x, y) { x + y }; 
			`
	env := object.NewEnvironment()
	program := testParseProgram(t, input)

	DefineMacros(program, env)

	if len(program.Statements) != 2 {
		t.Fatalf("Wrong number of statements. got=%d, want=2.", len(program.Statements))
	}
	_, ok := env.Get("number")
	if ok {
		t.Fatalf("number should not be defined.")
	}

	_, ok = env.Get("function")
	if ok {
		t.Fatalf("function should not be defined.")

	}

	obj, ok := env.Get("mymacro")
	if !ok {
		t.Fatalf("mymacro not in environment.")
	}

	macro, ok := obj.(*object.Macro)
	if !ok {
		t.Fatalf("object is not Macro. got=%T(%+v)", obj, obj)
	}
	if len(macro.Parameters) != 2 {
		t.Fatalf("Wrong number of macro parameters. got=%d", len(macro.Parameters))
	}
	if macro.Parameters[0].String() != "x" {
		t.Fatalf("paramater is not 'x', got=%q.", macro.Parameters[0])
	}
	if macro.Parameters[1].String() != "y" {
		t.Fatalf("paramater is not '', got=%q.", macro.Parameters[1])
	}
	expectedBody := "(x + y)"
	if macro.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, macro.Body.String())
	}
}

func TestExpandMacro(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`
					let infixExpression = macro() {quote(1 + 2);};
					infixExpression()
				`,
			"(1 + 2)",
		},
		{`
				let reverse = macro(a, b) { quote(unquote(b) - unquote(a)); };
				reverse(2 + 2, 10 - 5);
			`,
			"(10 - 5) - (2 + 2)",
		},
		{`
				let unless = macro(condition, consequence, alternative) {
					quote(if (!(unquote(condition))) {
						unquote(consequence);
					} else {
						unquote(alternative);
					});
				};
		
				unless(10 > 5, print("not greater"), print("greater"));
			`,
			`if (!(10 > 5)) { print("not greater") } else { print("greater") }`,
		},
	}
	for _, tt := range tests {
		program := testParseProgram(t, tt.input)
		expected := testParseProgram(t, tt.expected)
		env := object.NewEnvironment()
		DefineMacros(program, env)
		expanded := ExpandMacros(program, env)
		if expanded.String() != expected.String() {
			t.Errorf("not equal. want=%q, got=%q", expected.String(), expanded.String())
		}
	}
}

func testParseProgram(t *testing.T, input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	errors := p.Errors()
	if len(errors) > 0 {
		t.Errorf("parser has %d errors.", len(errors))
		for _, msg := range errors {
			t.Errorf("parser error: %q", msg)
		}
		t.FailNow()
	}
	return program
}
