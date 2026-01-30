import net from "net";
import * as message from "../tracker/message.js";
import { MessageParser } from "./MessageParser.js";

export class PeerConnection {
  constructor(peer, torrent, pieceManager) {
    this.peer = peer;
    this.torrent = torrent;
    this.pm = pieceManager;
    this.parser = new MessageParser();

    this.handshakeReceived = false;
    this.choked = true;

    this.maxPipelined = 5;
    this.inflight = new Map();

    this.stats = {
      downloaded: 0,
      startTime: Date.now()
    };

    // progress bar
    this.totalSize = torrent.info.length;
    this.barWidth = 30;
    this.lastRender = 0;
    this.downloading = true;
    this.logs = [];
  }

  connect() {
    this.socket = net.createConnection(
      { host: this.peer.ip, port: this.peer.port },
      () => {
        this.log(`Connected to ${this.peer.ip}:${this.peer.port}`);
        this.socket.write(message.buildHandshake(this.torrent));
      }
    );

    this.socket.on("data", data => this.onData(data));
    this.socket.on("error", err => this.log(`Connection error: ${err.message}`));
  }

  onData(data) {
    if (!this.handshakeReceived && data.length >= 68) {
      this.handshakeReceived = true;
      this.log("Handshake received");

      this.socket.write(message.buildInterest());

      if (data.length > 68) {
        for (const msg of this.parser.push(data.slice(68))) {
          this.handle(msg);
        }
      }
      return;
    }

    if (!this.handshakeReceived) return;

    for (const msg of this.parser.push(data)) {
      this.handle(msg);
    }
  }

  handle({ id, payload }) {
    if (id === null) return;

    if (id === 0) {
      this.choked = true;
      this.log("Choked by peer");
      return;
    }

    if (id === 1) {
      this.choked = false;
      this.log("Unchoked by peer");
      this.fillPipeline();
      return;
    }

    if (id === 7) {
      this.inflight.delete(payload.begin);
      this.pm.addBlock(payload.begin, payload.block);
      this.stats.downloaded += payload.block.length;

      this.renderProgress();
      this.fillPipeline();

      if (this.pm.isComplete()) {
        if (!this.pm.verify()) {
          this.log("Piece corrupted, retrying");
          this.inflight.clear();
          this.pm.resetPiece();
          this.fillPipeline();
          return;
        }

        this.log(`Piece ${this.pm.currentPiece} completed`);
        this.pm.save();

        if (this.pm.moveToNextPiece()) {
          this.log(`Starting piece ${this.pm.currentPiece}`);
          this.inflight.clear();
          this.fillPipeline();
        } else {
          this.finish();
        }
      }
    }
  }

  fillPipeline() {
    if (this.choked) return;

    while (this.inflight.size < this.maxPipelined) {
      const req = this.pm.nextRequest();
      if (!req) return;

      this.socket.write(message.buildRequest(req));
      this.inflight.set(req.begin, req.length);
    }
  }

  renderProgress(force = false) {
    if (!process.stdout.isTTY) return;

    const now = Date.now();
    if (!force && now - this.lastRender < 200) return;
    this.lastRender = now;

    const percent = this.stats.downloaded / this.totalSize;
    const filled = Math.floor(this.barWidth * percent);
    const empty = this.barWidth - filled;

    const bar = "=".repeat(filled) + " ".repeat(empty);

    const elapsed = (now - this.stats.startTime) / 1000;
    const speed =
      elapsed > 0
        ? (this.stats.downloaded / 1024 / elapsed).toFixed(1)
        : "0.0";

    process.stdout.write(
      `\r\x1b[K[${bar}] ${(percent * 100).toFixed(1)}% | ${speed} KB/s`
    );
  }

  log(msg) {
    if (this.downloading) {
      this.logs.push(msg);
    } else {
      console.log(msg);
    }
  }

  finish() {
    this.downloading = false;

    // clear progress bar
    process.stdout.write("\r\x1b[K");

    // flush buffered logs
    for (const line of this.logs) {
      console.log(line);
    }

    console.log("Download completed");
    this.socket.end();
  }
}
