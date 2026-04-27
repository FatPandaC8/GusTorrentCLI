package parser

import (
	"fmt"
	"gustorrent/internal/dsa/stack"
	"strconv"
)

type ListFrame struct {
	data []any
}

type DictFrame struct {
	data        map[string]any
	expectValue bool
	key         string
}

// parse must move the globally
func Parse(data []byte) (any, error) {
	stack := stack.NewStack(32)
	var current any
	var hasValue bool
	dataSize := len(data)

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
				return nil, fmt.Errorf("unterminated integer")
			}

			n, err := strconv.Atoi(string(data[start:end]))
			if err != nil {
				return nil, err
			}
			current = n
			i = end
			hasValue = true

		case ch >= '0' && ch <= '9':
			start := i
			end := start

			for end < len(data) && data[end] != ':' {
				end++
			}

			if end >= len(data) {
				return nil, fmt.Errorf("unterminated string length")
			}

			length, err := strconv.Atoi(string(data[start:end]))
			if err != nil {
				return nil, err
			}

			dataStart := end + 1
			dataEnd := dataStart + length

			if dataEnd > len(data) {
				return nil, fmt.Errorf("string out of bounds")
			}

			current = data[dataStart:dataEnd]
			i = dataEnd - 1 // adjust for loop increment
			hasValue = true

		case ch == 'l':
			stack.Push(&ListFrame{data: []any{}}) // push the type into the stack
			hasValue = false

		case ch == 'd':
			stack.Push(&DictFrame{
				data: make(map[string]any),
			})
			hasValue = false

		case ch == 'e':
			val, err := stack.Pop()
			if err != nil {
				return nil, err
			}
			switch v := val.(type) {
			case *ListFrame:
				current = v.data
			case *DictFrame:
				current = v.data
			}
			hasValue = true

		default:
			return nil, fmt.Errorf("invalid character: %c", ch)
		}

		if hasValue {
			if !stack.IsEmpty() {
				top, _ := stack.Peek()
	
				switch frame := top.(type) {
	
				case *ListFrame:
					frame.data = append(frame.data, current)
	
				case *DictFrame:
					if !frame.expectValue {
						keyBytes, ok := current.([]byte)
						if !ok {
							return nil, fmt.Errorf("dict key must be string")
						}
						frame.key = string(keyBytes)
						frame.expectValue = true
					} else {
						frame.data[frame.key] = current
						frame.expectValue = false
					}
				}
			} else {
				return current, nil
			}
		}
	}

	return nil, fmt.Errorf("unexpected end of input")
}

// USAGE:
// Parse(byte, position)
// byte is the file byte, position is global position for recursion
