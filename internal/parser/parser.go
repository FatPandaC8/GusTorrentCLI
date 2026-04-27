package parser

import "fmt"

type ValueType byte

const (
	TypeInt  ValueType = 'i'
	TypeStr  ValueType = 's'
	TypeList ValueType = 'l'
	TypeDict ValueType = 'd'
)

type Value struct {
	t ValueType

	i int
	s []byte
	l []Value
	d map[string]Value
}

type Frame struct {
	typ ValueType

	list []Value

	dict map[string]Value
	key  string
	expectValue bool
}

func parseInt(data []byte, start, end int) (int, error) {
	if start >= end {
		return 0, fmt.Errorf("empty int")
	}

	n := 0
	sign := 1

	if data[start] == '-' {
		sign = -1
		start++
	}

	for i := start; i < end; i++ {
		b := data[i]
		if b < '0' || b > '9' {
			return 0, fmt.Errorf("invalid number")
		}
		n = n*10 + int(b-'0')
	}

	return sign * n, nil
}

func Parse(data []byte) (Value, error) {
	var stack []Frame
	dataSize := len(data)

	var current Value
	var hasValue bool

	for i := 0; i < dataSize; i++ {
		ch := data[i]
		hasValue = false

		switch {
		case ch == 'i':
			start := i + 1
			end := start

			for end < dataSize && data[end] != 'e' {
				end++
			}
			if end >= dataSize {
				return Value{}, fmt.Errorf("unterminated int")
			}

			n, err := parseInt(data, start, end)
			if err != nil {
				return Value{}, err
			}

			current = Value{t: TypeInt, i: n}
			i = end
			hasValue = true

		case ch >= '0' && ch <= '9':
			start := i
			end := start

			for end < dataSize && data[end] != ':' {
				end++
			}
			if end >= dataSize {
				return Value{}, fmt.Errorf("bad string len")
			}

			length, err := parseInt(data, start, end)
			if err != nil {
				return Value{}, err
			}

			dataStart := end + 1
			dataEnd := dataStart + length

			if dataEnd > dataSize {
				return Value{}, fmt.Errorf("string OOB")
			}

			current = Value{
				t: TypeStr,
				s: data[dataStart:dataEnd], // zero-copy
			}

			i = dataEnd - 1
			hasValue = true

		case ch == 'l':
			stack = append(stack, Frame{
				typ:  TypeList,
				list: make([]Value, 0, 8),
			})

		case ch == 'd':
			stack = append(stack, Frame{
				typ:  TypeDict,
				dict: make(map[string]Value, 8),
			})

		case ch == 'e':
			if len(stack) == 0 {
				return Value{}, fmt.Errorf("unexpected end")
			}

			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if top.typ == TypeList {
				current = Value{t: TypeList, l: top.list}
			} else {
				current = Value{t: TypeDict, d: top.dict}
			}

			hasValue = true

		default:
			return Value{}, fmt.Errorf("invalid char: %c", ch)
		}

		if hasValue {
			if len(stack) == 0 {
				return current, nil
			}

			top := &stack[len(stack)-1]

			if top.typ == TypeList {
				top.list = append(top.list, current)
			} else {
				if !top.expectValue {
					if current.t != TypeStr {
						return Value{}, fmt.Errorf("dict key must be string")
					}
					top.key = string(current.s) // unavoidable alloc
					top.expectValue = true
				} else {
					top.dict[top.key] = current
					top.expectValue = false
				}
			}
		}
	}

	return Value{}, fmt.Errorf("unexpected EOF")
}