package tracker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkDecodePeers(b *testing.B) {
	// 1000 peers -> realistic load
	body := make([]byte, 6*1000)

	for i := 0; i < len(body); i += 6 {
		body[i] = 127
		body[i+1] = 0
		body[i+2] = 0
		body[i+3] = 1
		body[i+4] = 0x1A
		body[i+5] = 0xE1
	}

	for b.Loop() {
		_ = DecodePeers(body)
	}
}

func BenchmarkBuildTrackerURL(b *testing.B) {
	var hash [20]byte
	copy(hash[:], "12345678901234567890")

	for b.Loop() {
		_ = BuildTrackerURL("http://tracker", hash, 1000)
	}
}

func BenchmarkGetPeers(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 50 peers
		peers := make([]byte, 6*50)
		for i := 0; i < len(peers); i += 6 {
			peers[i] = 127
			peers[i+1] = 0
			peers[i+2] = 0
			peers[i+3] = 1
			peers[i+4] = 0x1A
			peers[i+5] = 0xE1
		}

		resp := append([]byte("d5:peers"), []byte(fmt.Sprintf("%d:", len(peers)))...)
		resp = append(resp, peers...)
		resp = append(resp, 'e')

		w.Write(resp)
	}))
	defer server.Close()

	data := []byte(
		"d" +
			"8:announce" + fmt.Sprintf("%d:%s", len(server.URL), server.URL) +
			"4:infod" +
			"12:piece lengthi4e" +
			"6:pieces20:aaaaaaaaaaaaaaaaaaaa" +
			"4:name8:file.txt" +
			"6:lengthi100e" +
			"ee",
	)

	for b.Loop() {
		_, _ = GetPeers(standardClient, data)
	}
}

func BenchmarkBuildHandshake(b *testing.B) {
	data := make([]byte, 1024) // fake torrent data
	peerID := "-GT0001-123456789012"

	for i := 0; i < b.N; i++ {
		BuildHandshake(data, peerID)
	}
}

func BenchmarkBuildKeepAlive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildKeepAlive()
	}
}

func BenchmarkBuildRequest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildRequest(5, 16384, 16384)
	}
}

func BenchmarkBuildChoke(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildChoke()
	}
}

func BenchmarkBuildUnchoke(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildUnchoke()
	}
}

func BenchmarkBuildInterested(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildInterested()
	}
}

func BenchmarkBuildUninterested(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildUninterested()
	}
}

func BenchmarkBuildPiece(b *testing.B) {
	block := make([]byte, 16*1024) // 16KB

	for i := 0; i < b.N; i++ {
		BuildPiece(1, 0, block)
	}
}

func BenchmarkBuildHave(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildHave(12345)
	}
}

func BenchmarkBuildBitfield_Small(b *testing.B) {
	bf := make(Bitfield, 8) // 64 pieces

	for i := 0; i < b.N; i++ {
		BuildBitfield(bf)
	}
}

func BenchmarkBuildBitfield_Medium(b *testing.B) {
	bf := make(Bitfield, 128) // ~1000 pieces

	for i := 0; i < b.N; i++ {
		BuildBitfield(bf)
	}
}

func BenchmarkBuildBitfield_Large(b *testing.B) {
	bf := make(Bitfield, 1024) // large torrent

	for i := 0; i < b.N; i++ {
		BuildBitfield(bf)
	}
}

func BenchmarkBitfieldSet(b *testing.B) {
	bf := make(Bitfield, 128)

	for i := 0; i < b.N; i++ {
		bf.Set(i % 1000)
	}
}

func BenchmarkBitfieldHas(b *testing.B) {
	bf := make(Bitfield, 128)

	for i := 0; i < 1000; i++ {
		bf.Set(i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bf.Has(i % 1000)
	}
}

func BenchmarkBuildPiece_1KB(b *testing.B) {
	block := make([]byte, 1024)

	for i := 0; i < b.N; i++ {
		BuildPiece(1, 0, block)
	}
}

func BenchmarkBuildPiece_16KB(b *testing.B) {
	block := make([]byte, 16*1024)

	for i := 0; i < b.N; i++ {
		BuildPiece(1, 0, block)
	}
}

func BenchmarkBuildPiece_64KB(b *testing.B) {
	block := make([]byte, 64*1024)

	for i := 0; i < b.N; i++ {
		BuildPiece(1, 0, block)
	}
}

func BenchmarkBuildCancel(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildCancel(uint32(i%1000), 0, 16384)
	}
}

func BenchmarkBuildPort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildPort(uint16(6881 + i%100))
	}
}
