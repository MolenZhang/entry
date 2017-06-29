package dockerc

//与docker API通讯 并将结果转发写入管道 由webserver读取
import (
	"entry/config"
	"github.com/fsouza/go-dockerclient"
)

func Init() {

}

func getContainerID() string {
	return "50f55e114c56"
}

func getExecCmd() []string {
	return []string{"/bin/sh"}
}

// GetDockerOpts 获取docker客户端订制参数
func GetDockerOpts() docker.CreateExecOptions {

	//containerID := getContainerID()

	execCmd := getExecCmd()

	opts := docker.CreateExecOptions{
		//Container:    containerID,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          execCmd,
	}

	return opts
}

// NewDockerClient docker客户端
func NewDockerClient() (*docker.Client, error) {
	dockerServerURL := config.GetDockerAPIServerURL()

	return docker.NewClient(dockerServerURL)

}
