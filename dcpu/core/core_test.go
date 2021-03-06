package core

import (
	"testing"
)

func TestLoadProgram(t *testing.T) {
	state := new(State)
	if err := state.LoadProgram(notchAssemblerTestProgram[:], 0, true); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(notchAssemblerTestProgram); i++ {
		if state.Ram.GetWord(Word(i)) != notchAssemblerTestProgram[i] {
			t.Errorf("Expected word %04x, found word %04x at offset %d", notchAssemblerTestProgram[i], state.Ram.GetWord(Word(i)), i)
			break
		}
	}
}

func TestNotchAssemblerTest(t *testing.T) {
	state := new(State)
	if err := state.LoadProgram(notchAssemblerTestProgram[:], 0, true); err != nil {
		t.Fatal(err)
	}
	if err := state.Start(); err != nil {
		t.Fatal(err)
	}

	// step the program for 1000 cycles, or until it hits the opcode 0x85C3
	// hitting 1000 cycles is considered failure
	for i := 0; i < 1000; i++ {
		t.Logf("%d: %04x", state.PC(), state.Ram.GetWord(state.PC()))
		if err := state.StepCycle(); err != nil {
			t.Fatal(err)
			break
		}
		if state.Ram.GetWord(state.PC()) == 0x85C3 { // sub PC, 1
			break
		}
	}
	if state.Ram.GetWord(state.PC()) != 0x85C3 {
		// we exhausted our steps
		t.Error("Program exceeded 1000 cycles")
	}
	// check 0x8000 - 0x800B for "Hello world!"
	expected := "Hello world!"
	for i := 0; i < len(expected); i++ {
		if state.Ram.GetWord(Word(0x8000+i)) != Word(expected[i]) {
			t.Errorf("Unexpected output in video ram; expected %v, found %v", []byte(expected), state.Ram.GetSlice(0x8000, 0x800B))
			break
		}
	}
	if err := state.Stop(); err != nil {
		t.Fatal(err)
	}
}

var notchAssemblerTestProgram = [...]Word{
	//              set a, 0xbeef
	0x7C01, // 0
	0xBEEF, // 1
	//              set [0x1000], a
	0x01E1, // 2
	0x1000, // 3
	//              ifn a, [0x1000]
	0x780D, // 4
	0x1000, // 5
	//                  set PC, end
	0x7DC1, // 6
	32,     // 7
	//
	//              set i, 0
	0x8061, // 8
	// :nextchar    ife [data+i], 0
	0x816C, // 9
	19,     // 10
	//                  set PC, end
	0x7DC1, // 11
	32,     // 12
	//              set [0x8000+i], [data+i]
	0x5961, // 13
	0x8000, // 14
	19,     // 15
	//              add i, 1
	0x8462, // 16
	//              set PC, nextchar
	0x7DC1, // 17
	9,      // 18
	//
	// :data        dat "Hello world!", 0
	'H', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd', '!', 0, // 19-31
	//
	// :end         sub PC, 1
	0x85C3, // 32
}

func TestNotchSpecExample(t *testing.T) {
	state := new(State)
	if err := state.LoadProgram(notchSpecExampleProgram[:], 0, true); err != nil {
		t.Fatal(err)
	}
	if err := state.Start(); err != nil {
		t.Fatal(err)
	}

	// test the first section
	for i := 0; i < 11; i++ {
		if err := state.StepCycle(); err != nil {
			t.Fatal(err)
		}
	}
	if state.A() != 0x10 {
		t.Errorf("Unexpected value for register A; expected %#x, found %#x", 0x10, state.A())
	}
	if state.PC() != 10 {
		t.Errorf("Unexpected value for register PC; expected %#x, found %#x", 10, state.A())
	}
	// run 23 more cycles (12 more instructions, partway into the loop)
	for i := 0; i < 23; i++ {
		if err := state.StepCycle(); err != nil {
			t.Fatal(err)
		}
	}
	if state.I() != 7 {
		t.Errorf("Unexpected value for register I; expected %#x, found %#x", 7, state.I())
	}
	// 59 more cycles (29 more instructions) to finish the loop
	for i := 0; i < 59; i++ {
		if err := state.StepCycle(); err != nil {
			t.Fatal(err)
		}
	}
	if state.I() != 0 {
		t.Errorf("Unexpected value for register I; expected %#x, found %#x", 0, state.I())
	}
	if state.PC() != 19 {
		t.Errorf("Unexpected value for register PC; expected %#x, found t#x", 19, state.PC())
	}
	if state.SP() != 0 {
		t.Errorf("Unexpected value for register SP; expected %#x, found %#x", 0, state.SP())
	}
	// 4 more cycles (2 more instructions) to put us into the subroutine
	for i := 0; i < 4; i++ {
		if err := state.StepCycle(); err != nil {
			t.Fatal(err)
		}
	}
	if state.X() != 4 {
		t.Errorf("Unexpected value for register X; expected %#x, found %#x", 4, state.X())
	}
	if state.PC() != 24 {
		t.Errorf("Unexpected value for register PC; expected %#x, found %#x", 24, state.PC())
	}
	if state.SP() != 0xffff {
		t.Errorf("Unexpected value for register SP; expected %#x, found %#x", 0xffff, state.SP())
	}
	if state.Ram.GetWord(0xffff) != 22 {
		t.Errorf("Unexpected value at 0xffff; expected %#x, found %#x", 22, state.Ram.GetWord(0xffff))
		t.FailNow()
	}
	if t.Failed() {
		if err := state.Stop(); err != nil {
			t.Fatal(err)
		}
		t.FailNow()
	}
	// run the program for 1000 cycles, or until it hits the instruction 0x7DC1 PC
	success := false
	for i := 0; i < 1000; i++ {
		if err := state.StepCycle(); err != nil {
			t.Fatal(err)
		}
		if state.Ram.GetWord(state.PC()) == 0x7DC1 && state.Ram.GetWord(state.PC()+1) == state.PC() {
			success = true
			break
		}
	}
	if !success {
		// we exhausted our steps
		t.Error("Program exceeded 1000 cycles")
	}

	// Check register X, it should be 0x40
	if state.X() != 0x40 {
		t.Error("Unexpected value for register X; expected %#x, found %#x", 0x40, state.X())
	}

	if err := state.Stop(); err != nil {
		t.Fatal(err)
	}
}

var notchSpecExampleProgram = [...]Word{
	0x7c01, 0x0030, 0x7de1, 0x1000, 0x0020, 0x7803, 0x1000, 0xc00d,
	0x7dc1, 0x001a, 0xa861, 0x7c01, 0x2000, 0x2161, 0x2000, 0x8463,
	0x806d, 0x7dc1, 0x000d, 0x9031, 0x7c10, 0x0018, 0x7dc1, 0x001a,
	0x9037, 0x61c1, 0x7dc1, 0x001a, 0x0000, 0x0000, 0x0000, 0x0000,
}
