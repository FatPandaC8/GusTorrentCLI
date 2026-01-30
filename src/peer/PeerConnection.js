import net from "net";
import * as message from "../tracker/message.js";
import { MessageParser } from "./MessageParser.js";

export class PeerConnection {
  constructor(peer, torrent, pieceManager) {
    this.peer = peer;
    this.torrent = torrent;
    this.pm = pieceManager;
    this.parser = new MessageParser();
    this.bitfield = null;
    this.handshakeReceived = false;
    this.choked = true;
    this.maxPipelined = 5;
    this.inflight = new Map();
    this.stats = {
      downloaded: 0,
      startTime: Date.now()
    };
  }

  connect() {
    this.socket = net.createConnection(
      { host: this.peer.ip, port: this.peer.port },
      () => {
        console.log("Connected to: " + this.peer.ip + ":" + this.peer.port);
        this.socket.write(message.buildHandshake(this.torrent));
      }
    );

    this.socket.on("data", data => {
      this.onData(data);
    });
    
    this.socket.on("error", err => {
      console.log("Connection error:", err.message);
    });
    
    this.socket.on("end", () => {
      console.log("Connection closed");
    });
  }

  onData(data) {
    // Handle handshake first (68 bytes)
    // handshake: <pstrlen><pstr><reserved><info_hash><peer_id>
    //     pstrlen: string length of <pstr>, as a single raw byte
    //     pstr: string identifier of the protocol
    //     reserved: eight (8) reserved bytes. All current implementations use all zeroes. Each bit in these bytes can be used to change the behavior of the protocol. An email from Bram suggests that trailing bits should be used first, so that leading bits may be used to change the meaning of trailing bits.
    //     info_hash: 20-byte SHA1 hash of the info key in the metainfo file. This is the same info_hash that is transmitted in tracker requests.
    //     peer_id: 20-byte string used as a unique ID for the client. This is usually the same peer_id that is transmitted in tracker requests (but not always e.g. an anonymity option in Azureus).
    // In version 1.0 of the BitTorrent protocol, pstrlen = 19, and pstr = "BitTorrent protocol". 
    
    if (!this.handshakeReceived && data.length >= 68) {
      console.log("Handshake received");
      this.handshakeReceived = true;
      
      // Send interested message after handshake
      this.socket.write(message.buildInterest());
      
      // Process remaining data if any
      if (data.length > 68) {
        const remaining = data.slice(68);
        const messages = this.parser.push(remaining);
        for (const msg of messages) {
          this.handle(msg);
        }
      }
      return;
    }
    
    // If handshake not complete yet, accumulate data
    if (!this.handshakeReceived) {
      return;
    }

    // Parse and handle messages
    const messages = this.parser.push(data);
    for (const msg of messages) {
      this.handle(msg);
    }
  }

  handle({ id, payload }) {
    // Keep-alive
    if (id === null) {
      return;
    }

    // Choke
    if (id === 0) {
      console.log("Choked by peer");
      this.choked = true;
      return;
    }

    // Unchoke
    if (id === 1) {
      console.log("Unchoked by peer");
      this.choked = false;
      this.fillPipeline();
      return;
    }

    // Interested
    if (id === 2) {
      console.log("Peer is interested");
      return;
    }

    // Not interested
    if (id === 3) {
      console.log("Peer is not interested");
      return;
    }

    // Have
    if (id === 4) {
      console.log("Peer has piece:", payload);
      return;
    }

    // Bitfield
    if (id === 5) {
      console.log("Received bitfield");
      this.bitfield = payload;
      return;
    }

    // Piece
    if (id === 7) {
      this.inflight.delete(payload.begin);
      this.pm.addBlock(payload.begin, payload.block);
      this.stats.downloaded += payload.block.length;
      this.logSpeed();
      
      this.fillPipeline();

      if (this.pm.isComplete()) {
        if (!this.pm.verify()) {
          console.error("Piece corrupted!");
          this.inflight.clear();
          this.pm.resetPiece();
          this.fillPipeline();
          return;
        }
        
        console.log(`Piece ${this.pm.currentPiece} completed and verified`);
        this.pm.save();
        
        // Move to next piece
        if (this.pm.moveToNextPiece()) {
          console.log(`Starting piece ${this.pm.currentPiece}`);
          this.inflight.clear();
          this.fillPipeline();
        } else {
          console.log("Download complete!");
          this.socket.end();
        }
      }
    }
  }

  fillPipeline() {
    if (this.choked) {
      console.log("Cannot request: still choked");
      return;
    }

    while (this.inflight.size < this.maxPipelined) {
      const req = this.pm.nextRequest();
      if (!req) return;
      this.socket.write(message.buildRequest(req));
      this.inflight.set(req.begin, req.length);
    }
  }

  logSpeed() {
    const elapsed = (Date.now() - this.stats.startTime) / 1000;
    const kbps = (this.stats.downloaded / 1024 / elapsed).toFixed(2);
    console.log(`Speed: ${kbps} KB/s`);
  }
}