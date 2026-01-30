import crypto from "crypto";
import fs from "fs";
import { numPieces } from "../torrent-parser.js";

// The reality is near all clients will now use 2^14 (16KB) requests. 
// Due to clients that enforce that size, it is recommended that implementations make requests of that size
const BLOCK_SIZE = Math.pow(2, 14);

export class PieceManager {
  constructor(torrent) {
    this.torrent = torrent;
    this.pieceLength = torrent.info['piece length'];
    this.pieces = torrent.info.pieces;
    // console.log(pieces);
    // string consisting of the concatenation of all 20-byte SHA1 hash values, one per piece (byte string, i.e. not urlencoded)
    this.totalPieces = this.pieces.length / 20;
    this.trackingPieces = new Map();
    for (let i = 0; i < numPieces(torrent); i++) {
      this.trackingPieces.set(i, 'missing');
    }
    console.log(this.trackingPieces);
  }

  startPiece(pieceIndex) {
    const pieceSize = Math.min(
      this.pieceLength,
      this.torrent.info.length - pieceIndex * this.pieceLength
    );

    return {
      index: pieceIndex,
      size: pieceSize,
      received: 0,
      buffer: Buffer.alloc(pieceSize)
    };
  }

  completePiece(pieceIndex, buffer) {
    const expected = this.pieces.slice(pieceIndex * 20, (pieceIndex + 1) * 20);
    const actual = crypto.createHash("sha1").update(buffer).digest();

    if (!actual.equals(expected)) {
      this.trackingPieces.set(pieceIndex, 'missing');
      return false;
    }

    fs.writeFileSync(`piece${pieceIndex}.bin`, buffer);
    this.trackingPieces.set(pieceIndex, 'completed');
    return true;
  }

  nextRequest(begin) {
    const remaining = this.pieceSize - begin;
    return {
      index: this.currentPiece,
      begin,
      length: Math.min(BLOCK_SIZE, remaining)
    };
  }

  addBlock(begin, block) {
    block.copy(this.buffer, begin);
    this.received += block.length;
  }

  isComplete() {
    return this.received >= this.pieceSize;
  }

  verify() {
    const expected = this.pieces.slice(
      this.currentPiece * 20,
      (this.currentPiece + 1) * 20
    );

    const actual = crypto
      .createHash("sha1")
      .update(this.buffer)
      .digest();

    return actual.equals(expected);
  }

  save() {
    const filename = `piece${this.currentPiece}.bin`;
    fs.writeFileSync(filename, this.buffer);
    this.downloadedPieces.push(this.currentPiece);
    console.log(`Saved ${filename} (${this.downloadedPieces.length}/${this.totalPieces})`);
  }

  getPieceForPeer(availablePieces) {
    for (const piece of availablePieces) {
      if (this.trackingPieces.get(piece) === 'missing') {
        this.trackingPieces.set(piece, 'downloading');
        return piece;
      }
    }
    return null;
  }
}