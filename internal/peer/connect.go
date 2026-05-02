package peer

import (
	"encoding/binary"
	"gustorrent/internal/tracker"
	"io"
	"net"
)

const MaxRequest = 5
const blockSize = 16 * 1024

func DownloadPiece(conn net.Conn, pieceIndex int, pieceLength int) ([]byte, error) {
	buf := make([]byte, pieceLength)

	type blockState struct {
		requested bool
		received  bool
	}

	// track each block by offset
	blocks := make(map[int]*blockState)

	// initialize all blocks
	for offset := 0; offset < pieceLength; offset += blockSize {
		blocks[offset] = &blockState{}
	}

	inflight := 0
	completed := 0
	totalBlocks := len(blocks)

	nextOffset := 0

	for completed < totalBlocks {

		// fill pipeline
		for inflight < MaxRequest && nextOffset < pieceLength {
			state := blocks[nextOffset]
			if state.requested {
				nextOffset += blockSize
				continue
			}

			length := blockSize
			if nextOffset+length > pieceLength {
				length = pieceLength - nextOffset
			}

			req := tracker.BuildRequest(
				uint32(pieceIndex),
				uint32(nextOffset),
				uint32(length),
			)

			if _, err := conn.Write(req); err != nil {
				return nil, err
			}

			state.requested = true
			inflight++

			nextOffset += blockSize
		}

		// receive response
		id, payload, err := ReadMessage(conn)
		if err != nil {
			return nil, err
		}

		if id != 7 {
			continue
		}

		msgIndex := int(binary.BigEndian.Uint32(payload[0:4]))
		begin := int(binary.BigEndian.Uint32(payload[4:8]))
		block := payload[8:]

		// validate
		if msgIndex != pieceIndex {
			continue
		}

		state, ok := blocks[begin]
		if !ok {
			continue
		}

		if state.received {
			continue // duplicate block
		}

		if begin+len(block) > len(buf) {
			continue
		}

		// copy data
		copy(buf[begin:], block)

		state.received = true
		inflight--
		completed++
	}

	return buf, nil
}

func ReadMessage(conn net.Conn) (uint8, []byte, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lenBuf); err != nil { // use this if tcp decides to split messages 
		return 0, nil, err
	}

	length := binary.BigEndian.Uint32(lenBuf)

	// 2. keep-alive
	if length == 0 {
		return 255, nil, nil // special ID for keep-alive
	}

	// 3. read full message
	msg := make([]byte, length)
	if _, err := io.ReadFull(conn, msg); err != nil {
		return 0, nil, err
	}

	// 4. first byte = message ID
	id := msg[0]
	payload := msg[1:]

	return id, payload, nil
}