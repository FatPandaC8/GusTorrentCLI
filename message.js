"use strict";
import * as util from './utils.js';
import * as fromTorrent from './torrent.js';

export function buildHandshake(torrent) {
    /*
    handshake: <pstrlen><pstr><reserved><info_hash><peer_id>

    pstrlen: string length of <pstr>, as a single raw byte
    pstr: string identifier of the protocol
    reserved: eight (8) reserved bytes. All current implementations use all zeroes.
    peer_id: 20-byte string used as a unique ID for the client.

    In version 1.0 of the BitTorrent protocol, pstrlen = 19, and pstr = "BitTorrent protocol".
    */
    const buf = Buffer.alloc(68);

    // pstrlen
    buf.writeUint8(19, 0);

    // pstr
    buf.write("BitTorrent protocol", 1);

    // reserved
    buf.fill(0, 20, 28);

    // info hash
    fromTorrent.infoHash(torrent).copy(buf, 28);

    // peer id
    util.genPeerID().copy(buf, 48);

    return buf;
}

export function buildKeepAlive() {
    /*
    keep-alive: <len=0000>

    The keep-alive message is a message with zero bytes, specified with the length prefix set to zero.
    There is no message ID and no payload. 
    Peers may close a connection if they receive no messages (keep-alive or any other message) for a certain period of time, 
    so a keep-alive message must be sent to maintain the connection alive if no command have been sent for a given amount of time. 
    This amount of time is generally two minutes.  
    */
    return Buffer.alloc(4);
}

export function buildChoke() {
    /*
    choke: <len=0001><id=0>
    The choke message is fixed-length and has no payload. 
     */
    const buf = Buffer.alloc(5);

    buf.writeUint32BE(1, 0);
    buf.writeUint8(1, 4);
    return buf;
}

export function buildInterest() {
    /**
     * interested: <len=0001><id=2>
     * The interested message is fixed-length and has no payload. 
     */
    const buf = Buffer.alloc(5);

    buf.writeUInt32BE(1, 0); // length = 1
    buf.writeUInt8(2, 4);   // id = 2 (interested)

    return buf;
}

export function buildUninterested() {
    /*
    * not interested: <len=0001><id=3>
    * The not interested message is fixed-length and has no payload.
    */
    const buf = Buffer.alloc(5);
    buf.writeUInt32BE(1, 0); 
    buf.writeUInt8(3, 4);
    return buf;
}

export function buildHave(payload) {
    /*
    * have: <len=0005><id=4><piece index>
    * len = 4 bytes, id = 1 byte, pieceIndex = 4 bytes
    * The have message is fixed length. The payload is the zero-based
    * index of a piece that has just been successfully downloaded and
    * verified via the hash.
    */

    const buf = Buffer.alloc(9);
    buf.writeUInt32BE(5, 0);
    buf.writeUInt8(4, 4);
    buf.writeUInt32BE(payload, 5);
    return buf;
}

export function buildBitfield(payload) {
    /*
    * bitfield: <len=0001+X><id=5><bitfield>
    * The bitfield message is variable length, where X is the length of
    * the bitfield. The payload is a bitfield representing the pieces that
    * have been successfully downloaded. The high bit in the first byte
    * corresponds to piece index 0. Bits that are cleared indicate a
    * missing piece, and set bits indicate a valid and available piece.
    * Spare bits at the end are set to zero.
    */

    const buf = Buffer.alloc(payload.length + 1 + 4);
    buf.writeUInt32BE(1 + payload.length, 0);
    buf.writeUInt8(5, 4);
    payload.copy(buf, 5);

    return buf;
}

export function buildRequest(payload) {
    /*
    * request: <len=0013><id=6><index><begin><length>
    * The request message is fixed length, and is used to request a block
    
    * The payload contains the following information:
    * index: integer specifying the zero-based piece index
    * begin: integer specifying the zero-based byte offset within the piece
    * length: integer specifying the requested length.
    */
    const buf = Buffer.alloc(17);
    buf.writeUInt32BE(13);
    buf.writeUInt8(6, 4);
    buf.writeUInt32BE(payload.index, 5);
    buf.writeUInt32BE(payload.begin, 9);
    buf.writeUInt32BE(payload.length, 13);

    return buf;
}

export function buildPiece(payload) {
    /*
    * piece: <len=0009+X><id=7><index><begin><block>
    * The piece message is variable length, where X is the length of the block. 

    * The payload contains the following information:
    * index: integer specifying the zero-based piece index
    * begin: integer specifying the zero-based byte offset within the piece
    * block: block of data, which is a subset of the piece specified by index.
    */
    const buf = Buffer.alloc(13 + payload.block.length);
    buf.writeUInt32BE(9 + payload.block.length, 0);
    buf.writeUInt8(7, 4);
    buf.writeUInt32BE(payload.index, 5);
    buf.writeUInt32BE(payload.begin, 9);
    payload.block.copy(buf, 13);

    return buf;
}

export function buildCancel() {
    /*
    * cancel: <len=0013><id=8><index><begin><length>
    * The cancel message is fixed length, and is used to cancel block requests.
    * The payload is identical to that of the "request" message.
    */
    const buf = Buffer.alloc(17);
    buf.writeUInt32BE(13, 0);
    buf.writeUInt32BE(8, 4);
    buf.writeUInt32BE(payload.index, 5);
    buf.writeUInt32BE(payload.begin, 9);
    buf.writeUInt32BE(payload.length, 13);

    return buf;
}

export function buildPort(payload) {
    /*
    * port: <len=0003><id=9><listen-port>
    * The port message is sent by newer versions of the Mainline that implements
    * a DHT tracker. The listen port is the port this peer's DHT node is listening on.
    */
    const buf = Buffer.alloc(7);
    buf.writeUInt32BE(3, 0);
    buf.writeUInt32BE(9, 4);
    buf.writeUInt16BE(payload, 5);

    return buf;
}
