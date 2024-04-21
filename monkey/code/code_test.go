package code

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
		{OpAdd, []int{}, []byte{byte(OpAdd)}},
		{OpPop, []int{}, []byte{byte(OpPop)}},
		{OpAdd, []int{}, []byte{byte(OpAdd)}},
		{OpSub, []int{}, []byte{byte(OpSub)}},
		{OpMul, []int{}, []byte{byte(OpMul)}},
		{OpDiv, []int{}, []byte{byte(OpDiv)}},
		{OpSetGlobal, []int{65534}, []byte{byte(OpSetGlobal), 255, 254}},
		{OpSetLocal, []int{255}, []byte{byte(OpSetLocal), 255}},
		{OpCall, []int{255}, []byte{byte(OpCall), 255}},
		{OpClosure, []int{65534, 255}, []byte{byte(OpClosure), 255, 254, 255}},
	}

	for _, tt := range tests {
		instruction := Make(tt.op, tt.operands...)

		if len(instruction) != len(tt.expected) {
			t.Errorf("instruction has wrong length. want=%d, got=%d", len(tt.expected), len(instruction))
		}

		for i, b := range tt.expected {
			if instruction[i] != b {
				t.Errorf("wrong byte as pos %d. want=%d, got=%d", i, b, instruction[i])
			}
		}
	}
}

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
		Make(OpPop),
		Make(OpSub),
		Make(OpMul),
		Make(OpDiv),
		Make(OpTrue),
		Make(OpFalse),
		Make(OpEqual),
		Make(OpNotEqual),
		Make(OpGreaterThan),
		Make(OpMinus),
		Make(OpBang),
		Make(OpJumpNotTruthy, 13),
		Make(OpJump, 23),
		Make(OpClosure, 4, 2),
	}

	expected := `0000 OpAdd
0001 OpConstant 2
0004 OpConstant 65535
0007 OpPop
0008 OpSub
0009 OpMul
0010 OpDiv
0011 OpTrue
0012 OpFalse
0013 OpEqual
0014 OpNotEqual
0015 OpGreaterThan
0016 OpMinus
0017 OpBang
0018 OpJumpNotTruthy 13
0021 OpJump 23
0024 OpClosure 4 2
`

	concatted := Instructions{}
	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}

	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted.\nwant=%q\ngot=%q", expected, concatted.String())
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
		{OpSetGlobal, []int{65535}, 2},
		{OpSetLocal, []int{255}, 1},
		{OpClosure, []int{65535, 255}, 3},
	}

	for _, tt := range tests {
		instruction := Make(tt.op, tt.operands...)

		def, err := Lookup(byte(tt.op))
		if err != nil {
			t.Fatalf("definition not found: %q\n", err)
		}
		operandsRead, n := ReadOperands(def, instruction[1:])
		if n != tt.bytesRead {
			t.Fatalf("n wrong. want=%d, got=%d", tt.bytesRead, n)
		}
		for i, want := range tt.operands {
			if operandsRead[i] != want {
				t.Errorf("operand wrong. want=%d, got=%d", want, operandsRead[i])
			}
		}
	}
}
