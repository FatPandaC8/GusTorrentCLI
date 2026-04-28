package utils

import "testing"

func TestAtoi(t *testing.T) {
	val, _ := Atoi([]byte("123"))

	if val != 123 {
		t.Fail()
	}
}

func TestAtoi_Empty(t *testing.T) {
	_, err := Atoi([]byte{})
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestAtoi_OnlyMinus(t *testing.T) {
	_, err := Atoi([]byte("-"))
	if err == nil {
		t.Fatal("expected error for invalid number")
	}
}

func TestAtoi_LeadingZero(t *testing.T) {
	_, err := Atoi([]byte("01"))
	if err == nil {
		t.Fatal("expected error for leading zero")
	}
}

func TestAtoi_InvalidDigit(t *testing.T) {
	_, err := Atoi([]byte("12a3"))
	if err == nil {
		t.Fatal("expected error for invalid digit")
	}
}

func TestAtoi_Overflow(t *testing.T) {
	_, err := Atoi([]byte("999999999999999999999999999999"))
	if err == nil {
		t.Fatal("expected overflow error")
	}
}