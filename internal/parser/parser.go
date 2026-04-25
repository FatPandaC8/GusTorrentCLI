package parser

import (
	"fmt"
	"strconv"
)

// parse must move the pos globally
func Parse(data []byte, pos *int) (interface{}, error) {
	if *pos >= len(data) {
    	return nil, fmt.Errorf("out of bounds at pos=%d", *pos)
	}
	
	ch := data[*pos]

	if ch == 'i' {
		start := *pos + 1
        end := start

        for end < len(data) && data[end] != 'e' {
            end++
        }

        if end >= len(data) {
            return nil, fmt.Errorf("unterminated integer in int case")
        }

        n, err := strconv.Atoi(string(data[start:end]))

        if err != nil {
            return nil, fmt.Errorf("unterminated integer in int case")
        }

        *pos = end + 1

        return n, nil
	}

	if ch >= '0' && ch <= '9' {
		// number is current pos -> :
		start := *pos
		end := start

		for end < len(data) && data[end] != ':' {
			end++
		}

		if end >= len(data) {
            return nil, fmt.Errorf("unterminated integer in string byte case")
        }

		number, err := strconv.Atoi(string(data[start:end]))
		if err != nil {
			fmt.Println("ERROR:", err)
			return nil, err
		}

		dataStart := end + 1
		dataEnd := dataStart + number

		if dataEnd > len(data) {
			return nil, fmt.Errorf("string length %d out of bounds", number)
		}

		*pos = dataEnd

		// read that number of byte from pos -> pos + number
		return (data[dataStart: dataEnd]), nil
	}

	if ch == 'l' {
		// make an empty list
		var list []interface{}

		// recursively parse and append the value of the parse to the list
		*pos = *pos + 1 // skip l

		for {
			if *pos >= len(data) {
				return nil, fmt.Errorf("unterminated list")
			}

			if data[*pos] == 'e' {
				*pos = *pos + 1
				break
			}

			val, err := Parse(data, pos)
			if err != nil {
				return nil, err
			}

			list = append(list, val)
		}

		return list, nil
	}

	if ch == 'd' {
		dict := make(map[string]interface{})

		*pos = *pos + 1 // skip d
		for {
			if *pos >= len(data) {
				return nil, fmt.Errorf("unterminated dict")
			}

			if data[*pos] == 'e' {
				*pos = *pos + 1
				break
			}

			keyVal, err := Parse(data, pos)
			if err != nil {
				return nil, err
			}

			keyBytes, ok := keyVal.([]byte)
			if !ok {
				return nil, fmt.Errorf("dict key must be string")
			}

			key := string(keyBytes)

			value, err := Parse(data, pos)
			if err != nil {
				return nil, err
			}

			dict[key] = value
		}

		return dict, nil
	}

	return nil, fmt.Errorf("unknown type: %c", ch)
}

// USAGE:
// Parse(byte, position)
// byte is the file byte, position is global position for recursion