import { BLOCK_LENGTH, blockLen, infoHash, pieceLen, size } from "../src/torrent-parser";

test("infoHash is deterministic", () => {
  const torrent = {
    info: {
      name: "test",
      length: 123,
      "piece length": 16,
      pieces: Buffer.alloc(20)
    }
  };

  const h1 = infoHash(torrent);
  const h2 = infoHash(torrent);

  expect(h1.equals(h2)).toBe(true);
  expect(h1.length).toBe(20);
});

test("size handles single-file torrent", () => {
  const torrent = { info: { length: 1000 } };
  expect(size(torrent)).toBe(1000);
});

test("size handles multi-file torrent", () => {
  const torrent = {
    info: {
      files: [{ length: 400 }, { length: 600 }]
    }
  };
  expect(size(torrent)).toBe(1000);
});

test("pieceLen handles last piece correctly", () => {
  const torrent = {
    info: {
      length: 100,
      "piece length": 32
    }
  };

  expect(pieceLen(torrent, 0)).toBe(32);
  expect(pieceLen(torrent, 1)).toBe(32);
  expect(pieceLen(torrent, 2)).toBe(32);
  expect(pieceLen(torrent, 3)).toBe(4); // last piece
});

test("blockLen handles last block", () => {
  const torrent = {
    info: {
      length: 40_000,
      "piece length": 40_000
    }
  };

  const lastBlock = Math.floor(40_000 / BLOCK_LENGTH);
  const len = blockLen(torrent, 0, lastBlock);

  expect(len).toBe(40_000 % BLOCK_LENGTH);
});
