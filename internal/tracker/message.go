package tracker

import (
	"encoding/binary"
	"gustorrent/internal/metadata"
)

type Handshake struct {
	Pstrlen byte
	Pstr string
	Reserved [8]byte
	Infohash [20]byte
	Peerid [20]byte
}

func BuildHandshake(data []byte, peerId string) ([]byte, error) {
	_, infoHash, err := metadata.GetMetadata(data)
	if err != nil {
		return []byte{}, err
	}

	pstr := "BitTorrent protocol" // according to version 1.0 of Bittorrent

	buf := make([]byte, 49 + len(pstr)) // 1 + 19 + 8 + 20 + 20 = 68

	// pstrlen
	buf[0] = byte(len(pstr))

	// pstr
	copy(buf[1:], pstr)

	// reserved 
	// 8 bytes already zeroed by make by default

	// infoHash 
	copy(buf[1 + len(pstr) + 8:], infoHash[:])

	// peer id
	copy(buf[1 + len(pstr) + 8 + 20:], []byte(peerId))

	return buf, nil
}

func BuildKeepAlive() []byte {
	// keep-alive: <len=0000>

	// The keep-alive message is a message with zero bytes, specified with the length prefix set to zero. 
	// There is no message ID and no payload. 
	// Peers may close a connection if they receive no messages (keep-alive or any other message) for a certain period of time, 
	// so a keep-alive message must be sent to maintain the connection alive if no command have been sent for a given amount of time. 
	// This amount of time is generally two minutes. 

	return []byte{
		0, 0, 0, 0,
	}
}

func BuildChoke() []byte {
	// choke: <len=0001><id=0>

	// The choke message is fixed-length and has no payload.

	return []byte{
		0, 0, 0, 1, // length = 1
		0,          // message id = choke
	}
}

func BuildUnchoke() []byte {
	// unchoke: <len=0001><id=1>

	// The choke message is fixed-length and has no payload.

	return []byte{
		0, 0, 0, 1, // length = 1
		1,          // message id = unchoke
	}
}

func BuildInterested() []byte {
	/*
        interested: <len=0001><id=2>

        The interested message is fixed-length and has no payload. 
    */
	return []byte{
		0, 0, 0, 1,
		2,
	}
}

func BuildUninterested() []byte {
	/*
        uninterested: <len=0001><id=3>

        The interested message is fixed-length and has no payload. 
    */
	return []byte{
		0, 0, 0, 1,
		3,
	}
}

func BuildHave(pieceIndex int) []byte {
	/*
        have: <len=0005><id=4><piece index>

        The have message is fixed length. 
        The payload is the zero-based index of a piece that has just been successfully downloaded and verified via the hash.
		Piece index is 32 bit in == 4 bytes int

        Implementer's Note: 
            That is the strict definition, in reality some games may be played. 
            In particular because peers are extremely unlikely to download pieces that they already have, 
            a peer may choose not to advertise having a piece to a peer that already has that piece
    */

	buf := make([]byte, 9)

	binary.BigEndian.PutUint32(buf[0:4], 5)
	buf[4] = 4

	binary.BigEndian.PutUint32(buf[5:9], uint32(pieceIndex))

	return buf
}

type Bitfield []byte

func (b Bitfield) Set(piece int) {
	byteIndex := piece / 8 // 1 byte = 8 bits
	bitIndex := piece % 8 // which bit in those 1 byte

	b[byteIndex] |= 1 << (7 - bitIndex)
}

func (b Bitfield) Has(piece int) bool {
	byteIndex := piece / 8
	bitIndex := piece % 8

	return b[byteIndex] & (1<<(7-bitIndex)) != 0
}

func BuildBitfield(bitfield Bitfield) []byte {
	/*
        bitfield: <len=0001+X><id=5><bitfield>

        The bitfield message may only be sent immediately after the handshaking sequence is completed, 
        and before any other messages are sent. It is optional, and need not be sent if a client has no pieces.

        The bitfield message is variable length, where X is the length of the bitfield / numebr of pieces.
        The payload is a bitfield representing the pieces that have been successfully downloaded. 
        The high bit in the first byte corresponds to piece index 0. 
        Bits that are cleared indicated a missing piece, and set bits indicate a valid and available piece. 
        Spare bits at the end are set to zero. 

		pieces:   0 1 2 3 4 5 6 7
		bits:     1 1 0 1 1 0 0 1
    */

	length := len(bitfield) + 1 // 1 byte for message ID

	buf := make([]byte, 4+length)

	binary.BigEndian.PutUint32(buf[0:4], uint32(length))

	buf[4] = 5

	copy(buf[5:], bitfield)

	return buf
}

func BuildRequest(index, begin, length uint32) []byte {
	/*
        request: <len=0013><id=6><index><begin><length>
		Everything except id is 4 byte, id is 1 byte

        The request message is fixed length, and is used to request a block. The payload contains the following information:

        index: integer specifying the zero-based piece index
        begin: integer specifying the zero-based byte offset within the piece
        length: integer specifying the requested length.
    */

	buf := make([]byte, 17)
	binary.BigEndian.PutUint32(buf[0:4], 13)
	buf[4] = 6

	binary.BigEndian.PutUint32(buf[5:9], index)
	binary.BigEndian.PutUint32(buf[9:13], begin)
	binary.BigEndian.PutUint32(buf[13:17], length)

	return buf
}

func BuildPiece(index, begin uint32, block []byte) []byte {
	length := 9 + len(block)

	buf := make([]byte, 4 + length)

	// length prefix
	binary.BigEndian.PutUint32(buf[0:4], uint32(length))

	// message id = 7
	buf[4] = 7

	// index
	binary.BigEndian.PutUint32(buf[5:9], index)

	// begin
	binary.BigEndian.PutUint32(buf[9:13], begin)

	// block data
	copy(buf[13:], block)

	return buf
}

func BuildCancel(index, begin, length uint32) []byte {
	buf := make([]byte, 17)

	binary.BigEndian.PutUint32(buf[0:4], 13)

	buf[4] = 8

	binary.BigEndian.PutUint32(buf[5:9], index)
	binary.BigEndian.PutUint32(buf[9:13], begin)
	binary.BigEndian.PutUint32(buf[13:17], length)

	return buf
}

func BuildPort(port uint16) []byte {
	buf := make([]byte, 7)

	// length = 3
	binary.BigEndian.PutUint32(buf[0:4], 3)

	// id = 9
	buf[4] = 9

	// port (2 bytes)
	binary.BigEndian.PutUint16(buf[5:7], port)

	return buf
}