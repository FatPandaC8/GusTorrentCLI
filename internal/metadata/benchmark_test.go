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

// [test] Used specially for benchmarking, don't use for production or anything else
func getMetadataNoHash(data []byte) (Info, error) {
	rootVal, _, err := bencode.Decode(data, 0)
	if err != nil {
		return Info{}, err
	}

	infoVal := rootVal.Dict["info"]

	info, err := parseInfo(infoVal)
	if err != nil {
		return Info{}, err
	}

	return info, nil
}
