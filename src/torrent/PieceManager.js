import crypto from "crypto";
import fs from "fs";

const BLOCK_SIZE = 16384;

export class PieceManager {
  constructor(torrent) {
    this.torrent = torrent;
    this.currentPiece = 0;
    this.pieceLength = torrent.info['piece length'];
    this.pieces = torrent.info.pieces;
    this.totalPieces = this.pieces.length / 20;
    this.downloadedPieces = [];
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

  getProgress() {
    return {
      current: this.currentPiece,
      total: this.totalPieces,
      percentage: ((this.downloadedPieces.length / this.totalPieces) * 100).toFixed(2)
    };
  }
}