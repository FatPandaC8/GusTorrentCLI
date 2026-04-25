package parser

import ( 
	"testing"
)

func TestParserInt(t *testing.T) {
	data := []byte("i42e")
	pos := 0

	val, err := Parse(data, &pos)
	if err != nil {
		t.Fatal(err)
	}

	if val != 42 {
		t.Fatalf("expected 42, got %v", val)
	}

	if pos != 4 {
		t.Fatalf("expected pos=4, got %d", pos)
	}
}

func TestBadInt(t *testing.T) {
	data := []byte("i42") // missing 'e'
	pos := 0

	_, err := Parse(data, &pos)
	if err == nil {
		t.Fatal("expected error")
	}
}