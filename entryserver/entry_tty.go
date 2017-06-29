package entryserver

import (
	"entry/logdebug"
	"strconv"
	"strings"
)

func getWidthAndHeight(data []byte) (int, int) {
	sizeStr := string(data)
	sizeArr := strings.Split(sizeStr, " ")

	if len(sizeArr) != 2 {
		return -1, -1
	}
	var width, height int
	var err error

	if width, err = strconv.Atoi(sizeArr[0]); err != nil {
		return -1, -1
	}
	if height, err = strconv.Atoi(sizeArr[1]); err != nil {
		return -1, -1
	}

	return width, height
}

func (server *Server) procTTYMsg(execID string, msgContent []byte) {
	width, height := getWidthAndHeight(msgContent)

	if width >= 0 && height >= 0 {
		logdebug.Printf(logdebug.LevelDebug, "---调整终端---", width, height)
		server.dockerClient.ResizeExecTTY(execID, height, width)
	}

	return
}
