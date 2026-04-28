package bencode

import (
	"testing"
)

func TestUnknownType(t *testing.T) {
	data := []byte("x")

	_, _, err := Decode(data, 0)
	if err == nil {
		t.Fatal("expected unknown type error")
	}
}
