package stack

import "testing"

func TestStack_PushPop(t *testing.T) {
	s := NewStack(2)

	s.Push(1)
	s.Push(2)

	if s.Size() != 2 {
		t.Fatalf("expected size 2, got %d", s.Size())
	}

	v, err := s.Pop()
	if err != nil || v != 2 {
		t.Fatalf("expected 2, got %v (err=%v)", v, err)
	}

	v, err = s.Pop()
	if err != nil || v != 1 {
		t.Fatalf("expected 1, got %v (err=%v)", v, err)
	}

	if !s.IsEmpty() {
		t.Fatalf("expected empty stack")
	}
}

func TestStack_Peek(t *testing.T) {
	s := NewStack(2)

	s.Push(10)
	s.Push(20)

	v, err := s.Peek()
	if err != nil || v != 20 {
		t.Fatalf("expected 20, got %v", v)
	}

	if s.Size() != 2 {
		t.Fatalf("peek should not remove element")
	}
}

func TestStack_EmptyErrors(t *testing.T) {
	s := NewStack(1)

	if _, err := s.Pop(); err == nil {
		t.Fatalf("expected error on Pop from empty stack")
	}

	if _, err := s.Peek(); err == nil {
		t.Fatalf("expected error on Peek from empty stack")
	}
}

func TestStack_MixedTypes(t *testing.T) {
	s := NewStack(2)

	s.Push(42)
	s.Push("hello")

	v, _ := s.Pop()
	if v != "hello" {
		t.Fatalf("expected 'hello', got %v", v)
	}

	v, _ = s.Pop()
	if v != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestStack_NilValues(t *testing.T) {
	s := NewStack(2)

	s.Push(nil)

	v, err := s.Pop()
	if err != nil {
		t.Fatal(err)
	}

	if v != nil {
		t.Fatalf("expected nil, got %v", v)
	}
}

func TestStack_CapacityGrowth(t *testing.T) {
	s := NewStack(1)

	for i := 0; i < 100; i++ {
		s.Push(i)
	}

	if s.Size() != 100 {
		t.Fatalf("expected size 100, got %d", s.Size())
	}
}

func TestStack_PeekIdempotent(t *testing.T) {
	s := NewStack(2)

	s.Push(1)

	for i := 0; i < 10; i++ {
		v, _ := s.Peek()
		if v != 1 {
			t.Fatalf("peek changed value")
		}
	}

	if s.Size() != 1 {
		t.Fatalf("peek should not modify stack")
	}
}

func TestStack_ClearReference(t *testing.T) {
	s := NewStack(1)

	type Big struct {
		data [1024 * 1024]byte // 1MB
	}

	s.Push(&Big{})
	s.Pop()

	// Not a strict assertion test, but ensures no panic / misuse
}