/**
 * Implements the attach method, that attaches the terminal to a WebSocket stream.
 * @module xterm/addons/attach/attach
 * @license MIT
 */
function stringToByte(str) {
    var bytes = new Array();
    var len, c;
    len = str.length;
    for (var i = 0; i < len; i++) {
        c = str.charCodeAt(i);
        if (c >= 0x010000 && c <= 0x10FFFF) {
            bytes.push(((c >> 18) & 0x07) | 0xF0);
            bytes.push(((c >> 12) & 0x3F) | 0x80);
            bytes.push(((c >> 6) & 0x3F) | 0x80);
            bytes.push((c & 0x3F) | 0x80);
        }
        else if (c >= 0x000800 && c <= 0x00FFFF) {
            bytes.push(((c >> 12) & 0x0F) | 0xE0);
            bytes.push(((c >> 6) & 0x3F) | 0x80);
            bytes.push((c & 0x3F) | 0x80);
        }
        else if (c >= 0x000080 && c <= 0x0007FF) {
            bytes.push(((c >> 6) & 0x1F) | 0xC0);
            bytes.push((c & 0x3F) | 0x80);
        }
        else {
            bytes.push(c & 0xFF);
        }
    }
    return bytes;
}
// private method for UTF-8 decoding
function _utf8_decode(utftext) {
    var string = "";
    var i = 0;
    var c = c1 = c2 = 0;
    while (i < utftext.length) {
        c = utftext.charCodeAt(i);
        if (c < 128) {
            string += String.fromCharCode(c);
            i++;
        }
        else if ((c > 191) && (c < 224)) {
            c2 = utftext.charCodeAt(i + 1);
            string += String.fromCharCode(((c & 31) << 6) | (c2 & 63));
            i += 2;
        }
        else {
            c2 = utftext.charCodeAt(i + 1);
            c3 = utftext.charCodeAt(i + 2);
            string += String.fromCharCode(((c & 15) << 12) | ((c2 & 63) << 6) | (c3 & 63));
            i += 3;
        }
    }
    return string;
}
function Base64(input) {
    // private property
    _keyStr = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=";
    // public method for decoding
    var output = "";
    var chr1, chr2, chr3;
    var enc1, enc2, enc3, enc4;
    var i = 0;
    input = input.replace(/[^A-Za-z0-9\+\/\=]/g, "");
    while (i < input.length) {
        enc1 = _keyStr.indexOf(input.charAt(i++));
        enc2 = _keyStr.indexOf(input.charAt(i++));
        enc3 = _keyStr.indexOf(input.charAt(i++));
        enc4 = _keyStr.indexOf(input.charAt(i++));
        chr1 = (enc1 << 2) | (enc2 >> 4);
        chr2 = ((enc2 & 15) << 4) | (enc3 >> 2);
        chr3 = ((enc3 & 3) << 6) | enc4;
        output = output + String.fromCharCode(chr1);
        if (enc3 != 64) {
            output = output + String.fromCharCode(chr2);
        }
        if (enc4 != 64) {
            output = output + String.fromCharCode(chr3);
        }
    }
    output = _utf8_decode(output);
    return output;
}
(function (attach) {
    if (typeof exports === 'object' && typeof module === 'object') {
        /*
         * CommonJS environment
         */
        module.exports = attach(require('../../xterm'));
    }
    else if (typeof define == 'function') {
        /*
         * Require.js is available
         */
        define(['../../xterm'], attach);
    }
    else {
        /*
         * Plain browser environment
         */
        attach(window.Terminal);
    }
})(function (Xterm) {
    'use strict';
    var exports = {};
    /**
     * Attaches the given terminal to the given socket.
     *
     * @param {Xterm} term - The terminal to be attached to the given socket.
     * @param {WebSocket} socket - The socket to attach the current terminal.
     * @param {boolean} bidirectional - Whether the terminal should send data
     *                                  to the socket as well.
     * @param {boolean} buffered - Whether the rendering of incoming data
     *                             should happen instantly or at a maximum
     *                             frequency of 1 rendering per 10ms.
     */
    exports.attach = function (term, socket, bidirectional, buffered) {
        bidirectional = (typeof bidirectional == 'undefined') ? true : bidirectional;
        term.socket = socket;
        term._flushBuffer = function () {
            term.write(term._attachSocketBuffer);
            term._attachSocketBuffer = null;
            clearTimeout(term._attachSocketBufferTimer);
            term._attachSocketBufferTimer = null;
        };
        term._pushToBuffer = function (data) {
            if (term._attachSocketBuffer) {
                term._attachSocketBuffer += data;
            }
            else {
                term._attachSocketBuffer = data;
                setTimeout(term._flushBuffer, 10);
            }
        };
        //收到ws发送的消息
        term._getMessage = function (ev) {
            if (buffered) {
                term._pushToBuffer(ev.data);
                console.log('[src-attach.js]term push to buffer----ev.data' + ev.data + '----时间----' + new Date().getTime());
                //term.write("-----buffer-----helloworld");
            }
            else {
                console.log('[src-attach.js]-----显示在屏幕上----ev.data' + ev.data + '----时间----' + new Date().getTime());
                var reader = new FileReader();
                reader.readAsBinaryString(ev.data); //blob
                reader.onloadend = function () {
                    var object = eval('(' + reader.result + ')');
                    term.write(Base64(object.content)); //4.回显
                };
                //term.write(ev.data);//4.回显
            }
        };
        term._sendData = function (data) {
            var stdinMsg = {
                "MsgType": 0,
                "Content": stringToByte(data),
            };
            console.log('[src-attach.js]-----抓到键盘的输入-----' + stdinMsg);
            socket.send(JSON.stringify(stdinMsg)); //1.抓到键盘 发给ws
            //socket.send(data);//1.抓到键盘 发给ws
        };
        socket.addEventListener('message', term._getMessage);
        if (bidirectional) {
            term.on('data', term._sendData);
        }
        socket.addEventListener('close', term.detach.bind(term, socket));
        socket.addEventListener('error', term.detach.bind(term, socket));
    };
    /**
     * Detaches the given terminal from the given socket
     *
     * @param {Xterm} term - The terminal to be detached from the given socket.
     * @param {WebSocket} socket - The socket from which to detach the current
     *                             terminal.
     */
    exports.detach = function (term, socket) {
        term.off('data', term._sendData);
        socket = (typeof socket == 'undefined') ? term.socket : socket;
        if (socket) {
            socket.removeEventListener('message', term._getMessage);
        }
        delete term.socket;
    };
    /**
     * Attaches the current terminal to the given socket
     *
     * @param {WebSocket} socket - The socket to attach the current terminal.
     * @param {boolean} bidirectional - Whether the terminal should send data
     *                                  to the socket as well.
     * @param {boolean} buffered - Whether the rendering of incoming data
     *                             should happen instantly or at a maximum
     *                             frequency of 1 rendering per 10ms.
     */
    Xterm.prototype.attach = function (socket, bidirectional, buffered) {
        return exports.attach(this, socket, bidirectional, buffered);
    };
    /**
     * Detaches the current terminal from the given socket.
     *
     * @param {WebSocket} socket - The socket from which to detach the current
     *                             terminal.
     */
    Xterm.prototype.detach = function (socket) {
        return exports.detach(this, socket);
    };
    return exports;
});
//# sourceMappingURL=attach.js.map