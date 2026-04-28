package bencode

import (
	"testing"
)

func TestParserInt(t *testing.T) {
	data := []byte("i42e")

	val, _, err := Decode(data, 0)
	if err != nil {
		t.Fatal(err)
	}

	if val.Int != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}

func TestBadInt(t *testing.T) {
	data := []byte("i42") // missing 'e'

	_, _, err := Decode(data, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOverflowInt(t *testing.T) {
	data := []byte("i999999999999999999999999999999e")

	_, _, err := Decode(data, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}
