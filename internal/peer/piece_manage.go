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
	done  int

	states []PieceState
	freq   map[int]int

	endgame  bool
	inFlight map[int]bool
}

func NewPieceManager(totalPieces int) *PieceManager {
	return &PieceManager{
		states:   make([]PieceState, totalPieces),
		total:    totalPieces,
		freq:     make(map[int]int),
		inFlight: make(map[int]bool),
	}
}

// You should NOT double count if same peer reconnects, but we’ll ignore that for now (MVP is fine).
func (p *PieceManager) AddPeerBitfield(b tracker.Bitfield) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 0; i < p.total; i++ {
		if b.Has(i) {
			p.freq[i]++
		}
	}
}

func (p *PieceManager) checkEndgame() {
	remaining := 0

	for i := 0; i < p.total; i++ {
		if p.states[i] == Missing {
			remaining++
		}
	}

	if remaining <= 5 { // threshold (common heuristic)
		p.endgame = true
	}
}

// Pick a piece from the bitfield then return the piece_index and error
// score = rarity + peer speed + availability + choking state <- real bittorrent client do this
func (p *PieceManager) PickPiece(bitfield tracker.Bitfield) (int, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.checkEndgame()

	if !p.endgame {

		best := -1
		// ^ is bitwise not in golang -> flip 0 to all ones
		bestFreq := int(^uint(0) >> 1) // max int

		for i := 0; i < p.total; i++ {
			if p.states[i] != Missing {
				continue
			}
			if !bitfield.Has(i) {
				continue
			}

			freq := p.freq[i]
			if freq == 0 {
				freq = 1
			}

			if freq < bestFreq {
				bestFreq = freq
				best = i
			}
		}

		if best == -1 {
			return -1, false
		}
		p.states[best] = Downloading
		return best, true
	}

	for i := 0; i < p.total; i++ {
		if p.states[i] == Missing && bitfield.Has(i) {
			if p.inFlight[i] {
				continue
			}

			p.inFlight[i] = true
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

	delete(p.inFlight, index)
	p.checkEndgame()
}

func (p *PieceManager) MarkFailed(index int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.states[index] = Missing
	delete(p.inFlight, index)
}

func (p *PieceManager) MarkInFlightDone(index int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.inFlight, index)
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
			pm.AddPeerBitfield(tracker.Bitfield(bitfield))
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
