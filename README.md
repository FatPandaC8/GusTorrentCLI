# GusTorrentCLI

An experimental, minimal BitTorrent client implemented in Node.js. This repository provides a compact command-line implementation of the BitTorrent wire protocol: it parses .torrent files, contacts an HTTP tracker for peers, connects to peers over TCP, performs the BitTorrent handshake, requests piece blocks, verifies pieces using SHA-1, and writes downloaded pieces to disk. The project is intentionally small and includes many areas marked for improvement.

----

## Quick facts

- Language: JavaScript (ES modules)
- Primary modules: torrent parsing, HTTP tracker client, BitTorrent wire message builders/parsers, per-peer TCP connection handler, piece manager
- Current behaviour: single-peer download (tries peers sequentially and uses the first successful connection), saves downloaded pieces as `piece<N>.bin` files in the current working directory.
- Project description (from repo): "An implementation of torrent protocol, need a lot of improvement"

----

## Requirements

- Node.js 14+ (ES module support and modern Buffer/crypto APIs)
- npm (for installing dependencies listed below)

External dependencies used in source:

- bencode — parse/encode ".torrent" and tracker bencoded responses

If this repo lacks a `package.json`, install dependencies manually:

```bash
npm install bencode
```

----

## Usage

Basic usage from repository root:

```bash
node index.js <path-to-torrent-file>
```

Example:

```bash
node index.js ubuntu-22.04-desktop.iso.torrent
```

What to expect:

- The CLI prints parsed torrent metadata (name, total size, piece count).
- The client contacts the first HTTP tracker listed in the torrent's `announce` field to get a compact peer list.
- It attempts to connect to peers (tries up to the first 5 peers by default), and uses the first peer that successfully completes a TCP handshake.
- Pieces are downloaded block-by-block and saved as `piece<N>.bin` files. There is currently no reassembly into the original file(s).

----

## Architecture and important files

This section explains the responsibilities of the main source files so you can quickly understand and extend the code.

- index.js — CLI entrypoint and download orchestration
  - Parses the torrent file, prints metadata, generates a peer ID, fetches peers from the tracker, creates a PieceManager, and opens a PeerConnection to a peer.

- src/torrent-parser.js — .torrent parsing and piece/block geometry helpers
  - Uses `bencode` and `crypto` to compute `infoHash`, determine piece lengths, block sizes, number of pieces and total torrent size.

- src/tracker/httpTracker.js — HTTP tracker client
  - Builds the tracker announce URL (compact mode), performs a fetch, decodes the bencoded response, and converts the compact peers blob into `{ ip, port }` entries.

- src/tracker/message.js — BitTorrent wire message builders and a small parser
  - Creates handshake, keep-alive, choke/unchoke, interested/uninterested, have, bitfield, request, piece, cancel and port messages with correct length-prefixing.

- src/peer/PeerConnection.js — Peer TCP connection and message handling
  - Manages a single TCP connection to a peer: sends the handshake, receives the peer's handshake/bitfield, sends `interested`, handles `unchoke`, requests blocks, and processes incoming `piece` messages to assemble and verify pieces via PieceManager.

- src/torrent/PieceManager.js — Piece buffer management, verification and saving
  - Buffers incoming blocks for the current piece, verifies the piece using SHA-1 against the hash list in the torrent, and writes completed pieces to disk as `piece<N>.bin`.

- src/utils.js — small utilities (peer ID generator, etc.)

Other referenced code (implement/verify these exist or update as needed):
- src/peer/MessageParser.js — must correctly parse and reassemble length-prefixed messages from the TCP stream and return message objects used by PeerConnection.

----

## Current limitations (important for contributors)

This project is an intentionally small, educational implementation. It is missing many features expected of a robust BitTorrent client:

- Single-peer focus: the client connects to a single peer (the first successful connection) and downloads pieces sequentially from that peer.
- No file reassembly: pieces are saved as individual `piece<N>.bin` files; the repository does not yet re-create the original filename(s) or directory structure from `torrent.info.files`.
- Minimal concurrency and pipelining: there is no parallel downloading from many peers, no request window/pipelining per peer, and no optimistic unchoke or rarest-first logic.
- Partial protocol features: UDP trackers, DHT, BEP-10 extension messages and encryption are not implemented.
- Basic error handling: limited retry/backoff, no peer blacklisting and no timeouts.
- No tests or CI configured in the repository (unit and integration tests recommended).

----

## Development notes and recommended improvements

If you want to contribute or extend GusTorrentCLI, here are prioritized suggestions with rationale:

1. Message framing and parser robustness
   - Ensure `src/peer/MessageParser.js` handles partial TCP frames, multiple messages per data event, and maintains internal buffering state between `socket.on('data')` events.

2. Multi-peer concurrency and pipelining
   - Connect to and manage multiple peers concurrently.
   - Implement a request pipeline (N outstanding requests per peer) to increase throughput.
   - Add a request slot manager and simple rate limiting.

3. File reassembly and disk layout
   - Rebuild the original single-file or multi-file layout from `torrent.info.files` and write blocks at correct offsets to the final files instead of producing piece files.

4. Piece selection strategies
   - Add rarest-first selection and an end-game mode to avoid long-tail stalls.

5. Tracker and peer improvements
   - Support UDP trackers, parse non-compact peer lists, and support IPv6 addresses if needed.

6. Resilience and observability
   - Add connection timeouts, retry/backoff, metrics for download speed and ETA, and more user-friendly progress output.

7. Tests and CI
   - Add unit tests for torrent parsing, piece verification, and message builders/parsers, and an integration test using a small local test torrent and a controlled peer server.

----

## Example developer workflow

Install dependencies (if package.json exists, run `npm install`):

```bash
npm install
```

Run the client:

```bash
node index.js <torrent-file>
```

Run tests (TBD): create tests under `test/` and add a `test` script to `package.json`.

----

## Contributing

Contributions are welcome. If you plan to change core protocol logic or introduce concurrency, please open an issue describing your plan so we can discuss design choices (request/pipeline windows, piece selection, on-disk layout). When submitting PRs:

- Keep changes small and focused (e.g., one PR for parser fixes, one PR for file reassembly).
- Add tests for new logic where practical.
- Update this README with any new runtime flags or behavior.

----

## License

This repository does not include an explicit license file. If you plan to publish or accept contributions, add a LICENSE file (for example, MIT) to clarify terms.

----

## Roadmap / TODO (short)
- [ ] Implement MessageParser that handles TCP stream framing robustly
- [ ] Implement file reassembly / write final files instead of piece files
- [ ] Support multiple concurrent peer connections & pipelining
- [ ] Add unit tests and CI configuration
- [ ] Implement UDP tracker support and non-compact peer decoding

----

If you want, I can:
- open a set of targeted issues for the TODOs above (ready-to-assign), or
- create a PR that implements one small improvement (for example, basic file reassembly for single-file torrents or a robust MessageParser), or
- provide a minimal package.json and npm scripts to make running and testing the project easier.

Tell me which of the above you want next and I will proceed.