"use strict";

import bencode from 'bencode';
import crypto from 'crypto';

export function infoHash(torrent) {
  // to uniquely identify the torrent
  const info = bencode.encode(torrent.info);
  return crypto.createHash("sha1").update(info).digest();
}