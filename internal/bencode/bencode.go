package bencode

import (
	"bytes"
	"fmt"
	"gustorrent/internal/utils"
)

type Value struct {
	Type rune // 'i', 's', 'l', 'd'
	Int  int
	Str  []byte
	List []Value
	Dict map[string]Value
}

func Decode(data []byte, pos int) (Value, int, error) {
	if pos >= len(data) {
		return Value{}, 0, fmt.Errorf("unexpected eof")
	}

	switch data[pos] {
	case 'i':
		end := bytes.IndexByte(data[pos:], 'e')
		if end == -1 {
			return Value{}, 0, fmt.Errorf("unterminated int")
		}
		end += pos
		n, err := utils.Atoi(data[pos+1 : end])
		if err != nil {
			return Value{}, 0, err
		}
		return Value{Type: 'i', Int: n}, end + 1 - pos, nil

	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		colon := bytes.IndexByte(data[pos:], ':')
		if colon == -1 {
			return Value{}, 0, fmt.Errorf("missing colon")
		}
		colon += pos
		length, err := utils.Atoi(data[pos:colon])
		if err != nil {
			return Value{}, 0, err
		}
		start := colon + 1
		end := start + length
		if end > len(data) {
			return Value{}, 0, fmt.Errorf("string out of bounds")
		}
		return Value{Type: 's', Str: data[start:end]}, end - pos, nil

	case 'l':
		var items []Value
		cur := pos + 1 // skip 'l'
		for cur < len(data) && data[cur] != 'e' {
			v, n, err := Decode(data, cur)
			if err != nil {
				return Value{}, 0, err
			}
			items = append(items, v)
			cur += n
		}
		if cur >= len(data) {
			return Value{}, 0, fmt.Errorf("unterminated list")
		}
		return Value{Type: 'l', List: items}, cur + 1 - pos, nil // +1 for 'e'

	case 'd':
		dict := map[string]Value{}
		cur := pos + 1 // skip 'd'
		for cur < len(data) && data[cur] != 'e' {
			key, n, err := Decode(data, cur)
			if err != nil {
				return Value{}, 0, err
			}
			if key.Type != 's' {
				return Value{}, 0, fmt.Errorf("dict key must be a string")
			}
			cur += n
			val, n, err := Decode(data, cur)
			if err != nil {
				return Value{}, 0, err
			}
			dict[string(key.Str)] = val
			cur += n
		}
		if cur >= len(data) {
			return Value{}, 0, fmt.Errorf("unterminated dict")
		}
		return Value{Type: 'd', Dict: dict}, cur + 1 - pos, nil // +1 for 'e'
	}

	return Value{}, 0, fmt.Errorf("unexpected byte: %c", data[pos])
}