package parser

import "testing"

func FuzzParse(f *testing.F) {
	// seed with valid examples
	f.Add([]byte("i42e")) // int
	f.Add([]byte("5:hello")) // string
	f.Add([]byte("li1ei2ee")) // list
	f.Add([]byte("d3:cow3:moo4:spam4:eggse")) // dict

	f.Fuzz(func(t *testing.T, data []byte) {
		pos := 0

		_, err := Parse(data, &pos)

		// Example invariant: parser should never go out of bounds
		if pos > len(data) {
			t.Fatalf("pos out of bounds: %d > %d", pos, len(data))
		}

		// Optional: reject weird behavior
		if err == nil && len(data) == 0 {
			t.Fatalf("empty input should probably fail")
		}
	})
}