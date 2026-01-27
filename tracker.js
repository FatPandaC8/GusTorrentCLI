// import dgram from 'dgram';
// import { Buffer } from 'buffer';

// module.exports.getPeers = (torrent, callback) => {
//     const socket = dgram.createSocket('udp4');
//     const url = torrent.announce.toString('utf-8');

//     udpSend(socket, buildConnReq(), url);
//     socket.on('message', response => {
//         if (respType(response) === 'connect') {
//             const connResp = parseConnResp(response);
//             const announceReq = buildAnnounceReq(connResp.connectionId);
//             udpSend(socket, announceReq, url);
//         } else if (respType(response) === 'announce') {
//             const announceReq = parseAnnounceResp(response);
//             callback(announceReq.peers)
//         }
//     });
// };

// function respType(resp) {
//   // ...
// }

// function buildConnReq() {
//   // ...
// }

// function parseConnResp(resp) {
//   // ...
// }

// function buildAnnounceReq(connId) {
//   // ...
// }

// function parseAnnounceResp(resp) {
//   // ...
// }