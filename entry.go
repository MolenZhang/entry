package main

import (
	//"fmt"
	"entry/config"
	//"entry/convert"
	//"entry/dockerc"
	//"entry/dockerc"
	//"entry/logdebug"
	"entry/entryserver"
	//"entry/message"
)

func init() {
	config.Init()
	//logdebug.Init()
	//convert.Init()
	//webserver.Init()
	//dockerc.Init()
	//message.Init()
}

func main() {
	entryServer := entryserver.Server{}

	entryServer.StartServer()

	return
}
