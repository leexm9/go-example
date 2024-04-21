package vm

import (
	"fmt"
	"go-example/monkey/code"
	"go-example/monkey/compiler"
	"go-example/monkey/object"
)

const (
	StackSize  = 2048
	GlobalSize = 65536
	MaxFrames  = 1024
)

type VM struct {
	constants []object.Object
	globals   []object.Object

	stack []object.Object
	sp    int //始终指向栈中的下一个空闲槽

	frames     []*Frame
	frameIndex int
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants: bytecode.Constants,
		globals:   make([]object.Object, GlobalSize),

		stack: make([]object.Object, StackSize),
		sp:    0,

		frames:     frames,
		frameIndex: 1,
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

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++
		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			constIdx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.constants[constIdx])
			if err != nil {
				return err
			}
		case code.OpClosure:
			idx := code.ReadUint16(ins[ip+1:])
			numFree := code.ReadUint8(ins[ip+3:])
			vm.currentFrame().ip += 3
			err := vm.pushClosure(int(idx), int(numFree))
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			idx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[idx] = vm.pop()
		case code.OpGetGlobal:
			idx := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.globals[idx])
			if err != nil {
				return err
			}
		case code.OpSetLocal:
			idx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			vm.stack[frame.basePointer+int(idx)] = vm.pop()
		case code.OpGetLocal:
			idx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			err := vm.push(vm.stack[frame.basePointer+int(idx)])
			if err != nil {
				return err
			}
		case code.OpGetBuiltin:
			btIdx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			definition := object.Builtins[btIdx]
			err := vm.push(definition.Builtin)
			if err != nil {
				return err
			}
		case code.OpGetFree:
			idx := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			currentClosure := vm.currentFrame().cl
			err := vm.push(currentClosure.Free[idx])
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
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !object.IsTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case code.OpNull:
			err := vm.push(object.NULL)
			if err != nil {
				return err
			}
		case code.OpArray:
			numElems := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			array := vm.buildArray(vm.sp-numElems, vm.sp)
			vm.sp = vm.sp - numElems
			err := vm.push(array)
			if err != nil {
				return err
			}
		case code.OpHash:
			numElems := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			hash, err := vm.buildHash(vm.sp-numElems, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElems
			err = vm.push(hash)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()
			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return nil
			}
		case code.OpCall:
			numArgs := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			err := vm.executeCall(int(numArgs))
			if err != nil {
				return err
			}
		case code.OpCurrentClosure:
			currentClosure := vm.currentFrame().cl
			err := vm.push(currentClosure)
			if err != nil {
				return err
			}
		case code.OpReturnValue:
			retValue := vm.pop()

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(retValue)
			if err != nil {
				return nil
			}
		case code.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1
			err := vm.push(object.NULL)
			if err != nil {
				return nil
			}
		}
	}
	return nil
}

func (vm *VM) pushClosure(index, numFree int) error {
	constant := vm.constants[index]
	fn, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function:%+v", constant)
	}

	free := make([]object.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.sp-numFree+i]
	}
	vm.sp = vm.sp - numFree

	closure := &object.Closure{Fn: fn, Free: free}
	return vm.push(closure)
}

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.Closure:
		return vm.callClosure(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltin(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function")
	}
}

func (vm *VM) callClosure(cl *object.Closure, numArgs int) error {
	if numArgs != cl.Fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d", cl.Fn.NumParameters, numArgs)
	}

	frame := NewFrame(cl, vm.sp-numArgs)
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + cl.Fn.NumLocals
	return nil
}

func (vm *VM) callBuiltin(builtin *object.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]
	result := builtin.Fn(args...)
	if result != nil {
		return vm.push(result)
	} else {
		return vm.push(object.NULL)
	}
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.BOOLEAN_OBJ && rightType == object.BOOLEAN_OBJ:
		return vm.executeBinaryBooleanOperation(op, left, right)
	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
	}
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

func (vm *VM) executeBinaryStringOperation(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch op {
	case code.OpAdd:
		return vm.push(&object.String{Value: leftVal + rightVal})
	default:
		return fmt.Errorf("unknown string operator: %d", op)
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

func (vm *VM) buildArray(start, end int) object.Object {
	elements := make([]object.Object, end-start)
	for i := start; i < end; i++ {
		elements[i-start] = vm.stack[i]
	}
	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(start, end int) (object.Object, error) {
	pairs := make(map[object.HashKey]object.HashPair)
	for i := start; i < end; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]
		pair := object.HashPair{Key: key, Value: value}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}
		pairs[hashKey.HashKey()] = pair
	}
	return &object.Hash{Pairs: pairs}, nil
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObj := array.(*object.Array)
	i := index.(*object.Integer).Value
	length := len(arrayObj.Elements) - 1
	if i < 0 || i > int64(length) {
		return vm.push(object.NULL)
	}
	return vm.push(arrayObj.Elements[i])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObj := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}
	pair, ok := hashObj.Pairs[key.HashKey()]
	if !ok {
		return vm.push(object.NULL)
	}
	return vm.push(pair.Value)
}

func nativeBool2Object(input bool) *object.Boolean {
	if input {
		return object.True
	} else {
		return object.False
	}
}

func NewWithGlobalStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}
