package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

type ContainerInfo struct {
	Pid         string `json:"pid"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"` // the command that init process execute
	CreatedTime string `json:"created_time"`
	Status      string `json:"status"`
}

var (
	RUNNING             string = "running"
	STOP                string = "stopped"
	Exit                string = "exited"
	DefaultInfoLocation string = "/var/run/simple-docker/%s"
	ConfigName          string = "config.json"
	ContainerLogFile    string = "container.log"
)

func NewParentProcess(tty bool, volume string, containerName string) (*exec.Cmd, *os.File) {

	readPipe, writePipe, err := os.Pipe()

	if err != nil {
		logrus.Errorf("New Pipe Error: %v", err)
		return nil, nil
	}
	// create a new command which run itself
	// the first arguments is `init` which is in the "container/init.go" file
	// so, the <cmd> will be interpret as "docker init <containerCmd>"
	cmd := exec.Command("/proc/self/exe", "init")

	// new namespaces, thanks to Linux
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
	}

	// this is what presudo terminal means
	// link the container's stdio to os
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// create container.log i
		dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
		if err := os.MkdirAll(dirURL, 0622); err != nil {
			logrus.Errorf("new parent process mkdir %s error: %v", dirURL, err)
		}
		stdLogFilePath := dirURL + "/" + ContainerLogFile
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			logrus.Errorf("new parent process create file %s error: %v", stdLogFilePath, err)
			return nil, nil
		}
		// put stdLogFile as stdout
		cmd.Stdout = stdLogFile
	}

	cmd.ExtraFiles = []*os.File{readPipe}
	rootURL := "/root/"
	mntURL := "/root/mnt/"
	NewWorkSpace(rootURL, mntURL, volume)
	cmd.Dir = mntURL
	return cmd, writePipe
}

func readCommand() []string {
	// 3 is the index of readPipe
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		logrus.Errorf("read pipe failed: %v", err)
		return nil
	}
	return strings.Split(string(msg), " ")
}
