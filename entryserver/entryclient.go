package entryserver

//在于web联调前 模拟webSocker的客户端
import (
	"entry/config"
	"entry/logdebug"
	"github.com/gorilla/websocket"
	"io"
	"sync"
	"time"
)

type clientInfo struct {
	ws                   *websocket.Conn
	stdinPipeReader      io.ReadCloser
	stdinPipeWriter      io.WriteCloser
	stdoutPipeReader     io.ReadCloser
	stdoutPipeWriter     io.WriteCloser
	stderrPipeReader     io.ReadCloser
	stderrPipeWriter     io.WriteCloser
	copyStdinPipeReader  io.ReadCloser
	copyStdinPipeWriter  io.WriteCloser
	copyStdoutPipeReader io.ReadCloser
	copyStdoutPipeWriter io.WriteCloser
	copyStderrPipeReader io.ReadCloser
	copyStderrPipeWriter io.WriteCloser
	timer                *time.Timer
	clientID             int
}

//Clients 客户端信息结构
type Clients struct {
	allClientsInfo map[int]clientInfo
	mutexLock      *sync.Mutex
	clientIDsList  []bool
}

//ClientsInfo 客户端信息
var ClientsInfo Clients

//ClientsInfoInit 客户端信息初始化
func ClientsInfoInit() {
	ClientsInfo.allClientsInfo = make(map[int]clientInfo, 0)
	ClientsInfo.mutexLock = new(sync.Mutex)
	ClientsInfo.clientIDsList = make([]bool, 0)

	return
}

//GetClientInfo 根据ClientID 获取client信息
func getClientInfo(clientID int) clientInfo {
	client := ClientsInfo.allClientsInfo[clientID]

	return client
}

//AddClientInfo 添加一条客户端信息节点
func addClientInfo(client clientInfo) int {
	ClientsInfo.mutexLock.Lock()

	defer ClientsInfo.mutexLock.Unlock()

	clientID := allocClientID()

	logdebug.Println(logdebug.LevelDebug, "申请clientID=", clientID)

	client.timer = time.AfterFunc(config.GetHeartCycle()*time.Second, func() {
		ClientsInfo.mutexLock.Lock()

		defer ClientsInfo.mutexLock.Unlock()

		delete(ClientsInfo.allClientsInfo, clientID)

		freeClientID(clientID)

		logdebug.Println(logdebug.LevelDebug, "------心跳超时 删除客户端信息----clientID !", clientID)

		client.ws.Close()
		//
		//web.stdoutPipeWriter.Close()
		//web.stderrPipeWriter.Close()
		//web.stdinPipeReader.Close()

		return
	})

	client.clientID = clientID

	ClientsInfo.allClientsInfo[clientID] = client

	return clientID
}

//UpdateClientInfo 更新客户端信息
func updateClientInfo(clientID int) {
	ClientsInfo.mutexLock.Lock()

	defer ClientsInfo.mutexLock.Unlock()

	if _, ok := ClientsInfo.allClientsInfo[clientID]; !ok {
		logdebug.Println(logdebug.LevelError, "不存在的clientID=", clientID)

		return
	}

	logdebug.Println(logdebug.LevelDebug, "刷新clientID=", clientID)

	web := ClientsInfo.allClientsInfo[clientID]
	web.timer.Stop()

	web.timer = time.AfterFunc(config.GetHeartCycle()*time.Second, func() {
		ClientsInfo.mutexLock.Lock()

		defer ClientsInfo.mutexLock.Unlock()

		delete(ClientsInfo.allClientsInfo, clientID)

		freeClientID(clientID)

		logdebug.Println(logdebug.LevelDebug, "------心跳超时 删除客户端信息----clientID !", clientID)

		web.ws.Close()

		//web.stdoutPipeWriter.Close()
		//web.stderrPipeWriter.Close()
		//web.stdinPipeReader.Close()

		return
	})

	ClientsInfo.allClientsInfo[clientID] = web

	return
}

func allocClientID() int {
	clientID := 0

	//排查已经存在位中 有没有空闲
	for clientID = 0; clientID < len(ClientsInfo.clientIDsList); clientID++ {
		if ClientsInfo.clientIDsList[clientID] == false {
			ClientsInfo.clientIDsList[clientID] = true

			clientID++

			return clientID
		}
	}

	//所有已经存在的位 都被占用 追加新的位
	ClientsInfo.clientIDsList = append(ClientsInfo.clientIDsList, true)

	clientID++

	return clientID
}

func freeClientID(clientID int) {
	if clientID > len(ClientsInfo.clientIDsList) {
		logdebug.Println(logdebug.LevelError, "释放失败!不合法的clientID=", clientID)

		return
	}

	//恢复为空闲状态
	ClientsInfo.clientIDsList[clientID-1] = false

	return
}

func saveClientInfo(ws *websocket.Conn) (client clientInfo) {
	stdinPipeReader, stdinPipeWriter := io.Pipe()
	stdoutPipeReader, stdoutPipeWriter := io.Pipe()
	stderrPipeReader, stderrPipeWriter := io.Pipe()

	copyStdinPipeReader, copyStdinPipeWriter := io.Pipe()
	copyStdoutPipeReader, copyStdoutPipeWriter := io.Pipe()
	copyStderrPipeReader, copyStderrPipeWriter := io.Pipe()

	client = clientInfo{
		ws:                   ws,
		stdinPipeReader:      stdinPipeReader,
		stdinPipeWriter:      stdinPipeWriter,
		stdoutPipeReader:     stdoutPipeReader,
		stdoutPipeWriter:     stdoutPipeWriter,
		stderrPipeReader:     stderrPipeReader,
		stderrPipeWriter:     stderrPipeWriter,
		copyStdinPipeReader:  copyStdinPipeReader,
		copyStdinPipeWriter:  copyStdinPipeWriter,
		copyStdoutPipeReader: copyStdoutPipeReader,
		copyStdoutPipeWriter: copyStdoutPipeWriter,
		copyStderrPipeReader: copyStderrPipeReader,
		copyStderrPipeWriter: copyStderrPipeWriter,
	}

	clientID := addClientInfo(client)

	client.clientID = clientID

	return
}
