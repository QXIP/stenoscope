const sstable = require('./index.js');

var args = process.argv.slice(2);
var datapath = args[0] || '/var/lib/stenographer/thread0/index';
var fromtime = parseInt(args[1]) || parseInt(new Date().getTime()/1000) - 60;
var totime =   parseInt(args[2]) || parseInt(new Date().getTime()/1000);

console.log(
  sstable.sstj(datapath, fromtime, totime )
);
