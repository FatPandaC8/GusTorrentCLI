import net from "net";
import * as message from "../tracker/message.js";
import { MessageParser } from "./MessageParser.js";
import { numPieces } from "../torrent-parser.js";

export class PeerConnection {
  constructor(peer, torrent, pieceManager) {
    this.peer = peer;
    this.torrent = torrent;
    this.pm = pieceManager;
    this.parser = new MessageParser();
    this.bitfield = null;
    this.handshakeReceived = false;
    this.choked = true;
    this.currentPiece = null;
    this.pieceState = null;
    this.availablePieces = new Set();
  }

  connect() {
    this.socket = net.createConnection(
      { host: this.peer.ip, port: this.peer.port, timeout: 10000 },
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
      const piece = this.pm.getPieceForPeer(this.availablePieces);
      if (piece === null) return;
      this.currentPiece = piece;
      this.pieceState = this.pm.startPiece(piece);
      this.request(0);
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
      const pieceIndex = payload.readUInt32BE(0);
      this.availablePieces.add(pieceIndex);
      return;
    }

    // Bitfield
    if (id === 5) {
      console.log("Received bitfield");
      // tell which pieces does each peer have
      this.availablePieces = this.hasWhatPiece(payload, numPieces(this.torrent));
      this.bitfield = payload;
      return;
    }

    // Piece
    if (id === 7) {
      payload.block.copy(this.pieceState.buffer, payload.begin);
      this.pieceState.received += payload.block.length;

      if (this.pieceState.received >= this.pieceState.size) {
        const ok = this.pm.completePiece(
          this.currentPiece,
          this.pieceState.buffer
        );

        if (!ok) {
          console.log(`Piece ${this.currentPiece} failed hash`);
        } else {
          console.log(`Piece ${this.currentPiece} completed`);
        }

        this.currentPiece = null;
        this.pieceState = null;

        const next = this.pm.getPieceForPeer(this.availablePieces);
        if (next !== null) {
          this.currentPiece = next;
          this.pieceState = this.pm.startPiece(next);
          this.request(0);
        }
      } else {
        this.request(this.pieceState.received);
      }
    }
  }

  request(begin) {
    if (this.choked || !this.pieceState) return;

    const remaining = this.pieceState.size - begin;
    const length = Math.min(16384, remaining);

    this.socket.write(message.buildRequest({
      index: this.currentPiece,
      begin,
      length
    }));
  }

  hasWhatPiece(bitfield, numPieces) {
    const availablePieces = new Set();
    for (let piece = 0; piece < numPieces; piece++) {

      // each byte holds 8 pieces
      // pieces 0-7 → byte 0
      // pieces 8-15 → byte 1
      const byteIndex = Math.floor(piece / 8);

      // which bit inside that byte (left → right)
      const bitIndex = 7 - (piece % 8);

      const byte = bitfield[byteIndex];
      const hasPiece = (byte & (1 << bitIndex)) !== 0;
      if (hasPiece) {
        availablePieces.add(piece);
      }
    }
    return availablePieces;
  }
}