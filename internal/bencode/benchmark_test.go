package bencode

import (
	"bytes"
	"os"
	"testing"

	jack "github.com/jackpal/bencode-go"
)

func BenchmarkMyDecode(b *testing.B) {
	data, _ := os.ReadFile("../../torrent_files/ubuntu25.torrent")

	for b.Loop() {
		Decode(data, 0)
	}
}

func BenchmarkJackpalDecode(b *testing.B) {
	data, _ := os.ReadFile("../../torrent_files/ubuntu25.torrent")

	for b.Loop() {
		var v interface{}
		_ = jack.Unmarshal(bytes.NewReader(data), &v)
	}
}