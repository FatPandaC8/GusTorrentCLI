package peer

import (
	"crypto/sha1"
	"gustorrent/internal/metadata"
	"gustorrent/internal/tracker"
	"sync"
	"testing"
)

func TestPickPiece(t *testing.T) {
	pm := NewPieceManager(5)

	bf := tracker.Bitfield{0b10100000}

	idx, ok := pm.PickPiece(bf)
	if !ok {
		t.Fatal("expected a piece")
	}

	if idx != 0 {
		t.Fatalf("expected 0, got %d", idx)
	}

	// ensure state changed
	if pm.states[idx] != Downloading {
		t.Fatal("expected state to be Downloading")
	}
}

func TestPickPiece_NoneAvailable(t *testing.T) {
	pm := NewPieceManager(3)

	bf := tracker.Bitfield([]byte{0b00000000})

	_, ok := pm.PickPiece(bf)
	if ok {
		t.Fatal("expected no piece")
	}
}

func TestMarkDone_Idempotent(t *testing.T) {
	pm := NewPieceManager(2)

	pm.MarkDone(1)
	pm.MarkDone(1)

	if pm.done != 1 {
		t.Fatalf("expected done=1, got %d", pm.done)
	}
}

func TestMarkFailed(t *testing.T) {
	pm := NewPieceManager(2)

	pm.states[1] = Downloading

	pm.MarkFailed(1)

	if pm.states[1] != Missing {
		t.Fatal("expected state to reset to Missing")
	}
}

func TestPieceManager_Concurrency(t *testing.T) {
	pm := NewPieceManager(100)

	data := make([]byte, 13)
	for i := range data {
		data[i] = 0xFF // all bits = 1
	}

	bf := tracker.Bitfield(data)

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				idx, ok := pm.PickPiece(bf)
				if !ok {
					return
				}
				pm.MarkDone(idx)
			}
		}()
	}

	wg.Wait()

	if pm.Done() != 100 {
		t.Fatalf("expected all done, got %d", pm.Done())
	}
}

func TestVerifyPiece(t *testing.T) {
	data := []byte("hello")

	hash := sha1.Sum(data)

	info := &metadata.Info{
		Pieces: hash[:],
	}

	if !VerifyPiece(info, 0, data) {
		t.Fatal("expected valid piece")
	}
}

func TestVerifyPiece_Invalid(t *testing.T) {
	data := []byte("hello")

	info := &metadata.Info{
		Pieces: make([]byte, 20), // wrong hash
	}

	if VerifyPiece(info, 0, data) {
		t.Fatal("expected invalid piece")
	}
}