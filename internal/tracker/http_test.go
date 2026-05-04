package tracker

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeClient struct{}
var standardClient = http.DefaultClient

func (f fakeClient) Get(url string) (*http.Response, error) {
	return &http.Response{
		Body: io.NopCloser(badBody{}),
	}, nil
}

func TestDecodePeers_Valid(t *testing.T) {
	body := []byte{
		127, 0, 0, 1, 0x1A, 0xE1, // 127.0.0.1:6881
		192, 168, 1, 1, 0x00, 0x50, // 192.168.1.1:80
	}

	peers := DecodePeers(body)

	if len(peers) != 2 {
		t.Fatalf("expected 2 peers, got %d", len(peers))
	}

	if peers[0].IP != "127.0.0.1" || peers[0].Port != 6881 {
		t.Fatalf("unexpected peer 1: %+v", peers[0])
	}
}

func TestDecodePeers_SinglePeer(t *testing.T) {
	body := []byte{127, 0, 0, 1, 0x1A, 0xE1}

	peers := DecodePeers(body)

	if len(peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(peers))
	}
}

func TestBuildTrackerURL(t *testing.T) {
	var hash [20]byte
	copy(hash[:], "12345678901234567890")

	url := BuildTrackerURL("http://tracker", hash, 100)

	if !strings.Contains(url, "info_hash=") {
		t.Fatal("missing info_hash")
	}
	if !strings.Contains(url, "left=100") {
		t.Fatal("missing left param")
	}
}

func TestGetPeers_Success(t *testing.T) {
	// fake tracker server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// compact peer list (1 peer)
		peers := []byte{127, 0, 0, 1, 0x1A, 0xE1}

		// bencoded response: d8:intervali1800e5:peers6:<peer_bytes>e
		resp := append([]byte("d8:intervali1800e5:peers6:"), peers...)
		resp = append(resp, 'e')

		w.Write(resp)
	}))
	defer server.Close()

	// minimal valid torrent
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

	peers, err := GetPeers(standardClient, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(peers))
	}
}

func TestGetPeers_MetadataError(t *testing.T) {
	_, err := GetPeers(standardClient, []byte("invalid"))

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetPeers_HTTPError(t *testing.T) {
	data := []byte(
		"d8:announce13:http://invalid" +
			"4:infod" +
			"12:piece lengthi4e" +
			"6:pieces20:aaaaaaaaaaaaaaaaaaaa" +
			"4:name8:file.txt" +
			"6:lengthi100e" +
			"ee",
	)

	_, err := GetPeers(standardClient, data)
	if err == nil {
		t.Fatal("expected http error")
	}
}

func TestGetPeers_InvalidBencodeResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not-bencode"))
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

	_, err := GetPeers(standardClient, data)
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestGetPeers_MultiFileTotalLength(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty peers
		w.Write([]byte("d5:peers0:e"))
	}))
	defer server.Close()

	data := []byte(
		"d" +
			"8:announce" + fmt.Sprintf("%d:%s", len(server.URL), server.URL) +
			"4:infod" +
			"12:piece lengthi4e" +
			"6:pieces20:aaaaaaaaaaaaaaaaaaaa" +
			"4:name8:file.txt" +
			"5:filesl" +
			"d6:lengthi50e4:pathl9:file1.txtee" +
			"d6:lengthi70e4:pathl9:file2.txtee" +
			"e" +
			"ee",
	)

	_, err := GetPeers(standardClient, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type badBody struct{}

func (b badBody) Read(p []byte) (int, error) {
	return 0, errors.New("read failed")
}

func (b badBody) Close() error { return nil }

func TestGetPeers_LengthNilPanic(t *testing.T) {
	data := []byte(
		"d8:announce11:http://test" +
			"4:infod" +
			"12:piece lengthi4e" +
			"6:pieces20:aaaaaaaaaaaaaaaaaaaa" +
			"4:name8:file.txt" +
			"5:filesle" + // multi-file, no length
			"ee",
	)

	GetPeers(standardClient, data)
}

func TestToString(t *testing.T) {
	var p Peer = Peer{
		IP: "123.123.123.123",
		Port: 9090,
	}
	if p.String() != "123.123.123.123:9090" {
		t.Fatalf("expected 123.123.123.123 9090, got:%s", p.String())
	}
}

func TestGetPeers_ReadError(t *testing.T) {
	client := fakeClient{}
	data := []byte(
		"d8:announce55:http://bittorrent-test-tracker.codecrafters.io/announce" +
			"4:infod" +
			"12:piece lengthi4e" +
			"6:pieces20:aaaaaaaaaaaaaaaaaaaa" +
			"4:name8:file.txt" +
			"6:lengthi100e" +
			"ee",
	)

	_, err := GetPeers(client, data)
	if err == nil {
		t.Fatal("expected error")
	}
}