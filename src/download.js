"use strict";

import net from 'net';
import * as message from './message.js';
import * as tracker from './tracker.js'

export default function (torrent, path) {
    tracker.getPeers(torrent, (peers) => {
        console.log(torrent.info.files);

        peers.forEach(download);
    });
}

export function download(peer, torrent) {
    const socket = net.Socket();
    socket.on("error", err => {
      console.log("ERROR:", err.message);
    });

    socket.connect(peer.port, peer.ip, () => {
        socket.write(message.buildHandshake(torrent))
    })
}

function onWholeMessage(socket, callback) {
    let savedBuffer = Buffer.alloc(0);
    let handshake = true;

    socket.on('data', recvBuf => {
        const msgLen = () => handshake ? savedBuffer.readUInt8(0) + 49 : savedBuffer.readInt32BE(0) + 4;
    })
}