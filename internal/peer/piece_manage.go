package peer

import (
	"gustorrent/internal/tracker"
	"sync"
)

type PieceState int 

const (
	Missing PieceState = iota 
	Downloading 
	Done
)

type PieceManager struct {
	mu sync.Mutex
	states []PieceState
}

func NewPieceManager(totalPieces int) *PieceManager {
	return &PieceManager{
		states: make([]PieceState, totalPieces),
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

	p.states[index] = Done
}

func (p *PieceManager) MarkFailed(index int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.states[index] = Missing
}