package entryserver

import (
	"entry/logdebug"
	"entry/message"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"io"
	"sync"
)

func (server *Server) procCopyMsg(sessionWriter io.WriteCloser, msgContent []byte) {
	logdebug.Println(logdebug.LevelDebug, "--------COPY文件请求-------输入管道==", sessionWriter, string(msgContent))

	if len(msgContent) > 0 {
		sessionWriter.Write(msgContent)
	}

	return
}

func (server *Server) newCopyExec(dockerOpts docker.CreateExecOptions, clientID int, msgUnmarshaller Unmarshaler, msgMarshaller Marshaler, writeLock *sync.Mutex) {
	logdebug.Println(logdebug.LevelDebug, "创建copy消息用于展示dir的docker exec bash!")
	client := getClientInfo(clientID)

	var exec *docker.Exec
	var err error

	if exec, err = server.dockerClient.CreateExec(dockerOpts); err != nil {
		server.sendCloseMessage(client.ws, []byte("create docker client failed"), msgMarshaller, writeLock)

		logdebug.Println(logdebug.LevelError, "Exec docker error: %s", err.Error())

		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	//清理管道资源
	defer func() {
		client.copyStdoutPipeWriter.Close()
		client.copyStderrPipeWriter.Close()
		client.copyStdinPipeReader.Close()
		wg.Wait()
	}()

	//docker管道输出
	go server.handleResponse(client.ws, client.copyStdoutPipeReader, wg, message.ResponseMessage_COPY, msgMarshaller, writeLock)

	err = server.dockerClient.StartExec(exec.ID, docker.StartExecOptions{
		Detach:       false,
		OutputStream: client.copyStdoutPipeWriter,
		ErrorStream:  client.copyStderrPipeWriter,
		InputStream:  client.copyStdinPipeReader,
		RawTerminal:  false,
	})

	if err != nil {
		errMsg := fmt.Sprintf(errMsgTemplate, "连接COPY用bash失败.")
		server.sendCloseMessage(client.ws, []byte(errMsg), msgMarshaller, writeLock)

		return
	}
	server.sendCloseMessage(client.ws, []byte(byebyeMsg), msgMarshaller, writeLock)

	return
}
