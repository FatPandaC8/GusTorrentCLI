package bencode

import (
	"os"
	"testing"
)

func BenchmarkParse(b *testing.B) {
	data, _ := os.ReadFile("../../torrent_files/ubuntu25.torrent")

	for b.Loop() {
		Decode(data, 0)
	}
}
