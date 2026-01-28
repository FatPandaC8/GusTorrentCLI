import * as util from './src/utils.js';
import * as torrentParser from "./src/torrent-parser.js";
import ProgressBar from "progress";
import { getPeersHTTP } from "./tracker/httpTracker.js";
import { PieceManager } from "./torrent/PieceManager.js";
import { PeerConnection } from "./peer/PeerConnection.js";

const torrent = torrentParser.open(process.argv[2]);
// ---- INFO HASH ----
torrent.infoHash = torrentParser.infoHash(torrent);

const peerId = util.genPeerID();
const peers = await getPeersHTTP(torrent, peerId);
const pm = new PieceManager(torrent);
const peer = peers[Math.floor(Math.random() * peers.length)]

const bar = new ProgressBar('Downloading [:bar] :percent :etas', {
  width: 20,
  total: torrent.info.length
});

const conn = new PeerConnection(peer, torrent, pm, bar);
conn.connect()