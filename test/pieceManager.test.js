import { PieceManager } from "../src/torrent/PieceManager.js";
import crypto from "crypto";

test("PieceManager completes and verifies piece", () => {
  const data = Buffer.alloc(32768, 1);
  const hash = crypto.createHash("sha1").update(data).digest();

  const torrent = {
    info: {
      length: 32768,
      "piece length": 32768,
      pieces: hash
    }
  };

  const pm = new PieceManager(torrent);

  pm.addBlock(0, data.slice(0, 16384));
  pm.addBlock(16384, data.slice(16384));

  expect(pm.isComplete()).toBe(true);
  expect(pm.verify()).toBe(true);
});
