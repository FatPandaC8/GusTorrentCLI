"use strict";

import bencode from 'bencode';
import crypto from 'crypto';
import fs from 'fs';

export const BLOCK_LENGTH = Math.pow(2, 14);

export function infoHash(torrent) {
  // to uniquely identify the torrent
  const info = bencode.encode(torrent.info);
  return crypto.createHash("sha1").update(info).digest();
}

export function size(torrent) {
  const fileSize = torrent.info.files
    ? torrent.info.files
      .map((file) => file.length)
      .reduce((acc, curr) => acc + curr)
    : torrent.info.length;

  return fileSize;
}

export function fileSize(torrent) {
  return size(torrent);
}

export function open(filePath) {
  const torrent = bencode.decode(fs.readFileSync(filePath));
  return torrent;
}

export function pieceLen(torrent, pieceIndex) {
  const totalLength = Number(size(torrent));
  const pieceLength = torrent.info["piece length"];

  const lastPieceLength = totalLength % pieceLength;
  const lastPieceIndex = Math.floor(totalLength / pieceLength);

  return lastPieceIndex === pieceIndex ? lastPieceLength : pieceLength;
}

export function blocksPerPiece(torrent, pieceIndex) {
  const pieceLength = pieceLen(torrent, pieceIndex);
  return Math.ceil(pieceLength / BLOCK_LENGTH);
}

export function blockLen(torrent, pieceIndex, blockIndex) {
  const pieceLength = pieceLen(torrent, pieceIndex);
  const lastBlockLength = pieceLength % BLOCK_LENGTH;
  const lastBlockIndex = Math.floor(pieceLength / BLOCK_LENGTH);

  return blockIndex === lastBlockIndex ? lastBlockLength : BLOCK_LENGTH;
}

export function numPieces(torrent) {
  return torrent.info.pieces.length / 20;
}