package metadata

import "testing"

func FuzzGetMetadata(f *testing.F) {
	// Seed with valid inputs (VERY important)
	f.Add([]byte(
		"d4:infod" +
			"12:piece lengthi4e" +
			"6:pieces20:aaaaaaaaaaaaaaaaaaaa" +
			"4:name8:file.txt" +
			"6:lengthi100e" +
			"ee",
	))

	f.Add([]byte("i123e")) // invalid root
	f.Add([]byte("d3:foo3:bare")) // missing info

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic on input %q: %v", data, r)
			}
		}()

		_, _ = GetMetadata(data)
	})
}