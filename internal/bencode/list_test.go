package bencode

import (
	"testing"
)

func TestParserList(t *testing.T) {
	data := []byte("li1ei2ei3ee")

	val, _, err := Decode(data, 0)
	if err != nil {
		t.Fatal(err)
	}

	list := val.List

	if len(list) != 3 || list[0].Int != 1 {
		t.Fatalf("expected 3 elements")
	}
}

func TestBadList(t *testing.T) {
	data := []byte("li1ei2e") // more final 'e'

	_, _, err := Decode(data, 0)
	if err == nil {
		t.Fatal("expected list error")
	}
}

func TestDecode_ListValueDecodeError(t *testing.T) {
	data := []byte("li-01ee")

	_, _, err := Decode(data, 0)
	if err == nil {
		t.Fatal("expected error in list decode")
	}
}
