package stacks

import (
	"errors"
	"sync"
)

// ErrStackUnderflow is returned when attempting to Pop from an empty Stack.
var ErrStackUnderflow = errors.New("stack underflow")

// Stack is an abstract data type that serves as a collection of elements.
type Stack[T comparable] struct {
	elements []T
	sync.Mutex
}

// NewStack creates and returns an initialized stack.
func NewStack[T comparable]() Stack[T] {
	return Stack[T]{elements: []T{}}
}

// Pop pops the top element off the stack, it can return an error if the Stack
// is empty.
func (s *Stack[T]) Pop() (T, error) {
	s.Lock()
	defer s.Unlock()

	var result T
	if len(s.elements) == 0 {
		return result, ErrStackUnderflow
	}

	result = s.elements[len(s.elements)-1]
	s.elements = s.elements[:len(s.elements)-1]

	return result, nil
}

// Push pushes a new element onto the top of the Stack.
func (s *Stack[T]) Push(e T) {
	s.Lock()
	defer s.Unlock()

	s.elements = append(s.elements, e)
}

// Peek returns the top element on the Stack, without removing it.
//
// If the Stack is empty, nil is returned.
func (s *Stack[T]) Peek() T {
	s.Lock()
	defer s.Unlock()

	var zero T
	if len(s.elements) == 0 {
		return zero
	}

	return s.elements[len(s.elements)-1]
}
