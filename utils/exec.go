package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"simple-docker/container"
	"strings"

	"github.com/sirupsen/logrus"
)

const ENV_EXEC_PID = "simple_docker_pid"
const ENV_EXEC_CMD = "simple_docker_cmd"

func getContainerPidByName(containerName string) (string, error) {
	// get the path that store container info
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + "/" + container.ConfigName
	// read files in target path
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var containerInfo container.ContainerInfo
	// unmarshal json to containerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}

func ExecContainer(containerName string, comArray []string) {
	// get the pid according the containerName
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		logrus.Errorf("exec container getContainerPidByName %s error %v", containerName, err)
		return
	}
	// divide command by blank space and combine as a string
	cmdStr := strings.Join(comArray, " ")
	logrus.Infof("container pid %s", pid)
	logrus.Infof("command %s", cmdStr)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = os.Setenv(ENV_EXEC_PID, pid)
	if err != nil {
		logrus.Errorf("set env exec pid %s error %v", pid, err)
	} 
	err = os.Setenv(ENV_EXEC_CMD, cmdStr)
	if err != nil {
		logrus.Errorf("set env exec command %s error %v", cmdStr, err)
	}	

	if err := cmd.Run(); err != nil {
		logrus.Errorf("exec container %s error %v", containerName, err)
	}
}
