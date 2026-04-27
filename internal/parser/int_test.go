package parser

import ( 
	"testing"
)

func TestParserInt(t *testing.T) {
	data := []byte("i42e")

	val, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	if val != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}

func TestBadInt(t *testing.T) {
	data := []byte("i42") // missing 'e'

	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error")
	}
}