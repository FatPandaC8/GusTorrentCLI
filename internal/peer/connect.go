package peer

import (
	"encoding/binary"
	"gustorrent/internal/tracker"
	"io"
	"net"
)

// For now: one peer one piece
func DownloadPiece(conn net.Conn, pieceIndex int, pieceLength int) ([]byte, error) {
	buf := make([]byte, pieceLength)

	const blockSize = 16 * 1024

	for offset := 0; offset < pieceLength; offset += blockSize {
		length := blockSize
		if offset+length > pieceLength {
			length = pieceLength - offset
		}

		// send ONE request
		req := tracker.BuildRequest(uint32(pieceIndex), uint32(offset), uint32(length))
		if _, err := conn.Write(req); err != nil {
			return nil, err
		}

		// wait for ONE block
		for {
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

			// VERY IMPORTANT check
			if msgIndex != pieceIndex {
				continue
			}

			if begin + len(block) > len(buf) {
				continue
			}
			
			copy(buf[begin:], block)
			break
		}
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