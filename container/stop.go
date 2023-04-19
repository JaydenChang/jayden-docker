package container

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"simple-docker/common"
	"strconv"
	"syscall"

	"github.com/sirupsen/logrus"
)

func StopContainer(containerName string) {
	info, err := getContainerInfo(containerName)
	if err != nil {
		logrus.Errorf("get container info, err: %v", err)
		return
	}
	if info.Pid != "" {
		pid, _ := strconv.Atoi(info.Pid)
		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			logrus.Errorf("stop container, pid: %d, err: %v", pid, err)
			return
		}
		info.Status = common.Stop
		info.Pid = ""
		bs, _ := json.Marshal(info)
		fileName := path.Join(common.DefaultContainerInfoPath, containerName, common.ContainerInfoFileName)
		err := ioutil.WriteFile(fileName, bs, 0622)
		if err != nil {
			logrus.Errorf("write container config.json, err: %v", err)
		}
	}
}
