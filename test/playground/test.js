// const buffer = new Uint8Array([201, 11]);
// // convert all to hex
// console.log(buffer[0]*256 + buffer[1]);

// // concat them
// // convert them to int

// import ProgressBar from "progress";
// var bar = new ProgressBar('Downloading [:bar] :rate/bps :percent :etas', {total: 50});
// var timer = setInterval(function () {
// bar.tick(); 
// if (bar.complete) {
//     console.log('\ncomplete\n')
//     clearInterval(timer);
// }
// }, 100);

const EventEmitter = require('node:events');
const eventEmitter = new EventEmitter();

eventEmitter.on('start', number => {
  console.log(`started ${number}`);
});

eventEmitter.emit('start', 23);