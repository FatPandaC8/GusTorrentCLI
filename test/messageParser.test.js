import { MessageParser } from "../src/peer/MessageParser.js";
import * as message from "../src/tracker/message.js";

test("MessageParser handles partial messages", () => {
  const parser = new MessageParser();
  const buf = message.buildKeepAlive(); // 00 00 00 00

  // split the message into halves
  const part1 = buf.slice(0, 2);
  const part2 = buf.slice(2);

  expect(parser.push(part1).length).toBe(0); // as the message is valid only if it's >= 4 bytes
  expect(parser.push(part2).length).toBe(1);
});
