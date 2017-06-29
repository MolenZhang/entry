var socket;

// private method for UTF-8 decoding  
     function _utf8_decode (utftext) {  
        var string = "";  
        var i = 0;  
        var c = c1 = c2 = 0;  
        while ( i < utftext.length ) {  
            c = utftext.charCodeAt(i);  
            if (c < 128) {  
                string += String.fromCharCode(c);  
                i++;  
            } else if((c > 191) && (c < 224)) {  
                c2 = utftext.charCodeAt(i+1);  
                string += String.fromCharCode(((c & 31) << 6) | (c2 & 63));  
                i += 2;  
            } else {  
                c2 = utftext.charCodeAt(i+1);  
                c3 = utftext.charCodeAt(i+2);  
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

 function byteToString(arr) {  
        if(typeof arr === 'string') {  
            return arr;  
        }  
        var str = '',  
            _arr = arr;  
        for(var i = 0; i < _arr.length; i++) {  
            var one = _arr[i].toString(2),  
                v = one.match(/^1+?(?=0)/);  
            if(v && one.length == 8) {  
                var bytesLength = v[0].length;  
                var store = _arr[i].toString(2).slice(7 - bytesLength);  
                for(var st = 1; st < bytesLength; st++) {  
                    store += _arr[st + i].toString(2).slice(2);  
                }  
                str += String.fromCharCode(parseInt(store, 2));  
                i += bytesLength - 1;  
            } else {  
                str += String.fromCharCode(_arr[i]);  
            }  
        }  
        return str;  
    }  
 
$("#connect").click(function(event){
    //socket = new WebSocket("ws://192.168.0.76:28015/v1.22/containers/68600354b940/attach/ws?logs=0&stream=1&stdin=1&stdout=1&stderr=1");
	socket = new WebSocket("ws://192.168.252.138:8011/enter?method=web&containerID=9af31acfaf81&dockerServerURL=http://192.168.0.80:28015");
	//socket = new WebSocket("ws://192.168.252.138:8011/enter");
 
    socket.onopen = function(){
        alert("Socket has been opened");
    }
 
    socket.onmessage = function(msg){
		
		 //var blob = msg.data;
            //先把blob进行拆分，第一个字节是标识
         //var newblob = blob.slice(0,4);
		
		//alert(msg.data);

		var reader = new FileReader();
		
		reader.readAsBinaryString(msg.data);
		//reader.readAsBinaryString(newblob);
		
		reader.onloadend = function () {
			alert(reader.result);
			//alert(_utf8_decode(reader.result));
			var object = eval('('+reader.result+')');
			
			$("#tbody").append("<tr><td>"+object+"</td></tr>");
			$("#tbody").append("<tr><td>"+Base64(object.content)+"</td></tr>");
			$("#tbody").append("<tr><td>"+object.msgType+"</td></tr>");
			
			//alert(Base64(object.content));
			//alert(object.msgType);
			
	    }
		
    }
 
    socket.onclose = function() {
        alert("Socket has been closed");
    }
});

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
 
$("#send").click(function(event){
	alert("----send data----");
	
	var stdinMsg = {
		"MsgType" : 0,
		"Content" : stringToByte("ls\n"),
	};
	
	var ttyMsg = {
		"MsgType" : 1,
		"Content" : stringToByte("20 30"),
	};
	
	var copyMsg = {
		"MsgType" : 2,
		"Content" : stringToByte("ls -ap\n"),
	};
	
	var keepAliveMsg = {
		"MsgType" : 3,
		"Content" : stringToByte("js---keepalive"),
	};
	
	
    //socket.send(JSON.stringify(stdinMsg));
	socket.send(JSON.stringify(copyMsg));
	//alert("发送后");
});
 
$("#close").click(function(event){
	alert("----close----");
    socket.close();
})