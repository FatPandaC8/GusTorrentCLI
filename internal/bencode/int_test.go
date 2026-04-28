package bencode

import (
	"testing"
)

func TestParserInt(t *testing.T) {
	data := []byte("i42e")
	pos := 0

	val, _, err := Decode(data, pos)
	if err != nil {
		t.Fatal(err)
	}

	if val.Int != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}

func TestBadInt(t *testing.T) {
	data := []byte("i42") // missing 'e'
	pos := 0

	_, _, err := Decode(data, pos)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOverflowInt(t *testing.T) {
	data := []byte("i999999999999999999999999999999e")
	pos := 0

	_, _, err := Decode(data, pos)
	if err == nil {
		t.Fatal("expected error")
	}
}
