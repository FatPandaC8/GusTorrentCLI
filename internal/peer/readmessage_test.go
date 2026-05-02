package peer

import (
	"encoding/binary"
	"net"
	"testing"
)

func TestReadMessage_Normal(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	go func() {
		// length = 1 (id only)
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, 1)

		server.Write(buf)
		server.Write([]byte{1}) // id = unchoke
	}()

	id, payload, err := ReadMessage(client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 1 {
		t.Fatalf("expected id=1, got %d", id)
	}

	if len(payload) != 0 {
		t.Fatalf("expected empty payload")
	}
}

func TestReadMessage_KeepAlive(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	go func() {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, 0)
		server.Write(buf)
	}()

	id, payload, err := ReadMessage(client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 255 {
		t.Fatalf("expected keep-alive id=255, got %d", id)
	}

	if payload != nil {
		t.Fatalf("expected nil payload")
	}
}

func TestReadMessage_SplitPackets(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	go func() {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, 5)

		// send in chunks
		server.Write(buf[:2])
		server.Write(buf[2:])

		server.Write([]byte{7, 0, 0, 0, 0}) // id + payload
	}()

	id, payload, err := ReadMessage(client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 7 {
		t.Fatalf("expected id=7, got %d", id)
	}

	if len(payload) != 4 {
		t.Fatalf("expected payload size 4, got %d", len(payload))
	}
}

func TestDownloadPiece_SingleBlock(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	pieceIndex := 0
	pieceLength := 8

	expected := []byte("abcdefgh")

	go func() {
		// read request (ignore content for now)
		req := make([]byte, 17)
		server.Read(req)

		// build piece message
		payload := make([]byte, 8+len(expected))

		binary.BigEndian.PutUint32(payload[0:4], uint32(pieceIndex)) // index
		binary.BigEndian.PutUint32(payload[4:8], 0)                  // begin
		copy(payload[8:], expected)

		msgLen := uint32(1 + len(payload))

		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, msgLen)

		server.Write(lenBuf)
		server.Write([]byte{7}) // piece id
		server.Write(payload)
	}()

	data, err := DownloadPiece(client, pieceIndex, pieceLength)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(data) != string(expected) {
		t.Fatalf("expected %q, got %q", expected, data)
	}
}