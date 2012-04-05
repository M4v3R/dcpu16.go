package main

import (
	"./dcpu"
	"./dcpu/core"
	"fmt"
//	"strconv"
)

var notchAssemblerTestProgram = [...]core.Word{
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

func main() {
	vm := new(dcpu.Machine)
	fmt.Printf("Loading program...\n")
	vm.State.LoadProgram(notchAssemblerTestProgram[:], 0, true)
	fmt.Printf("Starting 0x10c VM...\n")
	vm.Start(1)
	fmt.Printf("Executing program...\n")
	for {
		if err := vm.State.StepCycle(); err != nil {
			break
		}
		if vm.State.Ram.GetWord(vm.State.PC()) == 0x85C3 { // sub PC, 1
			break
		}
	}

	for i := 0; ; i++{
		word := vm.State.Ram.GetWord(core.Word(0x8000+i))
		c1 := word & 0xff
		c2 := word >> 8
		fmt.Printf("%c%c", c1, c2)
		if word == 0 { break }
	}

	fmt.Printf("\n")
	fmt.Printf("Program ended.\n")
}
