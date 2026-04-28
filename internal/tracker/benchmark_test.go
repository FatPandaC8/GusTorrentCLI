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
		_, _ = GetPeers(data)
	}
}