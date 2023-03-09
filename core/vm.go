package core

import (
	"fmt"
	"github.com/fzft/crypto-simple-blockchain/util"
)

type Instruction byte

const (
	InstrPushInt  Instruction = 0x0a //10
	InstrAdd      Instruction = 0x0b //11
	InstrPushByte Instruction = 0x0c //12
	InstrPack     Instruction = 0x0d //13
	InstrSub      Instruction = 0x0e //14
	InstrStore    Instruction = 0x0f //15
	InstrGet      Instruction = 0xae //16
	InstrMul      Instruction = 0xaf //17
	InstrDiv      Instruction = 0xb0 //18
)

type Stack struct {
	data []any
	sp   int
}

func NewStack(size int) *Stack {
	return &Stack{
		data: make([]any, size),
		sp:   0,
	}
}

func (s *Stack) Push(v any) {
	s.data = append([]any{v}, s.data...)
	s.sp++
}

func (s *Stack) Pop() any {
	s.sp--
	value := s.data[0]
	s.data = append(s.data[:0], s.data[1:]...)
	return value
}

type VM struct {
	data          []byte
	ip            int // instruction pointer
	stack         *Stack
	contractState *State
}

func NewVM(data []byte, contractState *State) *VM {
	return &VM{
		data:          data,
		ip:            0,
		stack:         NewStack(128),
		contractState: contractState,
	}
}

func (vm *VM) Run() error {
	for {
		instr := Instruction(vm.data[vm.ip])
		if err := vm.Exec(instr); err != nil {
			return err
		}
		vm.ip++
		if vm.ip > len(vm.data)-1 {
			break
		}

	}
	return nil
}

func (vm *VM) Exec(instr Instruction) error {
	switch instr {
	case InstrGet:
		key := vm.stack.Pop().([]byte)
		value, err := vm.contractState.Get(key)
		if err != nil {
			return err
		}
		vm.stack.Push(value)

	case InstrStore:
		var serialized []byte
		key := vm.stack.Pop().([]byte)
		value := vm.stack.Pop()
		switch v := value.(type) {
		case int:
			serialized = util.SerializeInt64(int64(v))
		default:
			panic("not implemented")
		}

		vm.contractState.Put(key, serialized)

	case InstrPushInt:
		vm.stack.Push(int(vm.data[vm.ip-1]))
	case InstrAdd:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		vm.stack.Push(a + b)
	case InstrPushByte:
		vm.stack.Push(byte(vm.data[vm.ip-1]))
	case InstrPack:
		n := vm.stack.Pop().(int)
		b := make([]byte, n)
		for i := 0; i < n; i++ {
			b[i] = vm.stack.Pop().(byte)
		}
		vm.stack.Push(b)
	case InstrSub:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		vm.stack.Push(a - b)
	case InstrMul:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		vm.stack.Push(a * b)
	case InstrDiv:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		if b == 0 {
			return fmt.Errorf("division by zero")
		}
		vm.stack.Push(a / b)
	}
	return nil
}
