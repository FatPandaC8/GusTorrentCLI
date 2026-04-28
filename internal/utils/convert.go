package utils

import "errors"

// [HELPER] Default is base 10
// NOTE: need to implement int overflow from input
// uint8  : 0 to 255 
// uint16 : 0 to 65535 
// uint32 : 0 to 4294967295 
// uint64 : 0 to 18446744073709551615 
// int8   : -128 to 127 
// int16  : -32768 to 32767 
// int32  : -2147483648 to 2147483647 
// int64  : -9223372036854775808 to 9223372036854775807

const (
	MaxUint = ^uint(0)
	MinUint = 0
	MaxInt = int(MaxUint >> 1)
	MinInt = -MaxInt - 1
)

func Atoi(s []byte) (int, error) {
	if len(s) == 0 {
		return 0, errors.New("empty string")
	}

	i := 0
	sign := 1

	// handle sign
	if s[0] == '-' {
		sign = -1
		i++
		if i == len(s) {
			return 0, errors.New("invalid number")
		}
	}

	// leading zero check (bencode strict)
	if s[i] == '0' && i + 1 < len(s) {
		return 0, errors.New("leading zero")
	}

	n := 0

	for ; i < len(s); i++ {
		ch := s[i]
		if ch < '0' || ch > '9' {
			return 0, errors.New("invalid digit")
		}

		digit := int(ch - '0')

		// ensure the next result can be held: n * 10 + digit <= MaxInt
		if n > (MaxInt - digit) / 10 {
			return 0, errors.New("overflow")
		}

		n = (n << 3) + (n << 1) + digit
	}

	return sign * n, nil
}