import net from "net";
import * as message from "../tracker/message.js";
import { MessageParser } from "./MessageParser.js"; // TCP-safe parser

export class PeerConnection {
  constructor(peer, torrent, pieceManager, bar) {
    this.peer = peer;
    this.torrent = torrent;
    this.pm = pieceManager;
    this.bar = bar;
    this.parser = new MessageParser();
    this.bitfield = null;
  }

  connect() {
    this.socket = net.createConnection(
      { host: this.peer.ip, port: this.peer.port },
      () => {
        console.log("Connected to: " + this.peer.ip + ": " + this.peer.port);
        this.socket.write(message.buildHandshake(this.torrent));
        this.socket.write(message.buildInterest());
      }
    );

    this.socket.on("data", data => {
        this.onData(data);
    });
    this.socket.on("error", err => console.log(err.message));
}

onData(data) {
    console.log(data);
    const messages = this.parser.push(data);
    for (const msg of messages) {
      this.handle(msg);
    }
  }

  handle({ id, payload }) {
    if (id === 1) { // unchoke
      this.request(0);
    }

    if (id === 5) {
        this.bitfield = payload;
        return;
    }

    if (id === 7) { // piece
      this.pm.addBlock(payload.begin, payload.block);
      this.bar.tick(payload.block.length);

      if (this.pm.isComplete()) {
        if (!this.pm.verify()) {
          throw new Error("Piece corrupted");
        }
        this.pm.save();
        this.socket.end();
      } else {
        this.request(payload.begin + payload.block.length);
      }
    }
  }

  request(begin) {
    const req = this.pm.nextRequest(begin);
    this.socket.write(message.buildRequest(req));
  }
}
