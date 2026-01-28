"use strict";

import net from 'net';
import fs from 'fs';
import * as message from './message.js';
import * as tracker from './tracker.js';
import * as Queue from './Queue.js';
import * as Piece from './Pieces.js';

export default function (torrent, path) {
    tracker.getPeers(torrent, (peers) => {
        const pieces = new Piece(torrent);
        const file = fs.openSync(path, 'w');
        peers.forEach(peer => download(peer, torrent, pieces, file));
    });
}

export function download(peer, torrent, pieces, file) {
    const socket = net.Socket();
    socket.on("error", err => {
      console.log("ERROR:", err.message);
    });

    socket.connect(peer.port, peer.ip, () => {
        socket.write(message.buildHandshake(torrent))
    })

    const queue = new Queue(torrent);
    onWholeMessage(socket, (msg) => {
        messageHandler(msg, socket, pieces, queue, torrent, file);
    })
}

function onWholeMessage(socket, callback) {
    let savedBuffer = Buffer.alloc(0);
    let handshake = true;

    socket.on('data', recvBuf => {
        // The handshake is a required message and must be the first message transmitted by the client. It is (49+len(pstr)) bytes long.
        // handshake: <pstrlen><pstr><reserved><info_hash><peer_id>
        const msgLen = () => handshake ? savedBuffer.readUInt8(0) + 49 : savedBuffer.readInt32BE(0) + 4;
        savedBuffer = Buffer.concat([savedBuffer, recvBuf]);

        while (savedBuffer.length >= 4 && savedBuffer.length >= msgLen()) {
            callback(savedBuffer.subarray(0, msgLength()));
            savedBuffer = savedBuffer.subarray(msgLength()); // clear saved buffer
            handshake = false;
        }
    });
}

function messageHandler(msg, socket, pieces, queue, torrent, file) {
    if (isHandshake(msg)) {
        socket.write(message.buildInterest());
    } else {
        const parsedMsg = message.parse(msg);
        switch (parsedMsg.id) {
            case 0: {
                chokeHandler(socket);
                break;
            }
            case 1: {
                unchokeHandler(socket, pieces, queue);
                break;
            }
            case 4: {
                haveHandler(socket, pieces, queue, parsedMsg.payload);
                break;
            }
            case 5: {
                bitfieldHandler(socket, pieces, queue, parsedMsg.payload);
                break;
            }
            case 7: {
                pieceHandler(socket, pieces, queue, torrent, file, parsedMsg.payload);
                break;
            }
        }
    }
}

function isHandshake(msg) {
    return (
        msg.length === msg.readUInt8(0) + 49 &&
        msg.toString("utf8", 1) === "BitTorrent protocol"
    );
}

function chokeHandler(socket) {
    socket.end();
}

function unchokeHandler(socket, pieces, queue) {
    queue.choked = false;
    requestPiece(socket, pieces, queue);
}

function haveHandler(socket, pieces, queue, payload) {
    const pieceIndex = payload.readUInt32BE(0);
    const queueEmpty = queue.length === 0;

    queue.queue(pieceIndex);

    if (queueEmpty) requestPiece(socket, pieces, queue);
}

function bitfieldHandler(socket, pieces, queue, payload) {
    const queueEmpty = queue.length === 0;
    payload.forEach((byte, i) => {
        for (let j = 0; j < 8; j++) {
            if (byte % 2) queue.queue(i * 8 + 7 - j);
            byte = Math.floor(byte / 2);
        }
    });
    if (queueEmpty) requestPiece(socket, pieces, queue);
}

function pieceHandler(socket, pieces, queue, torrent, file, pieceResp) {
  pieces.printPercentDone();

  pieces.addReceived(pieceResp);

  const offset = pieceResp.index * torrent.info['piece length'] + pieceResp.begin;
  fs.write(file, pieceResp.block, 0, pieceResp.block.length, offset, () => {});

  if (pieces.isDone()) {
    console.log('DONE!');
    socket.end();
    try { fs.closeSync(file); } catch(e) {}
  } else {
    requestPiece(socket,pieces, queue);
  }
}

function requestPiece(socket, pieces, queue) {
    if (queue.choked) return null;

    while (queue.length()) {
        const pieceBlock = queue.deque();
        if (pieces.needed(pieceBlock)) {
            socket.write(message.buildRequest(pieceBlock));
            pieces.addRequested(pieceBlock);
            break;
        }
    }
}