// Package stack Contains the definition and implementation of a stack
// data structure.
package stack

type Stack[T any] struct {
	data []T
}

func New[T any](v ...T) Stack[T] {
	return Stack[T]{
		data: v,
	}
}

// Count Returns the amount of elements in the stack.
func (s Stack[T]) Count() int {
	return len(s.data)
}

// Empty Returns whether the stack has no element.
func (s Stack[T]) Empty() bool {
	return len(s.data) == 0
}

// Pop Returns there at the top of the stack and pops it.
// Returns the zero value of 'T' if the stack is empty.
func (s *Stack[T]) Pop() (value T) {
	if s.Empty() {
		return
	}

	end := len(s.data) - 1
	value = s.data[end]
	s.data = s.data[:end]
	return
}

// Push Adds a value to the top of the stack.
func (s *Stack[T]) Push(value T) {
	s.data = append(s.data, value)
}

// Top Returns the value at the top of the stack.
// returns the zero value of 'T' if the stack is empty.
func (s Stack[T]) Top() (v T) {
	if s.Empty() {
		return
	}

	return s.data[len(s.data)-1]
}
