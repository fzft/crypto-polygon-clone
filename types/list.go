package types

import (
	"fmt"
	"reflect"
)

type List[T any] struct {
	Data []T
}

func NewList[T any]() *List[T] {
	return &List[T]{Data: []T{}}
}

func (l *List[T]) Get(index int) T {
	if index > len(l.Data)-1 {
		err := fmt.Sprintf("the given index %d is out of range", index)
		panic(err)
	}

	return l.Data[index]
}

func (l *List[T]) Insert(item T) {
	l.Data = append(l.Data, item)
}

func (l *List[T]) GetIndex(v T) int {
	for i:=0; i<len(l.Data); i++ {
		if reflect.DeepEqual(l.Data[i], v) {
			return i
		}
	}
	return -1
}

func (l *List[T]) Remove(v T) {
	index := l.GetIndex(v)
	if index == -1 {
		return
	}
	l.Pop(index)
}

// Pop ...
func (l *List[T]) Pop(index int) {
	if index > len(l.Data)-1 {
		err := fmt.Sprintf("the given index %d is out of range", index)
		panic(err)
	}

	l.Data = append(l.Data[:index], l.Data[index+1:]...)
}

// Clear ...
func (l *List[T]) Clear() {
	l.Data = []T{}
}

// Contains ...
func (l *List[T]) Contains(v T) bool {
	return l.GetIndex(v) != -1
}

// Last ...
func (l *List[T]) Last() T {
	return l.Data[len(l.Data)-1]
}

// Len ...
func (l *List[T]) Len() int {
	return len(l.Data)
}
