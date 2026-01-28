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

    // Try connecting to peers (try multiple if first fails)
    let connected = false;
    for (let i = 0; i < Math.min(5, peers.length) && !connected; i++) {
      const peer = peers[i];
      console.log(`\nTrying peer ${i + 1}: ${peer.ip}:${peer.port}`);
      
      try {
        const conn = new PeerConnection(peer, torrent, pm);
        conn.connect();
        connected = true;
        
        // Keep the process alive
        process.on('SIGINT', () => {
          console.log("\nShutting down...");
          if (conn.socket) {
            conn.socket.end();
          }
          process.exit(0);
        });
        
      } catch (err) {
        console.error(`Failed to connect to peer: ${err.message}`);
      }
    }

    if (!connected) {
      console.error("Could not connect to any peers");
      process.exit(1);
    }

  } catch (err) {
    console.error("Error:", err.message);
    console.error(err.stack);
    process.exit(1);
  }
}

main();