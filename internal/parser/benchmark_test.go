package parser

import (
	"os"
	"testing"
)

func BenchmarkParse(b *testing.B) {
	data, _ := os.ReadFile("../../torrent_files/ubuntu25.torrent")

	for i := 0; i < b.N; i++ {
		p := 0 // reset position every run
		Parse(data, &p)
	}
}