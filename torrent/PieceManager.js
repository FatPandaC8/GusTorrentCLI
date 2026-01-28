import crypto from "crypto";
import fs from "fs";

const BLOCK_SIZE = 16384;

export class PieceManager {
  constructor(torrent) {
    this.torrent = torrent;
    this.currentPiece = 0;
    this.pieceLength = torrent.info['piece length'];
    this.pieces = torrent.info.pieces;
    this.resetPiece();
  }

  resetPiece() {
    this.received = 0;
    this.pieceSize = Math.min(
      this.pieceLength,
      this.torrent.info.length - this.currentPiece * this.pieceLength
    );
    this.buffer = Buffer.alloc(this.pieceSize);
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
    console.log("COMPLETED");
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
    fs.writeFileSync(`piece${this.currentPiece}.bin`, this.buffer);
  }
}