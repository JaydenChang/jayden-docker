package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"simple-docker/cgroups"
	"simple-docker/cgroups/subsystem"
	"simple-docker/container"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func Run(tty bool, cmdArray []string, res *subsystem.ResourceConfig, volume, containerName, imageName string, envSlice []string) {
	containerID := randStringBytes(10)
	if containerName == "" {
		containerName = containerID
	}
	// this is "docker init <cmdArray>"
	initProcess, writePipe := container.NewParentProcess(tty, volume, containerName, imageName, envSlice)
	if initProcess == nil {
		logrus.Errorf("new parent process error")
		return
	}

	// start the init process
	if err := initProcess.Start(); err != nil {
		logrus.Error(err)
	}
	// container info
	containerName, err := recordContainerInfo(initProcess.Process.Pid, cmdArray, containerName, volume)
	if err != nil {
		logrus.Errorf("record container info error: %v", err)
		return
	}

	// create container manager to control resource config on all hierarchies
	cm := cgroups.NewCgroupManager("simple-docker-container")
	defer cm.Remove()
	cm.Set(res)
	cm.AddProcess(initProcess.Process.Pid)

	// send command to write side
	// will close the plug
	sendInitCommand(cmdArray, writePipe)

	if tty {
		initProcess.Wait()
		deleteContainerInfo(containerName)
		container.DeleteWorkSpace(volume, containerName)
	}
	os.Exit(0)
}

func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func deleteContainerInfo(containerID string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerID)
	if err := os.RemoveAll(dirURL); err != nil {
		logrus.Errorf("remove dir %s error %v", dirURL, err)
	}
}

func recordContainerInfo(containerPID int, commandArray []string, containerName, volume string) (string, error) {
	// create an ID that length is 10
	id := randStringBytes(10)
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")
	// if containerName is nil, make containerID as name
	if containerName == "" {
		containerName = id
	}
	containerInfo := &container.ContainerInfo{
		Id:          id,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      container.RUNNING,
		Name:        containerName,
		Volume:      volume,
	}
	// trun containerInfo info string
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		logrus.Errorf("record container info error: %v", err)
		return "", err
	}
	jsonStr := string(jsonBytes)

	// container path
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.MkdirAll(dirURL, 0622); err != nil {
		logrus.Errorf("mkdir error %s error: %v", dirURL, err)
		return "", err
	}
	fileName := dirURL + "/" + container.ConfigName
	// create config.json
	file, err := os.Create(fileName)
	if err != nil {
		logrus.Errorf("create %s error %v", fileName, err)
		return "", err
	}
	defer file.Close()
	// write jsonify data to file
	if _, err := file.WriteString(jsonStr); err != nil {
		logrus.Errorf("write %s error %v", fileName, err)
		return "", err
	}
	return containerName, nil
}

func sendInitCommand(cmdArray []string, writePipe *os.File) {
	cmdString := strings.Join(cmdArray, " ")
	logrus.Infof("whole init command is: %v", cmdString)
	writePipe.WriteString(cmdString)
	writePipe.Close()
}
