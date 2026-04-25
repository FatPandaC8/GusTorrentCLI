package parser

import ( 
	"testing"
)

func TestParserDict(t *testing.T) {
	data := []byte("d3:cow3:moo4:spam4:eggse")
	pos := 0

	val, err := Parse(data, &pos)
	if err != nil {
		t.Fatal(err)
	}

	dict := val.(map[string]interface{})

	if string(dict["cow"].([]byte)) != "moo" {
		t.Fatalf("wrong value for cow")
	}
}

func TestBadDict(t *testing.T) {
	data := []byte("d3:cow3:moo") // missing 'e'
	pos := 0

	_, err := Parse(data, &pos)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDictKeyNotString(t *testing.T) {
	data := []byte("di42e3:fooee") // key = int
	pos := 0

	_, err := Parse(data, &pos)
	if err == nil {
		t.Fatal("expected error")
	}
}