package parser

import ( 
	"testing"
)

func TestParserDict(t *testing.T) {
	data := []byte("d3:cow3:moo4:spam4:eggse")

	val, err := Parse(data)
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

	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDictKeyNotString(t *testing.T) {
	data := []byte("di42e3:fooee") // key = int

	_, err := Parse(data)
	if err == nil {
		t.Fatal("expected error")
	}
}