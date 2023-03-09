package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewList(t *testing.T) {
	l := NewList[int]()
	assert.Equal(t, l.Data, []int{})
}

func TestListClear(t *testing.T) {
	l := NewList[int]()
	l.Insert(1)
	l.Insert(2)
	l.Insert(3)
	l.Clear()
	assert.Equal(t, l.Data, []int{})
}

func TestListContains(t *testing.T) {
	l := NewList[int]()
	l.Insert(1)
	l.Insert(2)
	l.Insert(3)
	assert.Equal(t, l.Contains(1), true)
	assert.Equal(t, l.Contains(4), false)
}

func TestListGetIndex(t *testing.T) {
	l := NewList[int]()
	l.Insert(1)
	l.Insert(2)
	l.Insert(3)
	assert.Equal(t, l.GetIndex(1), 0)
	assert.Equal(t, l.GetIndex(4), -1)
}

func TestListInsert(t *testing.T) {
	l := NewList[int]()
	l.Insert(1)
	l.Insert(2)
	l.Insert(3)
	assert.Equal(t, l.Data, []int{1, 2, 3})
}

func TestListRemove(t *testing.T) {
	l := NewList[int]()
	l.Insert(1)
	l.Insert(2)
	l.Insert(3)
	l.Remove(2)
	assert.Equal(t, l.Data, []int{1, 3})
}

func TestRemoveAt(t *testing.T) {
	l := NewList[int]()
	l.Insert(1)
	l.Insert(2)
	l.Insert(3)
	l.Pop(1)
	assert.Equal(t, l.Data, []int{1, 3})
}

func TestListLast(t *testing.T) {
	l := NewList[int]()
	l.Insert(1)
	l.Insert(2)
	l.Insert(3)
	assert.Equal(t, l.Last(), 3)
}