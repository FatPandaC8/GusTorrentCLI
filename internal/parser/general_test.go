package parser

import (
	"testing"
)

func TestUnknownType(t *testing.T) {
	data := []byte("x")

	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected unknown type error")
	}
}