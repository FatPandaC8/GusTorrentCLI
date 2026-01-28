import bencode from "bencode";

function decodePeers(buf) {
  const peers = [];
  for (let i = 0; i < buf.length; i += 6) {
    peers.push({
      ip: `${buf[i]}.${buf[i+1]}.${buf[i+2]}.${buf[i+3]}`,
      port: buf.readUInt16BE(i + 4)
    });
  }
  return peers;
}

function encode(buf) {
  return [...buf].map(b => "%" + b.toString(16).padStart(2, "0")).join("");
}

export async function getPeersHTTP(torrent, peerId) {
    const announce = Buffer
        .from(torrent.announce)
        .toString("utf8");

    const url =
        announce +
        `?info_hash=${encode(torrent.infoHash)}` +
        `&peer_id=${encode(peerId)}` +
        `&port=6881` +
        `&uploaded=0&downloaded=0` +
        `&left=${torrent.info.length}` +
        `&compact=1`;

    const res = await fetch(url);
    const buf = Buffer.from(await res.arrayBuffer());
    const response = bencode.decode(buf);

    if (response['failure reason']) {
    throw new Error(response['failure reason'].toString());
    }

    return decodePeers(Buffer.from(response.peers));
}
