package parser

import (
	"testing"
)

func TestParserList(t *testing.T) {
	data := []byte("li1ei2ei3ee")

	val, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	if val.t != TypeList {
		t.Fail()
	}

	list := val.l

	if len(list) != 3 || list[0].i != 1 {
		t.Fatalf("expected 3 elements")
	}
}

func TestBadList(t *testing.T) {
	data := []byte("li1ei2e") // more final 'e'

	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error")
	}
}