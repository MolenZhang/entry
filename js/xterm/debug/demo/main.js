var term,
    protocol,
    socketURL,
    socket,
    pid,
    charWidth,
    charHeight;

var terminalContainer = document.getElementById('terminal-container'),
    optionElements = {
      cursorBlink: document.querySelector('#option-cursor-blink')
    },
    colsElement = document.getElementById('cols'),
    rowsElement = document.getElementById('rows');

function setTerminalSize () {
  var cols = parseInt(colsElement.value),
      rows = parseInt(rowsElement.value),
      width = (cols * charWidth).toString() + 'px',
      height = (rows * charHeight).toString() + 'px';

  terminalContainer.style.width = width;
  terminalContainer.style.height = height;
  term.resize(cols, rows);
}

colsElement.addEventListener('change', setTerminalSize);
rowsElement.addEventListener('change', setTerminalSize);

optionElements.cursorBlink.addEventListener('change', createTerminal);

createTerminal();

function createTerminal() {
  // Clean terminal
  while (terminalContainer.children.length) {
    terminalContainer.removeChild(terminalContainer.children[0]);
  }
  term = new Terminal({
    cursorBlink: optionElements.cursorBlink.checked
  });
//  term.on('resize', function (size) {
//    if (!pid) {
//      return;
//    }
//    var cols = size.cols,
//        rows = size.rows,
//        url = '/terminals/' + pid + '/size?cols=' + cols + '&rows=' + rows;
//
//    fetch(url, {method: 'POST'});
//  });
  protocol = (location.protocol === 'https:') ? 'wss://' : 'ws://';
  socketURL = protocol + location.hostname + ((location.port) ? (':' + location.port) : '') + '/terminals/';

  term.open(terminalContainer);
  term.fit();

  var initialGeometry = term.proposeGeometry(),
      cols = initialGeometry.cols,
      rows = initialGeometry.rows;

  colsElement.value = cols;
  rowsElement.value = rows;

  //fetch('/terminals?cols=' + cols + '&rows=' + rows, {method: 'POST'}).then(function (res) {

    // charWidth = Math.ceil(term.element.offsetWidth / cols);
    // charHeight = Math.ceil(term.element.offsetHeight / rows);

    // res.text().then(function (pid) {
    //   window.pid = pid;
    //   socketURL += pid;
      //var soc = "ws://192.168.252.138:8011/enter?method=web&containerID="+getQueryString('containerID')+"&dockerServerURL="+getQueryString('dockerServerURL');
      var soc = "ws://192.168.252.138:8011/enter?method=web&containerID=12b6dafc1bbc&dockerServerURL=http://192.168.0.81:28015"
      socket = new WebSocket(soc);
      //socket = new WebSocket(socketURL);
      socket.onopen = runRealTerminal;
      socket.onclose = runFakeTerminal;
      socket.onerror = runFakeTerminal;
  //   });
  // });
}


function stringToByte(str) {
  var bytes = new Array();
  var len, c;
  len = str.length;
  for(var i = 0; i < len; i++) {
    c = str.charCodeAt(i);
    if(c >= 0x010000 && c <= 0x10FFFF) {
      bytes.push(((c >> 18) & 0x07) | 0xF0);
      bytes.push(((c >> 12) & 0x3F) | 0x80);
      bytes.push(((c >> 6) & 0x3F) | 0x80);
      bytes.push((c & 0x3F) | 0x80);
    } else if(c >= 0x000800 && c <= 0x00FFFF) {
      bytes.push(((c >> 12) & 0x0F) | 0xE0);
      bytes.push(((c >> 6) & 0x3F) | 0x80);
      bytes.push((c & 0x3F) | 0x80);
    } else if(c >= 0x000080 && c <= 0x0007FF) {
      bytes.push(((c >> 6) & 0x1F) | 0xC0);
      bytes.push((c & 0x3F) | 0x80);
    } else {
      bytes.push(c & 0xFF);
    }
  }
  return bytes;
}

function runRealTerminal() {
  term.attach(socket);
  term._initialized = true;

  // alert("-----send-----");
  //
  // var stdinMsg = {
  //   "MsgType" : 0,
  //   "Content" : stringToByte("ls\n"),
  // };
  //
  // socket.send(JSON.stringify(stdinMsg));
  // alert("-----aftger send-----");
}

function runFakeTerminal() {
  if (term._initialized) {
    return;
  }

  term._initialized = true;

  var shellprompt = '$ ';

  term.prompt = function () {
    term.write('\r\n' + shellprompt);
  };

  term.writeln('Welcome to xterm.js');
  term.writeln('This is a local terminal emulation, without a real terminal in the back-end.');
  term.writeln('Type some keys and commands to play around.');
  term.writeln('');
  term.prompt();

  term.on('key', function (key, ev) {
    var printable = (
      !ev.altKey && !ev.altGraphKey && !ev.ctrlKey && !ev.metaKey
    );

    if (ev.keyCode == 13) {
      term.prompt();
    } else if (ev.keyCode == 8) {
     // Do not delete the prompt
      if (term.x > 2) {
        term.write('\b \b');
      }
    } else if (printable) {
      term.write(key);
    }
  });

  term.on('paste', function (data, ev) {
    term.write(data);
  });
}
function getQueryString(name) {
  var reg = new RegExp("(^|&)" + name + "=([^&]*)(&|$)", "i");
  var r = window.location.search.substr(1).match(reg);
  if (r != null) return unescape(r[2]);
  return null;
}
