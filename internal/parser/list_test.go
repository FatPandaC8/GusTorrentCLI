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

	list := val.([]interface{})
	if len(list) != 3 {
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