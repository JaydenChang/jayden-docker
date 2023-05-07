package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"simple-docker/container"
	"strconv"
	"syscall"

	"github.com/sirupsen/logrus"
)

func stopContainer(containerName string) {
	// get pid by containerName
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		logrus.Errorf("get container pid by name %s error %v", containerName, err)
		return
	}
	// turn pid(string) to int
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		logrus.Errorf("convert pid from string to int error %v", err)
		return
	}
	// kill container main process
	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		logrus.Errorf("stop container %s error %v", containerName, err)
		return
	}
	// get info of the container
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		logrus.Errorf("get container info by name %s error %v", containerName, err)
		return
	}
	// process is killed, update process status
	containerInfo.Status = container.STOP
	containerInfo.Pid = " "
	// update info to json
	nweContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		logrus.Errorf("json marshal %s error %v", containerName, err)
		return
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + "/" + container.ConfigName
	// overwrite containerInfo
	if err := ioutil.WriteFile(configFilePath, nweContentBytes, 0622); err != nil {
		logrus.Errorf("write config file %s error %v", configFilePath, err)
	}
}

func getContainerInfoByName(containerName string) (*container.ContainerInfo, error) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + "/" + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		logrus.Errorf("read config file %s error %v", configFilePath, err)
		return nil, err
	}
	var containerInfo container.ContainerInfo
	// unmarshal json to container info
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		logrus.Errorf("unmarshal json to container info error %v", err)
		return nil, err
	}
	return &containerInfo, nil
}
