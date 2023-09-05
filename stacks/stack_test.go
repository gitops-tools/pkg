package stacks

import (
	"testing"
)

func TestStack_Pop_empty(t *testing.T) {
	s := NewStack[int]()

	_, err := s.Pop()

	if err != ErrStackUnderflow {
		t.Fatalf("got %#v, want stack underflow error", err)
	}
}

func TestStack_Push(t *testing.T) {
	s := NewStack[string]()

	s.Push("testing")

	v, err := s.Pop()
	if err != nil {
		t.Fatal(err)
	}

	if v != "testing" {
		t.Fatalf("got %#v, want %q", v, "testing")
	}
}

func TestStack_Peek_empty(t *testing.T) {
	s := NewStack[float32]()

	v := s.Peek()

	if v != 0.0 {
		t.Fatalf("got %#v, want 0.0", v)
	}
}

func TestStack_Peek(t *testing.T) {
	s := NewStack[string]()
	s.Push("test1")
	s.Push("test2")

	v := s.Peek()

	if v != "test2" {
		t.Fatalf("got %#v, want %q", v, "test2")
	}
	_, _ = s.Pop()
	v = s.Peek()
	if v != "test1" {
		t.Fatalf("got %#v, want %q", v, "test2")
	}

}
