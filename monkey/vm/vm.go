package vm

import (
	"fmt"
	"go-example/monkey/code"
	"go-example/monkey/compiler"
	"go-example/monkey/object"
)

const StackSize = 2048

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int //始终指向栈中的下一个空闲槽
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		constants:    bytecode.Constants,
		instructions: bytecode.Instructions,

		stack: make([]object.Object, StackSize),
		sp:    0,
	}
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) push(obj object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = obj
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object {
	obj := vm.stack[vm.sp-1]
	vm.sp--
	return obj
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			constIdx := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			err := vm.push(vm.constants[constIdx])
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv,
			code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return nil
			}
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return nil
			}
		case code.OpPop:
			vm.pop()
		case code.OpTrue:
			err := vm.push(object.True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(object.False)
			if err != nil {
				return err
			}
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2

			condition := vm.pop()
			if !object.IsTruthy(condition) {
				ip = pos - 1
			}
		case code.OpJump:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip = pos - 1
		case code.OpNull:
			err := vm.push(object.NULL)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(op, left, right)
	} else if leftType == object.BOOLEAN_OBJ && rightType == object.BOOLEAN_OBJ {
		return vm.executeBinaryBooleanOperation(op, left, right)
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case code.OpAdd:
		return vm.push(&object.Integer{Value: leftVal + rightVal})
	case code.OpSub:
		return vm.push(&object.Integer{Value: leftVal - rightVal})
	case code.OpMul:
		return vm.push(&object.Integer{Value: leftVal * rightVal})
	case code.OpDiv:
		return vm.push(&object.Integer{Value: leftVal / rightVal})
	case code.OpEqual:
		return vm.push(nativeBool2Object(leftVal == rightVal))
	case code.OpNotEqual:
		return vm.push(nativeBool2Object(leftVal != rightVal))
	case code.OpGreaterThan:
		return vm.push(&object.Boolean{Value: leftVal > rightVal})
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}
}

func (vm *VM) executeBinaryBooleanOperation(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Boolean).Value
	rightVal := right.(*object.Boolean).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBool2Object(leftVal == rightVal))
	case code.OpNotEqual:
		return vm.push(nativeBool2Object(leftVal != rightVal))
	default:
		return fmt.Errorf("unknown boolean operator: %d", op)
	}
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()
	switch operand {
	case object.True:
		return vm.push(object.False)
	case object.False:
		return vm.push(object.True)
	case object.NULL:
		return vm.push(object.True)
	default:
		return vm.push(object.False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupport type for negation: %s", operand.Type())
	}

	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
}

func nativeBool2Object(input bool) *object.Boolean {
	if input {
		return object.True
	} else {
		return object.False
	}
}
