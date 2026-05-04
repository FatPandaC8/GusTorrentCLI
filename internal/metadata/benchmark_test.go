package metadata

import (
	"crypto/sha1"
	"gustorrent/internal/bencode"
	"os"
	"testing"
)

func BenchmarkMetadata(b *testing.B) {
	data, _ := os.ReadFile("../../torrent_files/ubuntu25.torrent")

	for b.Loop() {
		GetMetadata(data)
	}
}

func BenchmarkMetadata_Hash(b *testing.B) {
	data, _ := os.ReadFile("../../torrent_files/ubuntu25.torrent")

	rootVal, _, _ := bencode.Decode(data, 0)
	infoVal := rootVal.Dict["info"]

	raw := data[infoVal.Start:infoVal.End]

	for b.Loop() {
		_ = sha1.Sum(raw)
	}
}

func BenchmarkMetadataNoHash(b *testing.B) {
	data, _ := os.ReadFile("../../torrent_files/ubuntu25.torrent")

	for b.Loop() {
		_, _ = getMetadataNoHash(data)
	}
}
