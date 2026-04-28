package bencode

import (
	"testing"
)

func TestUnknownType(t *testing.T) {
	data := []byte("x")
	pos := 0

	_, _, err := Decode(data, pos)
	if err == nil {
		t.Fatal("expected unknown type error")
	}
}
