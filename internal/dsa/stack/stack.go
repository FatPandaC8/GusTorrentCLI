package stack

import "errors"

type Stack struct {
	elements []any
}

func NewStack(initialSize int) *Stack {
	return &Stack{
		elements: make([]any, 0, initialSize), // make an slice with full 0 with initialSize 
	}
}

func (s *Stack) Push(v any) {
	s.elements = append(s.elements, v)
}

func (s *Stack) Pop() (any, error) {
	size := len(s.elements) // NOTE: dont use s.Size() to optimize func call (less jumping)

	if size == 0 {
		return nil, errors.New("stack empty")
	}

	top := s.elements[size - 1]
	s.elements[size - 1] = nil // Just resizing wont remove the actual value of the elements -> memory leak
	s.elements = s.elements[:size - 1] // resize the slices

	return top, nil
}

func (s *Stack) IsEmpty() bool {
	return len(s.elements) == 0
}

func (s *Stack) Size() int {
	return len(s.elements)
}

func (s *Stack) Peek() (any, error) {
	size := len(s.elements)

	if size == 0 {
		return nil, errors.New("stack empty")
	}

	return s.elements[size - 1], nil
}