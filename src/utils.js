"use strict"; // cannot use undeclared var

import crypto from 'crypto';

let peer_id = null;
// https://www.bittorrent.org/beps/bep_0020.html
// Follow this convention for peer id

export function genPeerID() {
    if (!peer_id) {
        peer_id = crypto.randomBytes(20);
        Buffer.from('-BA0001-').copy(peer_id, 0);
        // BA is my name and 0001 is version number 
    }
    return peer_id;
}

/*
    CJS: Common JS
    - Synchronously := block execution until the module is loaded => suitable for server-side

    MJS: ES Modules
    - Asynchronously := dynamic loading => for client-side
*/