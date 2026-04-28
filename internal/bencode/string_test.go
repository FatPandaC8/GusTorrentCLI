package bencode

import (
	"fmt"
	"testing"
)

func TestParserString(t *testing.T) {
	data := []byte("5:hello")

	val, _, err := Decode(data, 0)
	if err != nil {
		t.Fatal(err)
	}

	if string(val.Str) != "hello" {
		t.Fatalf("expected hello, got %v", val)
	}
}

func TestBadStringLength(t *testing.T) {
	data := []byte("999:abc")

	_, _, err := Decode(data, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBadStringColon(t *testing.T) {
	data := []byte("3abc")

	_, _, err := Decode(data, 0)
	if err == nil {
		t.Fatal("expected colon")
	}
}

func TestForeignLanguageString(t *testing.T) {
	str := "あいうえおか"
	data := fmt.Appendf(nil, "%d:%s", len([]byte(str)), str)

	val, _, err := Decode(data, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(val.Str) != str {
		t.Fatalf("got %q, want %q", val.Str, str)
	}
}
