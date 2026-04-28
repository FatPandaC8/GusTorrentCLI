package bencode

import "testing"

func FuzzParse(f *testing.F) {

	f.Add([]byte("i42e"))
	f.Add([]byte("i-1e"))
	f.Add([]byte("0:"))
	f.Add([]byte("5:hello"))
	f.Add([]byte("li1ei2ee"))
	f.Add([]byte("d3:cow3:moo4:spam4:eggse"))

	f.Add([]byte("i0e"))
	f.Add([]byte("i-0e"))
	f.Add([]byte("1:a"))
	f.Add([]byte("9999999999:abc"))

	f.Add([]byte("i42"))
	f.Add([]byte("3:ab"))
	f.Add([]byte("d3:keyi1ee"))
	f.Add([]byte("d3:keyi1e"))
	f.Add([]byte("l"))
	f.Add([]byte("d"))
	f.Add([]byte(""))
	f.Add([]byte{0x00, 0xff, 10})

	f.Fuzz(func(t *testing.T, data []byte) {

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic on input %q: %v", data, r)
			}
		}()

		val, _, err := Decode(data, 0)

		if err == nil {
			switch val.Type {
			case 'l':
				for _, v := range val.List {
					if v.Type == 0 {
						t.Fatalf("invalid list element")
					}
				}
			case 'd':
				for k := range val.Dict {
					if k == "" {
						t.Fatalf("empty dict key suspicious")
					}
				}
			}
		}
	})
}
