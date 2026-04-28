package bencode

import (
	"testing"
)

func TestParserDict(t *testing.T) {
	data := []byte("d3:cow3:moo4:spam4:eggse")
	pos := 0

	val, _, err := Decode(data, pos)
	if err != nil {
		t.Fatal(err)
	}

	dict := val.Dict

	if string(dict["cow"].Str) != "moo" {
		t.Fatalf("wrong value for cow")
	}
}

func TestBadDict(t *testing.T) {
	data := []byte("d3:cow3:moo") // missing 'e' in the content
	pos := 0

	_, _, err := Decode(data, pos)
	if err == nil {
		t.Fatal("expected dict error")
	}
}

func TestDecode_DictValueDecodeError(t *testing.T) {
	data := []byte("d4:infoi-01ee")

	_, _, err := Decode(data, 0)
	if err == nil {
		t.Fatal("expected error in dict value decode")
	}
}

func TestDictKeyNotString(t *testing.T) {
	data := []byte("di42e3:fooee") // key = int
	pos := 0

	_, _, err := Decode(data, pos)
	if err == nil {
		t.Fatal("expected key not string")
	}
}

func TestDecode_DictKeyDecodeError(t *testing.T) {
	data := []byte("d4:inf") // broken key (unterminated string)

	_, _, err := Decode(data, 0)
	if err == nil {
		t.Fatal("expected error during key decode")
	}
}