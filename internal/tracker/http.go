package tracker

import (
	"encoding/binary"
	"fmt"
	"gustorrent/internal/bencode"
	"gustorrent/internal/metadata"
	"io"
	"net"
	"net/http"
	"net/url"
)

type TrackerResponse struct {
	Interval int
	Peers    []Peer
}

type Peer struct {
	IP   string
	Port uint16
}

func (p Peer) String() string {
	return fmt.Sprintf("%s:%d", p.IP, (p.Port))
}

func BuildTrackerURL(announce string, infoHash [20]byte, totalLength int) string {
	params := url.Values{}

	// TODO: make the peer id from config
	params.Set("info_hash", string(infoHash[:]))  // raw bytes
	params.Set("peer_id", "-GT0001-123456789012") // 20 bytes exactly, can be anyname
	params.Set("port", "6881")
	params.Set("uploaded", "0")
	params.Set("downloaded", "0")
	params.Set("left", fmt.Sprintf("%d", totalLength))

	// 6 bytes per peer:
	// [IP (4 bytes)][PORT (2 bytes)]
	params.Set("compact", "1")

	return announce + "?" + params.Encode()
}

func DecodePeers(body []byte) []Peer {
	var peers []Peer
	// 6 is because that's the length of 1 peer (4 ip: 2 port above)
	for i := 0; i+6 <= len(body); i += 6 {
		peers = append(peers, Peer{
			IP:   net.IP(body[i : i+4]).String(),
			Port: binary.BigEndian.Uint16(body[i+4 : i+6]),
		})
	}

	return peers
}

func GetPeers(data []byte) ([]Peer, error) {
	info, infoHash, err := metadata.GetMetadata(data)
	if err != nil {
		return []Peer{}, err
	}

	var totalLength int
	if info.Length != nil {
		totalLength = *info.Length
	} else {
		for _, f := range info.Files {
			totalLength += f.Length
		}
	}

	trackerURL := BuildTrackerURL(info.Announce, infoHash, totalLength)

	resp, err := http.Get(trackerURL)
	if err != nil {
		return []Peer{}, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []Peer{}, err
	}

	parsed_body, _, err := bencode.Decode(body, 0)
	if err != nil {
		return []Peer{}, err
	}

	return DecodePeers(parsed_body.Dict["peers"].Str), nil
}