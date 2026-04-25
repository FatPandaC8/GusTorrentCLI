package parser

import ( 
	"testing"
)

func TestParserString(t *testing.T) {
	data := []byte("5:hello")
	pos := 0

	val, err := Parse(data, &pos)
	if err != nil {
		t.Fatal(err)
	}

	if string(val.([]byte)) != "hello" {
		t.Fatalf("expected hello, got %s", val)
	}
}

func TestBadStringLength(t *testing.T) {
	data := []byte("999:abc")
	pos := 0

	_, err := Parse(data, &pos)
	if err == nil {
		t.Fatal("expected error")
	}
}