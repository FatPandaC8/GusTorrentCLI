// All of the remaining messages in the protocol take the form of <length prefix><message ID><payload>. 
// The length prefix is a four byte big-endian value. 
// The message ID is a single decimal byte. 
// The payload is message dependent.

export class MessageParser {
  constructor() {
    this.buffer = Buffer.alloc(0);
  }

  push(data) {
    /*
      Sometimes, message does not come in full => need to concat then can get the whole msg
    */
    this.buffer = Buffer.concat([this.buffer, data]);
    const messages = [];

    while (this.buffer.length >= 4) {
      const length = this.buffer.readUInt32BE(0);

      // Not enough data yet
      if (this.buffer.length < length + 4) break;

      // Get the message from the front
      const msg = this.buffer.slice(0, length + 4); 

      // Throw what used, keep leftover
      this.buffer = this.buffer.slice(length + 4);

      messages.push(parseMessage(msg));
    }

    return messages;
  }
}

function parseMessage(msg) {
  const size = msg.readUInt32BE(0);

  // keep-alive
  if (size === 0) {
    return { id: null, payload: null };
  }

  const id = msg.readUInt8(4);
  let payload = msg.slice(5);

  // request / piece / cancel
  if (id === 6 || id === 7 || id === 8) {
    const index = payload.readUInt32BE(0);
    const begin = payload.readUInt32BE(4);

    if (id === 7) {
      payload = {
        index,
        begin,
        block: payload.slice(8)
      };
    } else {
      payload = {
        index,
        begin,
        length: payload.readUInt32BE(8)
      };
    }
  }

  return { size, id, payload };
}