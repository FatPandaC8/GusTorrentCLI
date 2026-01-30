const numPieces = 3;

const pieces = new Map();
for (let i = 0; i < numPieces; i++) {
    pieces.set(`piece${i}`, "missing");
}

// if a piece just arrived => set to downloading

// while the pieces length being downloaded < the that whole piece length:
pieces.set(`piece1`, "downloading");

pieces.set(`piece2`, "downloaded");

console.log(pieces);