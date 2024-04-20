package compiler

import (
	"fmt"
	"go-example/monkey/ast"
	"go-example/monkey/code"
	"go-example/monkey/object"
	"sort"
)

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type Compiler struct {
	constants   []object.Object
	symbolTable *SymbolTable

	scopes     []CompilationScope
	scopeIndex int
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	return &Compiler{
		constants:   []object.Object{},
		symbolTable: NewSymbolTable(),
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
	}
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)
	return pos
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}
	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewIns := len(c.currentInstructions())
	updatedIns := append(c.currentInstructions(), ins...)

	c.scopes[c.scopeIndex].instructions = updatedIns
	return posNewIns
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	oldIns := c.currentInstructions()
	newIns := oldIns[:last.Position]
	c.scopes[c.scopeIndex].instructions = newIns
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.currentInstructions()[opPos])
	newIns := code.Make(op, operand)
	c.replaceInstruction(opPos, newIns)
}

func (c *Compiler) replaceInstruction(pos int, newIns []byte) {
	ins := c.currentInstructions()
	for i := 0; i < len(newIns); i++ {
		ins[pos+i] = newIns[i]
	}
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))
	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
}

func (c *Compiler) leaveScope() code.Instructions {
	ins := c.currentInstructions()
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	return ins
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, stmt := range node.Statements {
			err := c.Compile(stmt)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		c.emit(code.OpSetGlobal, symbol.Index)
	case *ast.InfixExpression:
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}
			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			return nil
		}
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case ">":
			c.emit(code.OpGreaterThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "-":
			c.emit(code.OpMinus)
		case "!":
			c.emit(code.OpBang)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		jumpTruthyPos := c.emit(code.OpJumpNotTruthy, 0)
		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		jumpPos := c.emit(code.OpJump, 0)
		afterConPos := len(c.currentInstructions())
		c.changeOperand(jumpTruthyPos, afterConPos)

		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err = c.Compile(node.Alternative)
			if err != nil {
				return err
			}
		}
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		afterAltPos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAltPos)
	case *ast.BlockStatement:
		for _, stmt := range node.Statements {
			err := c.Compile(stmt)
			if err != nil {
				return err
			}
		}
	case *ast.ArrayLiteral:
		for _, elem := range node.Elements {
			err := c.Compile(elem)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(node.Elements))
	case *ast.HashLiteral:
		var keys []ast.Expression
		for k := range node.Pairs {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})
		for _, key := range keys {
			err := c.Compile(key)
			if err != nil {
				return err
			}
			err = c.Compile(node.Pairs[key])
			if err != nil {
				return err
			}
		}
		c.emit(code.OpHash, len(node.Pairs)*2)
	case *ast.FunctionLiteral:
		c.enterScope()
		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		ins := c.leaveScope()
		compiledFn := &object.CompiledFunction{Instructions: ins}
		c.emit(code.OpConstant, c.addConstant(compiledFn))
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}
		c.emit(code.OpReturnValue)
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}
		c.emit(code.OpCall)
	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Index)
		if err != nil {
			return err
		}
		c.emit(code.OpIndex)
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		c.emit(code.OpGetGlobal, symbol.Index)
	case *ast.IntegerLiteral:
		intg := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(intg))
	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	}

	return nil
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.constants = constants
	compiler.symbolTable = s
	return compiler
}
