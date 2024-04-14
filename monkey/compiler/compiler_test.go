package compiler

import (
	"fmt"
	"go-example/monkey/code"
	"go-example/monkey/lexer"
	"go-example/monkey/object"
	"go-example/monkey/parser"
	"testing"
)

type CompilerTestCase struct {
	input             string
	expectedConstants []any
	expectedIns       []code.Instructions
}

func TestIntegerArithmetic(t *testing.T) {
	test := []CompilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []any{1, 2},
			expectedIns: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
			},
		},
	}

	runCompilerTests(t, test)
}

func runCompilerTests(t *testing.T, tests []CompilerTestCase) {
	t.Helper()
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		prog := p.ParseProgram()

		compiler := New()
		err := compiler.Compile(prog)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()
		err = testInstructions(tt.expectedIns, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}
		err = testConstants(t, tt.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}
	}
}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concatted := concatInstructions(expected)
	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instructions length.\nwant=%q\ngot=%d", concatted, actual)
	}
	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction as %d.\nwant=%q\ngot=%q", i, ins, actual[i])
		}
	}
	return nil
}

func concatInstructions(s []code.Instructions) code.Instructions {
	out := code.Instructions{}
	for _, ins := range s {
		out = append(out, ins...)
	}
	return out
}

func testConstants(t *testing.T, expected []any, actual []object.Object) error {
	t.Helper()
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. want=%d, got=%d", len(expected), len(actual))
	}
	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testIntegerObject(int64(constant), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s", i, err)
			}
		}
	}
	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)", actual, actual)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. want=%d, got=%d", expected, result.Value)
	}
	return nil
}