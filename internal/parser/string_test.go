package parser

import ( 
	"testing"
)

func TestParserString(t *testing.T) {
	data := []byte("5:hello")

	val, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	if val.t != TypeStr || string(val.s) != "hello" {
		t.Fatalf("expected hello, got %v", val)
	}
}

func TestBadStringLength(t *testing.T) {
	data := []byte("999:abc")

	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error")
	}
}