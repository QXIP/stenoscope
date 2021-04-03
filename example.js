const sstable = require('./index.js');
console.log(sstable);

var fromtime = parseInt(new Date().getTime()/1000) - 60;
var totime = parseInt(new Date().getTime()/1000);
console.log(
  fromtime, totime,
  sstable.sstj('/var/lib/stenographer/thread0/index', fromtime, totime )
);
