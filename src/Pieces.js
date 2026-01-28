'use strict';

import * as torrentParser from './torrent-parser.js';

export default class Pieces {
    constructor(torrent) {
        function calculateTotalPieces() {
            let totalPieces = 0;
            const nFiles = torrent.info.files.length; // may vary

            for (let i = 0; i < nFiles; i++) {
                const fileLength = torrent.info.files[i].length;
                const piecesInFile = Math.ceil(fileLength / torrent.info['piece length'])
                totalPieces += piecesInFile;
            }

            return totalPieces;
        }

        function buildPiecesArray() {
        // torrent.info.pieces is a buffer that contains 20-byte SHA-1 hash of each piece,
        // and the length gives you the total number of bytes in the buffer.
            const nPieces = calculateTotalPieces(); // torrent.info.pieces.length / 20;
            const arr = new Array(nPieces).fill(null);
            return arr.map((_, i) =>
                new Array(torrentParser.blocksPerPiece(torrent, i)).fill(false)
            );
        }

        this._requested = buildPiecesArray();
        this._received = buildPiecesArray();
        this.totalBlocks = this._requested.map((piece) => {
            return piece.reduce((count, _) => count + 1)
        })
        .reduce((acc, cur) => acc + cur);
        this.totalReceivedBlocks = 0;
        
    }
    
    addRequested(pieceBlock) {
        const blockIndex = pieceBlock.begin / torrentParser.BLOCK_LENGTH;
        this._requested[pieceBlock.begin][blockIndex] = true;
    }

    addReceived(pieceBlock) {
        const blockIndex = pieceBlock.begin / torrentParser.BLOCK_LENGTH;
        this._received[pieceBlock.index][blockIndex] = true;
    }

    needed(pieceBlock) {
        // if every block has been requested once
        if (this._requested.every((block) => block.every((i) => i))) {
            // copy the received
            this._requested = this._received.map((block) => block.slice())
        }
        const blockIndex = pieceBlock.begin / torrentParser.BLOCK_LENGTH;
        return !this,this._requested.every(block => block.every((i) => i))
    }

    isDone() {
        return this._received.every(block => block.every((i) => i))
    }
}
