package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewVM(t *testing.T) {
	// Reverse Polish Notation.
	// 1 + 2 = 3
	// 1
	// push stack
	// 2
	// push stack
	// add
	// 3
	// push stack

	//data := []byte{0x01, 0x0a, 0x02, 0x0a, 0x0b}

	// string
	//data := []byte{0x02,0x0a, 0x61, 0x0c, 0x61, 0x0c, 0x0d}
	// F o o => pack
	// [store]
	//data := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d, 0x05, 0x0a, 0x0f}
	pushFoo := []byte{0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0xae}
	data := []byte{0x02, 0x0a, 0x03, 0x0a, 0x0b, 0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
	data = append(data, pushFoo...)
	contractState := NewState()
	vm := NewVM(data, contractState)
	vm.Run()
	t.Log(vm.stack.data)
}

func TestStack(t *testing.T) {
	s := NewStack(128)
	s.Push(1)
	s.Push(2)

	value := s.Pop()
	assert.Equal(t, 2, value)
	t.Log(s.data)

	value = s.Pop()
	assert.Equal(t, 1, value)
	t.Log(s.data)
}

func TestStackBytes(t *testing.T) {

}
