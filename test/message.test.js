import * as message from '../src/tracker/message.js';

describe("buid request", () => {
    test("encodes request correctly", () => {
        const req = {
            index: 1,
            begin: 16384,
            length: 16384
        }

        const buf = message.buildRequest(req);

        expect(buf.length).toBe(17);
        expect(buf.readUInt32BE(0)).toBe(13);
        expect(buf.readUInt8(4)).toBe(6);
        expect(buf.readUInt32BE(5)).toBe(1);
        expect(buf.readUInt32BE(9)).toBe(16384);
        expect(buf.readUInt32BE(13)).toBe(16384);
    })
})