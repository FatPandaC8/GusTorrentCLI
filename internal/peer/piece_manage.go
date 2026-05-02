package peer

import (
	"crypto/sha1"
	"gustorrent/internal/metadata"
	"gustorrent/internal/tracker"
	"io"
	"net"
	"sync"
	"time"
)

type PieceState int 

const (
	Missing PieceState = iota 
	Downloading 
	Done
)

type PieceManager struct {
	mu sync.Mutex

	total int
	done int

	states []PieceState
}

func NewPieceManager(totalPieces int) *PieceManager {
	return &PieceManager{
		states: make([]PieceState, totalPieces),
		total: totalPieces,
	}
}

func (p *PieceManager) PickPiece(bitfield tracker.Bitfield) (int, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, state := range p.states {
		if state == Missing && bitfield.Has(i) {
			p.states[i] = Downloading
			return i, true
		}
	}

	return -1, false
}

func (p *PieceManager) MarkDone(index int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.states[index] != Done {
		p.states[index] = Done
		p.done++
	}
}

func (p *PieceManager) MarkFailed(index int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.states[index] = Missing
}

func PeerWorker(addr string, pm *PieceManager, data []byte, info *metadata.Info, lastPieceLength int) {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return
	}
	defer conn.Close()

	// handshake
	hs, err := tracker.BuildHandshake(data, "-GT0001-123456789012")
	if err != nil {
		panic(err)
	}
	if _, err := conn.Write(hs); err != nil {
		return
	}

	resp := make([]byte, 68)
	if _, err := io.ReadFull(conn, resp); err != nil {
		return
	}
	// read bitfield
	var bitfield []byte
	for {
		id, payload, err := ReadMessage(conn)
		if err != nil {
			return
		}
		if id == 5 {
			bitfield = payload
			break
		}
	}

	// interested
	conn.Write(tracker.BuildInterested())

	// wait unchoke
	for {
		id, _, err := ReadMessage(conn)
		if err != nil {
			return
		}
		if id == 1 {
			break
		}
	}

	if pm.Done() >= pm.Total() {
		return
	}

	for {
		pieceIndex, ok := pm.PickPiece(bitfield)
		if !ok {
			return
		}

		pieceLength := info.PieceLength
		if pieceIndex == pm.total-1 {
			pieceLength = lastPieceLength
		}

		piece, err := DownloadPiece(conn, pieceIndex, pieceLength)
		if err != nil {
			pm.MarkFailed(pieceIndex)
			continue
		}

		if !VerifyPiece(info, pieceIndex, piece) {
			pm.MarkFailed(pieceIndex)
			continue
		}

		pm.MarkDone(pieceIndex)
	}
}

func VerifyPiece(info *metadata.Info, index int, data []byte) bool {
	hashStart := index * 20
	expected := info.Pieces[hashStart : hashStart+20]

	actual := sha1.Sum(data)

	return string(actual[:]) == string(expected)
}

func (pm *PieceManager) Done() int {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.done
}

func (pm *PieceManager) Total() int {
	return pm.total
}