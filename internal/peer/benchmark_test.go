package peer

import (
	"encoding/binary"
	"net"
	"testing"
)

func BenchmarkReadMessage(b *testing.B) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	go func() {
		for {
			buf := make([]byte, 4)
			binary.BigEndian.PutUint32(buf, 1)
			server.Write(buf)
			server.Write([]byte{1})
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ReadMessage(client)
	}
}

func BenchmarkDownloadPiece(b *testing.B) {
	for i := 0; i < b.N; i++ {
		server, client := net.Pipe()

		go func() {
			for {
				req := make([]byte, 17)
				_, err := server.Read(req)
				if err != nil {
					return
				}

				payload := make([]byte, 8+16384)

				msgLen := uint32(1 + len(payload))
				lenBuf := make([]byte, 4)
				binary.BigEndian.PutUint32(lenBuf, msgLen)

				server.Write(lenBuf)
				server.Write([]byte{7})
				server.Write(payload)
			}
		}()

		DownloadPiece(client, 0, 16384)

		server.Close()
		client.Close()
	}
}