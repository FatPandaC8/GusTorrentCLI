package parser

import (
	"testing"
)

func TestUnknownType(t *testing.T) {
	data := []byte("x")
	pos := 0

	_, err := Parse(data, &pos)
	if err == nil {
		t.Fatal("expected unknown type error")
	}
}

func TestOutOfBound(t *testing.T) {
	data := []byte("i42e")
	pos := 100

	_, err := Parse(data, &pos)
	if err == nil {
		t.Fatal("expected out of bound error")
	}
}