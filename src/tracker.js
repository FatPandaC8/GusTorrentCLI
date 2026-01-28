// Send a connect request
// Get the connect response and extract the connection id
// Use the connection id to send an announce request - this is where we tell the tracker which files we’re interested in
// Get the announce response and extract the peers list
// should also include the tcp as fallback incase udp is not what the torrent use
"use strict";
import dgram from 'dgram';
import crypto from 'crypto';
import { Buffer } from 'buffer';
import { URL } from 'url';

import * as util from '../utils.js'
import * as torrentParser from './torrent-parser.js';

export const getPeers = (torrent, callback) => {
    const socket = dgram.createSocket('udp4');
    // udp4 | udp6: 32-bit IP address | 128-bit IP address
    // const announceURL = torrent.announce.toString('utf-8'); == only a string
    const announceURL = new URL(new TextDecoder().decode(torrent.announce)); // an URL object
    
    if (announceURL.startsWith("udp")) {
        udpSend(socket, buildConnectReq(), announceURL);
    } else if (announceURL.startsWith("http")) { // a fallback in case torrent does not use udp
        // tcpSend(socket, buildConnReq(), announceURL);
        console.log("TODO: add tcp connection here")
    }

    socket.on('message', response => {
        if (respType(response) === 'connect') {
            const connResp = parseConnectResp(response);
            const announceResp = buildAnnounceReq(connResp.connectionId);
            udpSend(socket, announceResp, announceURL);
        } else if (respType(response) === 'announce') {
            const announceResp = parseAnnounceResp(response);
            callback(announceResp.peers);
        }
    });
};

function respType(resp) {
    const action = resp.readUInt32BE(0);
    if (action === 0) return "connect";
    if (action === 1) return "announce";
}

function udpSend(socket, message, rawURL, callback = (err) => {
    console.error({udp_send_error: err});
}) {
    socket.send(
        message,
        0,
        message.length,
        rawURL.port,
        rawURL.hostname,
        callback
    );
}

function buildConnectReq() {
    /*
    Offset  Size            Name            Value
    0       64-bit integer  protocol_id     0x41727101980 // magic constant
    8       32-bit integer  action          0 // connect
    12      32-bit integer  transaction_id
    */
    const buf = Buffer.alloc(16);

    // connection id
    buf.writeUInt32BE(0x417, 0);
    buf.writeUInt32BE(0x27101980, 4);

    // action
    buf.writeUInt32BE(0, 8);

    // transaction id
    crypto.randomBytes(4).copy(buf, 12);

    return buf;
}

function parseConnectResp(resp) {
    /*
    Offset  Size            Name            Value
    0       32-bit integer  action          0 // connect
    4       32-bit integer  transaction_id
    8       64-bit integer  connection_id
    */
    return {
        action: resp.readUInt32BE(0),
        transactionId: resp.readUInt32BE(4),
        connectionId: resp.slice(8),
    };
}

function buildAnnounceReq(connId, torrent, port = 6881) {
    /*
    Offset  Size    Name    Value
    0       64-bit integer  connection_id
    8       32-bit integer  action          1 // announce
    12      32-bit integer  transaction_id
    16      20-byte string  info_hash
    36      20-byte string  peer_id
    56      64-bit integer  downloaded
    64      64-bit integer  left
    72      64-bit integer  uploaded
    80      32-bit integer  event           0 // 0: none; 1: completed; 2: started; 3: stopped
    84      32-bit integer  IP address      0 // default
    88      32-bit integer  key
    92      32-bit integer  num_want        -1 // default
    96      16-bit integer  port
    */
    const buf = Buffer.allocUnsafe(98);

    // connection id
    connId.copy(buf, 0);

    // action
    buf.writeUInt32BE(1, 8);

    // transaction id
    crypto.randomBytes(4).copy(buf, 12);

    // info hash
    torrentParser.infoHash(torrent).copy(buf, 16);

    // peer id
    util.genPeerID().copy(buf, 36);

    // downloaded
    Buffer.alloc(8).copy(buf, 56);

    // left
    torrentParser.size(torrent).copy(buf, 64);

    // uploaded
    Buffer.alloc(8).copy(buf, 72);

    // event
    // 0: none; 1: completed; 2: started; 3: stopped
    buf.writeUInt32BE(0, 80);

    // ip address
    // 0 default
    buf.writeUInt32BE(0, 84);

    // key
    crypto.randomBytes(4).copy(buf, 88);

    // num want
    // -1 default
    buf.writeInt32BE(-1, 92);

    // port
    buf.writeUInt16BE(port, 96);

    return buf;
}

function parseAnnounceResp(resp) {
    /*
    Offset      Size            Name            Value
    0           32-bit integer  action          1 // announce
    4           32-bit integer  transaction_id
    8           32-bit integer  interval
    12          32-bit integer  leechers
    16          32-bit integer  seeders
    20 + 6 * n  32-bit integer  IP address
    24 + 6 * n  16-bit integer  TCP port
    */
    function group(iterable, groupSize) {
        let groups = [];
        for (let i = 0; i < iterable.length; i += groupSize) {
            groups.push(iterable.slice(i, i + groupSize));
        }
        return groups;
    }

    return {
        action: resp.readUInt32BE(0),
        transactionId: resp.readUInt32BE(4),
        interval: resp.readUInt32BE(8),
        leechers: resp.readUInt32BE(12),
        seeders: resp.readUInt32BE(16),
        peers: group(resp.slice(20), 6).map((address) => {
            return {
                ip: address.slice(0, 4).join("."),
                port: address.readUInt16BE(4),
            };
        }),
    };
}