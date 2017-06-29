package entryserver

import (
	"entry/logdebug"
)

func (server *Server) procKeepAliveMsg(msgContent []byte, clientID int) {
	logdebug.Println(logdebug.LevelDebug, "--------心跳 刷新定时器-------", string(msgContent))

	updateClientInfo(clientID)

	return
}
