package config

import (
	"flag"
	"fmt"
	"strconv"
	"time"
)

// 启动参数文件

// 启动参数 后续做拓展
type entryStartCfg struct {
	serverPort      string
	dockerServerURL string
	logPrintLevel   string
	heartCycle      time.Duration
}

var entryCfg entryStartCfg

// 默认启动参数
const (
	defaultServerPort      = "8011"
	defaultDockerServerURL = "http://192.168.0.76:28015"
	defaultLogPrintLevel   = "info"
)

//DefaultHeartTimeout 默认心跳超时时间 5 * 3 = 15秒
const DefaultHeartTimeout = 15

// GetServerPort 获取webSocket监听服务端口
func GetServerPort() (serverPort string) {
	serverPort = entryCfg.serverPort

	return
}

// GetDockerAPIServerURL 获取dockerAPI URL
func GetDockerAPIServerURL() (dockerServerURL string) {
	dockerServerURL = entryCfg.dockerServerURL

	return
}

// GetLogPrintLevel 获取打印界别
func GetLogPrintLevel() (logPrintLevel string) {
	logPrintLevel = entryCfg.logPrintLevel

	return
}

//GetHeartCycle 获取心跳周期
func GetHeartCycle() time.Duration {
	//fmt.Println("get heart cycle=", entryCfg.heartCycle)

	return entryCfg.heartCycle
}

// Init 初始化启动参数模块
func Init() {
	heartCycle := ""

	flag.StringVar(&entryCfg.serverPort, "port", defaultServerPort, "web socket server listen port")
	flag.StringVar(&entryCfg.dockerServerURL, "dockerserver", defaultDockerServerURL, "docker API server")
	flag.StringVar(&entryCfg.logPrintLevel, "loglevel", defaultLogPrintLevel, "log print level")
	flag.StringVar(&heartCycle, "heart", "15", "heart cycle")

	flag.Parse()

	entryCfg.heartCycle = DefaultHeartTimeout

	intTypeHeartCycle, err := strconv.Atoi(heartCycle)
	if err != nil {
		fmt.Println("解析心跳间隔失败使用默认值")

		return
	}

	entryCfg.heartCycle = time.Duration(intTypeHeartCycle)

	return
}
