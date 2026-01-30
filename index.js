import * as util from './src/utils.js';
import * as torrentParser from "./src/torrent-parser.js";
import { getPeersHTTP } from "./src/tracker/httpTracker.js";
import { PieceManager } from "./src/torrent/PieceManager.js";
import { PeerConnection } from "./src/peer/PeerConnection.js";

async function main() {
  try {
    // Check if torrent file is provided
    if (!process.argv[2]) {
      console.error("Usage: node index.js <torrent-file>");
      process.exit(1);
    }

    // Parse torrent file
    console.log("Parsing torrent file...");
    const torrent = torrentParser.open(process.argv[2]);
    torrent.infoHash = torrentParser.infoHash(torrent);

    console.log("Torrent info:");
    console.log("  Name:", torrent.info.name.toString());
    console.log("  Size:", torrentParser.size(torrent), "bytes");
    console.log("  Pieces:", torrentParser.numPieces(torrent));

    // Generate peer ID and get peers
    console.log("\nContacting tracker...");
    const peerId = util.genPeerID();
    const peers = await getPeersHTTP(torrent, peerId);
    
    if (!peers || peers.length === 0) {
      console.error("No peers found!");
      process.exit(1);
    }
    
    console.log(`Found ${peers.length} peers`);
    console.log(peers);

    // Initialize piece manager
    const pm = new PieceManager(torrent);
    const connections = [];

    for (let i = 0; i < Math.min(5, peers.length); i++) {
      const peer = peers[i];
      console.log(`Trying peer ${i + 1}: ${peer.ip}:${peer.port}`);
      
      try {
        const conn = new PeerConnection(peer, torrent, pm);
        conn.connect();
        connections.push(conn);

        // Keep the process alive
        process.on('SIGINT', () => {
          console.log("\nShutting down...");
          for (const conn of connections) {
            if (conn.socket) conn.socket.end();
          }
          process.exit(0);
        });
        
      } catch (err) {
        console.error(`Failed to connect to peer: ${err.message}`);
      }
    }
  } catch (err) {
    console.error("Error:", err.message);
    console.error(err.stack);
    process.exit(1);
  }
}

main();