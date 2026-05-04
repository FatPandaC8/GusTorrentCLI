package tracker

import (
	"encoding/binary"
	"testing"
)

func TestBuildChoke(t *testing.T) {
	msg := BuildChoke()

	if len(msg) != 5 {
		t.Fatalf("expected length 5, got %d", len(msg))
	}

	if binary.BigEndian.Uint32(msg[0:4]) != 1 {
		t.Fatalf("expected length prefix 1")
	}

	if msg[4] != 0 {
		t.Fatalf("expected id 0 (choke)")
	}
}

func TestBuildInterested(t *testing.T) {
	msg := BuildInterested()

	if binary.BigEndian.Uint32(msg[0:4]) != 1 {
		t.Fatalf("expected length 1")
	}

	if msg[4] != 2 {
		t.Fatalf("expected id 2 (interested)")
	}
}

func TestBuildHave(t *testing.T) {
	msg := BuildHave(5)

	if binary.BigEndian.Uint32(msg[0:4]) != 5 {
		t.Fatalf("expected length 5")
	}

	if msg[4] != 4 {
		t.Fatalf("expected id 4 (have)")
	}

	index := binary.BigEndian.Uint32(msg[5:9])
	if index != 5 {
		t.Fatalf("expected piece index 5, got %d", index)
	}
}

func TestBitfieldSetAndHas(t *testing.T) {
	bf := make(Bitfield, 1) // 8 pieces

	bf.Set(0)
	bf.Set(3)

	if !bf.Has(0) {
		t.Fatalf("expected piece 0 to be set")
	}

	if !bf.Has(3) {
		t.Fatalf("expected piece 3 to be set")
	}

	if bf.Has(1) {
		t.Fatalf("expected piece 1 to be unset")
	}
}

func TestBuildBitfield(t *testing.T) {
	bf := Bitfield{0b10100000} // pieces 0 and 2

	msg := BuildBitfield(bf)

	length := binary.BigEndian.Uint32(msg[0:4])
	if length != uint32(len(bf)+1) {
		t.Fatalf("wrong length prefix")
	}

	if msg[4] != 5 {
		t.Fatalf("expected id 5 (bitfield)")
	}

	if msg[5] != bf[0] {
		t.Fatalf("bitfield payload mismatch")
	}
}

func TestBuildRequest(t *testing.T) {
	msg := BuildRequest(2, 16384, 16384)

	if binary.BigEndian.Uint32(msg[0:4]) != 13 {
		t.Fatalf("expected length 13")
	}

	if msg[4] != 6 {
		t.Fatalf("expected id 6 (request)")
	}

	index := binary.BigEndian.Uint32(msg[5:9])
	begin := binary.BigEndian.Uint32(msg[9:13])
	length := binary.BigEndian.Uint32(msg[13:17])

	if index != 2 || begin != 16384 || length != 16384 {
		t.Fatalf("request fields incorrect")
	}
}

func TestBuildPiece(t *testing.T) {
	block := []byte{1, 2, 3, 4}

	msg := BuildPiece(1, 0, block)

	length := binary.BigEndian.Uint32(msg[0:4])
	if length != uint32(9+len(block)) {
		t.Fatalf("wrong piece length")
	}

	if msg[4] != 7 {
		t.Fatalf("expected id 7 (piece)")
	}

	index := binary.BigEndian.Uint32(msg[5:9])
	begin := binary.BigEndian.Uint32(msg[9:13])

	if index != 1 || begin != 0 {
		t.Fatalf("wrong piece header")
	}

	for i := range block {
		if msg[13+i] != block[i] {
			t.Fatalf("block mismatch at %d", i)
		}
	}
}

func TestBuildCancel(t *testing.T) {
	msg := BuildCancel(1, 0, 1024)

	if binary.BigEndian.Uint32(msg[0:4]) != 13 {
		t.Fatalf("expected length 13")
	}

	if msg[4] != 8 {
		t.Fatalf("expected id 8 (cancel)")
	}
}

func TestBuildPort(t *testing.T) {
	msg := BuildPort(6881)

	if binary.BigEndian.Uint32(msg[0:4]) != 3 {
		t.Fatalf("expected length 3")
	}

	if msg[4] != 9 {
		t.Fatalf("expected id 9 (port)")
	}

	port := binary.BigEndian.Uint16(msg[5:7])
	if port != 6881 {
		t.Fatalf("wrong port value")
	}
}
