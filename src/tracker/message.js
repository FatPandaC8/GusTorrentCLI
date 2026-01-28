"use strict";
import * as util from '../utils.js';
import * as fromTorrent from '../torrent-parser.js';

export function buildHandshake(torrent) {
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
    return Buffer.alloc(4);
}

export function buildChoke() {
    const buf = Buffer.alloc(5);
    buf.writeUint32BE(1, 0);
    buf.writeUint8(0, 4);
    return buf;
}

export function buildInterest() {
    const buf = Buffer.alloc(5);
    buf.writeUInt32BE(1, 0);
    buf.writeUInt8(2, 4);
    return buf;
}

export function buildUninterested() {
    const buf = Buffer.alloc(5);
    buf.writeUInt32BE(1, 0); 
    buf.writeUInt8(3, 4);
    return buf;
}

export function buildHave(payload) {
    const buf = Buffer.alloc(9);
    buf.writeUInt32BE(5, 0);
    buf.writeUInt8(4, 4);
    buf.writeUInt32BE(payload, 5);
    return buf;
}

export function buildBitfield(payload) {
    const buf = Buffer.alloc(payload.length + 1 + 4);
    buf.writeUInt32BE(1 + payload.length, 0);
    buf.writeUInt8(5, 4);
    payload.copy(buf, 5);
    return buf;
}

export function buildRequest(payload) {
    const buf = Buffer.alloc(17);
    buf.writeUInt32BE(13, 0);
    buf.writeUInt8(6, 4);
    buf.writeUInt32BE(payload.index, 5);
    buf.writeUInt32BE(payload.begin, 9);
    buf.writeUInt32BE(payload.length, 13);
    return buf;
}

export function buildPiece(payload) {
    const buf = Buffer.alloc(13 + payload.block.length);
    buf.writeUInt32BE(9 + payload.block.length, 0);
    buf.writeUInt8(7, 4);
    buf.writeUInt32BE(payload.index, 5);
    buf.writeUInt32BE(payload.begin, 9);
    payload.block.copy(buf, 13);
    return buf;
}

export function buildCancel(payload) {
    const buf = Buffer.alloc(17);
    buf.writeUInt32BE(13, 0);
    buf.writeUInt8(8, 4);
    buf.writeUInt32BE(payload.index, 5);
    buf.writeUInt32BE(payload.begin, 9);
    buf.writeUInt32BE(payload.length, 13);
    return buf;
}

export function buildPort(payload) {
    const buf = Buffer.alloc(7);
    buf.writeUInt32BE(3, 0);
    buf.writeUint8(9, 4);
    buf.writeUInt16BE(payload, 5);
    return buf;
}

export function parse(msg) {
    const size = msg.readUInt32BE(0);
    const id = msg.length > 4 ? msg.readInt8(4) : null;
    let payload = msg.length > 5 ? msg.slice(5) : null;

    if (id === 6 || id === 7 || id === 8) {
        const rest = payload.slice(8);
        payload = {
            index: payload.readInt32BE(0),
            begin: payload.readInt32BE(4),
        };
        payload[id === 7 ? 'block' : 'length'] = rest;
    }

    return {size, id, payload};
}