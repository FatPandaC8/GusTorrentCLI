import crypto from "crypto";
import fs from "fs";

// The reality is near all clients will now use 2^14 (16KB) requests. 
// Due to clients that enforce that size, it is recommended that implementations make requests of that size
const BLOCK_SIZE = Math.pow(2, 14);

export class PieceManager {
  constructor(torrent) {
    this.torrent = torrent;
    this.currentPiece = 0;
    this.pieceLength = torrent.info['piece length'];
    this.pieces = torrent.info.pieces;
    // string consisting of the concatenation of all 20-byte SHA1 hash values, one per piece (byte string, i.e. not urlencoded)
    this.totalPieces = this.pieces.length / 20;
    this.downloadedPieces = [];
    this.resetPiece();
  }

  resetPiece() {
    this.nextOffset = 0;
    this.received = 0;
    this.pieceSize = Math.min(
      this.pieceLength,
      this.torrent.info.length - this.currentPiece * this.pieceLength
    );
    this.buffer = Buffer.alloc(this.pieceSize);
  }

  nextRequest() {
    if (this.nextOffset >= this.pieceSize) return null;
    const begin = this.nextOffset;
    const length = Math.min(BLOCK_SIZE, this.pieceSize - begin);
  
    this.nextOffset += length;

    return {
      index: this.currentPiece,
      begin,
      length
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

  moveToNextPiece() {
    this.currentPiece++;
    if (this.currentPiece < this.totalPieces) {
      this.resetPiece();
      return true;
    }
    return false;
  }
}