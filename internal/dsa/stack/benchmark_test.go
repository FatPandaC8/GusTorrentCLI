package stack

import "testing"

func BenchmarkStack_Push(b *testing.B) {
	for b.Loop() {
		s := NewStack(0)
		for j := 0; j < 1000; j++ {
			s.Push(j)
		}
	}
}

func BenchmarkStack_Pop(b *testing.B) {
	for b.Loop() {
		s := NewStack(1000)
		for j := range 1000 {
			s.Push(j)
		}
		for range 1000 {
			s.Pop()
		}
	}
}

func BenchmarkStack_PushPop(b *testing.B) {
	s := NewStack(0)

	for i := 0; b.Loop(); i++ {
		s.Push(i)
		s.Pop()
	}
}

func BenchmarkStack_Peek(b *testing.B) {
	s := NewStack(1000)
	for i := range 1000 {
		s.Push(i)
	}

	for b.Loop() {
		s.Peek()
	}
}