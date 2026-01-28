"use strict";

import bencode from 'bencode';
import crypto from 'crypto';
import fs from 'fs';

export function infoHash(torrent) {
  // to uniquely identify the torrent
  const info = bencode.encode(torrent.info);
  return crypto.createHash("sha1").update(info).digest();
}

export function fileSize(torrent) {
  const fileSize = torrent.info.files
    ? torrent.info.files
      .map((file) => file.length)
      .reduce((acc, curr) => acc + curr) // iterate over an array and condense it into a single value
    : torrent.info.length;

  return fileSize;
}

export function open(filePath) {
  const torrent = bencode.decode(fs.readFileSync(filePath));
  return torrent;
}