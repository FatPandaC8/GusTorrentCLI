package metadata

import (
	"os"
	"testing"
)

func BenchmarkMetadata(b *testing.B) {
	data, _ := os.ReadFile("../../torrent_files/ubuntu25.torrent")
	
	for b.Loop() {
		GetMetadata(data)
	}
}