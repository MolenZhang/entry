var express = require('express');
var app = express();
var expressWs = require('express-ws')(app);
var os = require('os');
var pty = require('pty.js');

var terminals = {},
    logs = {};

app.use('/build', express.static(__dirname + '/../build'));

app.get('/', function(req, res){
  res.sendFile(__dirname + '/index.html');
});

app.get('/style.css', function(req, res){
  res.sendFile(__dirname + '/style.css');
});

app.get('/main.js', function(req, res){
  res.sendFile(__dirname + '/main.js');
});

app.get('/jquery-1.11.3.min.js', function(req, res){
  res.sendFile(__dirname + '/jquery-1.11.3.min.js');
});

app.post('/terminals', function (req, res) {
  var cols = parseInt(req.query.cols),
      rows = parseInt(req.query.rows),
      term = pty.spawn(process.platform === 'win32' ? 'cmd.exe' : 'bash', [], {
        name: 'xterm-color',
        cols: cols || 80,
        rows: rows || 24,
        cwd: process.env.PWD,
        env: process.env
      });

  console.log('Created terminal with PID: ' + term.pid);
  terminals[term.pid] = term;
  logs[term.pid] = '';
  term.on('data', function(data) {
    logs[term.pid] += data;
  });
  res.send(term.pid.toString());
  res.end();
});

app.post('/terminals/:pid/size', function (req, res) {
  var pid = parseInt(req.params.pid),
      cols = parseInt(req.query.cols),
      rows = parseInt(req.query.rows),
      term = terminals[pid];

  term.resize(cols, rows);
  console.log('Resized terminal ' + pid + ' to ' + cols + ' cols and ' + rows + ' rows.');
  res.end();
});

app.ws('/terminals/:pid', function (ws, req) {
  var term = terminals[parseInt(req.params.pid)];
  console.log('Connected to terminal ' + term.pid);
  ws.send(logs[term.pid]);

  term.on('data', function(data) {
   try {
      //通过webSocket发送数据
     // var entryMsg = {
     //   "MsgType" : 0,
     //   "Content" :  stringToByte("ls\n"),
     // };
     // ws.send(JSON.stringify(entryMsg));

      console.log('----转发给前端的term 用于屏幕显示----' + data + '----时间----' + new Date().getTime());
      ws.send(data);
    } catch (ex) {
      // The WebSocket is not open, ignore
    }
  });
  //将webSocket收到的消息发送给term 保留
  ws.on('message', function(msg) {
    console.log('----收到键盘输入----' + msg + '----时间----' + new Date().getTime());
    term.write(msg);//2.收到键盘的消息 再发给term
  });
  ws.on('close', function () {
    process.kill(term.pid);
    console.log('Closed terminal ' + term.pid);
    // Clean things up
    delete terminals[term.pid];
    delete logs[term.pid];
  });
});

var port = process.env.PORT || 3000,
    host = os.platform() === 'win32' ? '127.0.0.1' : '0.0.0.0';

console.log('App listening to http://' + host + ':' + port);
app.listen(port, host);
