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
      this.request(0);
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
      this.pm.addBlock(payload.begin, payload.block);

      if (this.pm.isComplete()) {
        if (!this.pm.verify()) {
          console.error("Piece corrupted!");
          this.socket.end();
          return;
        }
        
        console.log(`Piece ${this.pm.currentPiece} completed and verified`);
        this.pm.save();
        
        // Move to next piece
        if (this.pm.moveToNextPiece()) {
          console.log(`Starting piece ${this.pm.currentPiece}`);
          this.request(0);
        } else {
          console.log("Download complete!");
          this.socket.end();
        }
      } else {
        this.request(payload.begin + payload.block.length);
      }
    }
  }

  request(begin) {
    if (this.choked) {
      console.log("Cannot request: still choked");
      return;
    }
    
    const req = this.pm.nextRequest(begin);
    this.socket.write(message.buildRequest(req));
  }
}