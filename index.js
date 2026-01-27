import fs from "fs";
import bencode from "bencode";
import crypto from "crypto";
import net from 'net';

const BLOCK_SIZE = 16384;
const torrent = bencode.decode(fs.readFileSync("sample.torrent"));

let currentPiece = 0;
let pieceSize = Math.min(
  torrent.info['piece length'],
  torrent.info.length
);

let received = 0;
let pieceBuffer = Buffer.alloc(pieceSize);

// read torrent file

// ---- INFO HASH ----
const infoBencoded = bencode.encode(torrent.info);
const infoHash = crypto.createHash("sha1").update(infoBencoded).digest();
const decorder = new TextDecoder('utf-8');

console.log("TRACKER:", decorder.decode(torrent.announce));
console.log("NAME:", decorder.decode(torrent.info.name));
console.log("FILE LENGTH:", torrent.info.length);
console.log("PIECE LENGTH:", torrent.info['piece length']);
console.log("INFO HASH:", infoHash.toString('hex'));

// ---- PIECE HASHES ----
const pieces = torrent.info.pieces;
const numPieces = pieces.length / 20;

console.log("TOTAL PIECES:", numPieces);

const pieceHashes = [];

for (let i = 0; i < numPieces; i++) {
  const hash = pieces.slice(i * 20, (i + 1) * 20); // every 20 bytes
  pieceHashes.push(hash);
  console.log(`PIECE ${i}:`, Buffer.from(hash).toString('hex'));
}

const peerId = Buffer.from("-JS0001-" + crypto.randomBytes(12).toString("hex").slice(0, 12));

// url encode raw bytes
function encode(buf) {
  return [...buf].map(b => "%" + b.toString(16).padStart(2, "0")).join("");
}

// ---- BUILD ANNOUNCE URL ----
const url =
  decorder.decode(torrent.announce) +
  `?info_hash=${encode(infoHash)}` +
  `&peer_id=${encode(peerId)}` +
  `&port=5881` +
  `&uploaded=0` +
  `&downloaded=0` +
  `&left=${torrent.info.length}` +
  `&compact=1`;

console.log("ANNOUNCE URL:", url);

// ---- SEND REQUEST (GET) ----
const res = await fetch(url);
const buf = Buffer.from(await res.arrayBuffer());

// ---- DECODE TRACKER RESPONSE ----
const trackerResponse = bencode.decode(buf);

if (trackerResponse['failure reason']) {
  console.log("TRACKER ERROR:", Buffer.from(trackerResponse['failure reason']).toString());
} else {
  console.log("SUCCESS:", trackerResponse);
  console.log("NUM PEERS:", trackerResponse.peers.length / 6);
  console.log("PEERS:", decodePeer(Buffer.from(trackerResponse.peers)));
}

function decodePeer(buf) {
  const peers = []
  for (let i = 0; i < buf.length; i+=6){
    const ip = buf[i] + "." +
      buf[i + 1] + "." +
      buf[i + 2] + "." +
      buf[i + 3];

    const port = buf.readUInt16BE(i+4);
    peers.push({ip, port});
  }
  return peers;
}

function getRandomPeer(max) {
  return Math.floor(Math.random() * max);
}

const randomPeerId = getRandomPeer(trackerResponse.peers.length / 6);
const randomPeer = decodePeer(Buffer.from(trackerResponse.peers)).at(randomPeerId);
console.log("PEER:", randomPeer);

const options = new Uint8Array([2])

const socket = net.createConnection(
  { host: randomPeer.ip, port: randomPeer.port },
  () => {
    console.log("CONNECTED TO PEER ");

    // send something
    const handshake = buildHandshake(infoHash, peerId);
    console.log("HANDSHAKE:" + handshake);
    socket.write(
      handshake
    );

    sendInterested(socket);
  }
);

socket.on("data", data => {
  console.log("RECEIVED LEN:", data.length);
  console.log("RECEIVED:", data);
  const id = data.readUint8(4);
  console.log("ID", id)
  if (id === 1) {
    console.log("UNCHOKED - SEND REQUEST")
    requestBlock(0);
  } else if (id === 7) {
    // piece message for each block
    const pieceIndex = data.readUInt32BE(5);
    const begin = data.readUInt32BE(9);
    const block = data.slice(13);
    console.log("BLOCK RECEIVED:", pieceIndex, begin, block.length);

    // store block in correct position
    block.copy(pieceBuffer, begin);
    received += block.length;

    // request next block
    if (received < pieceSize) {
      requestBlock(begin + block.length);
    } else {
      console.log("PIECE COMPLETE");

      verifyAndSavePiece();
    }
  }

  // -- Peer message -- //
  // format: 00 00 00 00 00 (5 bytes in hex)
  
});

socket.on("error", err => {
  console.log("ERROR:", err.message);
});

function buildHandshake(infoHash, peerId) {
  const buf = Buffer.alloc(68);

  buf.writeUInt8(19, 0);                           
  buf.write("BitTorrent protocol", 1);             
  buf.fill(0, 20, 28);                              
  infoHash.copy(buf, 28);                           
  peerId.copy(buf, 48);                             

  return buf;
}

function sendInterested(socket) {
  const buf = Buffer.alloc(5);

  buf.writeUInt32BE(1, 0); // length = 1
  buf.writeUInt8(2, 4);   // id = 2 (interested)

  socket.write(buf);
  console.log("SENT INTERESTED");
}

function buildRequest(pieceIndex, begin, length) {
  const buf = Buffer.alloc(17);

  buf.writeUInt32BE(13, 0);        // message length
  buf.writeUInt8(6, 4);            // id = request
  buf.writeUInt32BE(pieceIndex, 5); // index
  buf.writeUInt32BE(begin, 9); // begin
  buf.writeUInt32BE(length, 13); // length

  return buf;
}

function requestBlock(begin) {
  const remaining = pieceSize - begin;
  const length = Math.min(BLOCK_SIZE, remaining);

  const req = buildRequest(currentPiece, begin, length);
  socket.write(req);

  console.log("REQUEST:", currentPiece, begin, length);
}

function verifyAndSavePiece() {
  const expectedHash = pieceHashes[currentPiece];

  const actualHash = crypto
    .createHash("sha1")
    .update(pieceBuffer)
    .digest();

  if (!actualHash.equals(Buffer.from(expectedHash))) {
    console.log("HASH MISMATCH — PIECE CORRUPTED");
    socket.destroy();
    return;
  }

  console.log("HASH OK — WRITING PIECE");

  fs.writeFileSync("piece0.bin", pieceBuffer);
  console.log("SAVED piece0.bin");

  socket.end();
}
