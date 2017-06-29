package entryserver

// 将普通连接升级为websocket 向前端提供服务
import (
	//"fmt"
	"entry/config"
	"entry/logdebug"
	"github.com/gorilla/websocket"
	"log"

	"entry/dockerc"

	"encoding/json"
	//"entry/communicate"
	"entry/message"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/golang/protobuf/proto"
	"io"
	//"io/ioutil"
	"net"
	"net/http"
	//"os"
	"sync"
	"time"
	"unicode/utf8"
)

// Server webSocketServer and dockerClient
type Server struct {
	dockerClient  *docker.Client
	webSocketPort string
}

const (
	readBufferSize         = 1024
	writeBufferSize        = 10240           //The write buffer size should be large
	aliveDecectionInterval = time.Second * 1 //一分钟后超时
	byebyeMsg              = "\033[32m>>> You quit the container safely.\033[0m"
	errMsgTemplate         = "\033[31m>>> %s\033[0m"
)

//type CoreInfo map[string]AppInfo
//type ViaMethod int

// Marshaler 编码函数指针
type Marshaler func(interface{}) ([]byte, error)

// Unmarshaler 解码函数指针
type Unmarshaler func([]byte, interface{}) error

var upgrader = websocket.Upgrader{
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (server *Server) sendCloseMessage(ws *websocket.Conn, content []byte, msgMarshaller Marshaler, writeLock *sync.Mutex) {
	closeMsg := &message.ResponseMessage{
		MsgType: message.ResponseMessage_CLOSE,
		Content: content,
	}
	if closeData, err := msgMarshaller(closeMsg); err != nil {
		//log.Errorf("Marshal close message failed: %s", err.Error())
	} else {
		writeLock.Lock()
		ws.WriteMessage(websocket.BinaryMessage, closeData)
		writeLock.Unlock()
	}
}

// 将webSocket客户端模拟的标准输入 传递给docker
func (server *Server) transWebMsgToDokcer(ws *websocket.Conn,
	sessionWriter io.WriteCloser,
	copyStdinPipeWriter io.WriteCloser,
	execID string,
	msgUnmarshaller Unmarshaler,
	clientID int) (err error) {
	wsMsg := []byte{}
	inMsg := message.RequestMessage{}

	_, wsMsg, err = ws.ReadMessage()
	if err != nil {
		return
	}

	logdebug.Println(logdebug.LevelDebug, "---收到webclient的信息 解析前---", string(wsMsg))

	//解析失败 借用错误码 返回
	unmarshalErr := msgUnmarshaller(wsMsg, &inMsg)
	if unmarshalErr != nil {
		err = unmarshalErr
		logdebug.Println(logdebug.LevelError, "Unmarshall request error: %s", unmarshalErr.Error())

		return
	}

	logdebug.Println(logdebug.LevelDebug, "---收到webclient的信息 解析之后--- ", inMsg)
	switch inMsg.MsgType {
	case message.RequestMessage_PLAIN:
		if len(inMsg.Content) > 0 {
			_, err = sessionWriter.Write(inMsg.Content)
		}
	case message.RequestMessage_WINCH:
		server.procTTYMsg(execID, inMsg.Content)
	case message.RequestMessage_COPY:
		server.procCopyMsg(copyStdinPipeWriter, inMsg.Content)
	case message.RequestMessage_KEEPALIVE:
		server.procKeepAliveMsg(inMsg.Content, clientID)
	}

	return
}

// handleRequest处理标准输入请求
func (server *Server) handleRequest(ws *websocket.Conn,
	sessionWriter io.WriteCloser,
	copyStdinPipeWriter io.WriteCloser,
	wg *sync.WaitGroup,
	execID string,
	msgUnmarshaller Unmarshaler,
	clientID int) {
	var err error

	time.Sleep(time.Second)

	// keep on reading msg from web socket client
	for err == nil {
		err = server.transWebMsgToDokcer(ws, sessionWriter, copyStdinPipeWriter, execID, msgUnmarshaller, clientID)
	}

	logdebug.Println(logdebug.LevelError, "HandleRequest ended: %s", err.Error())

	sessionWriter.Close()

	wg.Done()

	return
}

func getValidUT8Length(data []byte) int {
	validLen := 0
	for i := len(data) - 1; i >= 0; i-- {
		if utf8.RuneStart(data[i]) {
			validLen = i
			if utf8.Valid(data[i:]) {
				validLen = len(data)
			}
			break
		}
	}
	return validLen
}

// 检查读取结果
// 读取出错 或 读取其他异常 退出
func checkReadResult(size int, err error) bool {
	if err == nil {
		return true
	}

	if err == io.EOF && size > 0 {
		return true
	}

	//由此出口 err必然不是nil
	return false
}

func transDockerMsgToWeb(ws *websocket.Conn, sessionReader io.ReadCloser, respType message.ResponseMessage_ResponseType, msgMarshaller Marshaler, writeLock *sync.Mutex) (err error) {
	var size int
	buf := make([]byte, writeBufferSize)

	cursor := 0

	size, err = sessionReader.Read(buf[cursor:])

	isReadSuccess := checkReadResult(size, err)
	if isReadSuccess != true {
		return
	}

	validLen := getValidUT8Length(buf[:cursor+size])
	if validLen == 0 {
		// 保留原版逻辑 可能是为了for循环下次处理？

		logdebug.Println(logdebug.LevelError, "No valid UTF8 sequence prefix")

		return
	}

	outMsg := &message.ResponseMessage{
		MsgType: respType,
		Content: buf[:validLen],
	}

	//直接填数字 前端能解析出来
	logdebug.Println(logdebug.LevelDebug, "响应管道=", sessionReader, "回显给web respType ", string(outMsg.Content), outMsg.MsgType)

	//借用错误码
	data, marshalErr := msgMarshaller(outMsg)
	if marshalErr != nil {
		err = marshalErr
		logdebug.Println(logdebug.LevelError, "Marshal response error: %s", marshalErr.Error())

		return
	}

	writeLock.Lock()
	//从管道读取的消息通过websocket返回给前端
	err = ws.WriteMessage(websocket.BinaryMessage, data)

	writeLock.Unlock()

	cursor = size - validLen

	for i := 0; i < cursor; i++ {
		buf[i] = buf[cursor+i]
	}

	return
}

// 将docker的标准输出返回给web客户端
func (server *Server) handleResponse(ws *websocket.Conn, sessionReader io.ReadCloser, wg *sync.WaitGroup, respType message.ResponseMessage_ResponseType, msgMarshaller Marshaler, writeLock *sync.Mutex) {
	var err error

	for err == nil {
		err = transDockerMsgToWeb(ws, sessionReader, respType, msgMarshaller, writeLock)
	}

	logdebug.Println(logdebug.LevelError, "HandleResponse ended: %s", err.Error())

	sessionReader.Close()

	wg.Done()

	return
}

// 解析出containerID字段
func getContainerIDFromWebClient(r *http.Request) (containerID string) {
	containerID = r.URL.Query().Get("containerID")

	if containerID == "" {
		logdebug.Println(logdebug.LevelDebug, "没有获取到containerID, 使用假数据")

		containerID = "50f55e114c56"
	}

	return
}

func getDockerServerURL(r *http.Request) (dockerServerURL string) {
	dockerServerURL = r.URL.Query().Get("dockerServerURL")

	if dockerServerURL == "" {
		logdebug.Println(logdebug.LevelDebug, "没有获取到dockerServerURL, 使用假数据")

		dockerServerURL = config.GetDockerAPIServerURL()
	}

	return
}

func (server *Server) createDockerClient(r *http.Request) (dockerOpts docker.CreateExecOptions) {
	dockerServerURL := getDockerServerURL(r)

	client, err := docker.NewClient(dockerServerURL)
	if err != nil {
		logdebug.Println(logdebug.LevelError, "连接docker服务器失败")

		return
	}

	server.dockerClient = client
	dockerOpts = dockerc.GetDockerOpts()
	dockerOpts.Container = getContainerIDFromWebClient(r)

	return
}

func (server *Server) enter(w http.ResponseWriter, r *http.Request) {
	var (
		err error
		ws  *websocket.Conn
	)

	//升级出webSocket
	ws, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		logdebug.Println(logdebug.LevelError, "Upgrade websocket protocol error: %s", err.Error())

		return
	}

	if ws != nil {
		defer ws.Close()
	}

	dockerOpts := server.createDockerClient(r)

	var exec *docker.Exec

	// 获取处理编码解码的函数
	msgMarshaller, msgUnmarshaller := getMarshalers(r)
	writeLock := &sync.Mutex{}

	if exec, err = server.dockerClient.CreateExec(dockerOpts); err != nil {
		server.sendCloseMessage(ws, []byte("create docker client failed"), msgMarshaller, writeLock)

		logdebug.Println(logdebug.LevelError, "Exec docker error: %s", err.Error())

		return
	}

	logdebug.Println(logdebug.LevelDebug, "exec success ", exec.ID)

	client := saveClientInfo(ws)

	wg := &sync.WaitGroup{}
	wg.Add(3)

	defer func() {
		logdebug.Println(logdebug.LevelDebug, "----清理管道资源----")
		client.stdoutPipeWriter.Close()
		client.stderrPipeWriter.Close()
		client.stdinPipeReader.Close()
		wg.Wait()
	}()

	//创建用于显示docker内部文件目录结构的bash
	go server.newCopyExec(dockerOpts, client.clientID, msgUnmarshaller, msgMarshaller, writeLock)

	//注册2个输入管道 一个写入XTERM的docker标准输入 另一个写入COPY的docker标准输入
	go server.handleRequest(ws, client.stdinPipeWriter, client.copyStdinPipeWriter, wg, exec.ID, msgUnmarshaller, client.clientID)
	go server.handleResponse(ws, client.stdoutPipeReader, wg, message.ResponseMessage_STDOUT, msgMarshaller, writeLock)
	go server.handleResponse(ws, client.stderrPipeReader, wg, message.ResponseMessage_STDERR, msgMarshaller, writeLock)

	//管道一端交给docker 另一端放在协程中 随时读取数据转发给webSocket客户端
	logdebug.Println(logdebug.LevelDebug, "XTERM exec start------输入输入管道=", client.stdinPipeWriter, client.stdoutPipeReader)
	//启动定时器
	err = server.dockerClient.StartExec(exec.ID, docker.StartExecOptions{
		Detach:       false,
		OutputStream: client.stdoutPipeWriter,
		ErrorStream:  client.stderrPipeWriter,
		InputStream:  client.stdinPipeReader,
		RawTerminal:  false,
	})

	// docker startExec 理论上会阻塞住
	logdebug.Println(logdebug.LevelDebug, "exec start结果 =", err)
	if err != nil {
		errMsg := fmt.Sprintf(errMsgTemplate, "Can't enter your container, try again.")
		server.sendCloseMessage(ws, []byte(errMsg), msgMarshaller, writeLock)

		return
	}

	server.sendCloseMessage(ws, []byte(byebyeMsg), msgMarshaller, writeLock)

	return
}

func (server *Server) attach(w http.ResponseWriter, r *http.Request) {

}

// StartServer 开启entry服务
func (server *Server) StartServer() {
	server.webSocketPort = config.GetServerPort()
	ClientsInfoInit()

	http.HandleFunc("/enter", server.enter)
	http.HandleFunc("/attach", server.attach)

	log.Fatal(http.ListenAndServe(net.JoinHostPort("", server.webSocketPort), nil))

	return
}

func getMarshalers(r *http.Request) (Marshaler, Unmarshaler) {

	//logdebug.Println(logdebug.LevelDebug, "---收到webclient的信息 选择解析函数---", *r)
	if r.URL.Query().Get("method") == "web" {

		logdebug.Println(logdebug.LevelDebug, "---收到webclient的信息 选择解析函数json---", r)

		return json.Marshal, json.Unmarshal
	}

	return protoMarshalFunc, protoUnmarshalFunc
}

// Adapters
func protoMarshalFunc(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func protoUnmarshalFunc(data []byte, v interface{}) error {
	return proto.Unmarshal(data, v.(proto.Message))
}
