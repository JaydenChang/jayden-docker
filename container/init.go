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

// container/init.go
func NewProcess(tty bool, containerCmd string) *exec.Cmd {

	// create a new command which run itself
	// the first arguments is `init` which is the below exported function
	// so, the <cmd> will be interpret as "docker init <containerCmd>"
	args := []string{"init", containerCmd}
	cmd := exec.Command("/proc/self/exe", args...)

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
	}

	return cmd
}

// already in container
// initiate the container
func InitProcess() error {

	// read command from pipe, will plug if write side is not ready
	containerCmd := readCommand()
	if len(containerCmd) == 0 {
		return fmt.Errorf("Init process fails, containerCmd is nil")
	}
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV

	// mount proc filesystem
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	// look for the path of container command
	// so we don't need to type "/bin/ls", but "ls"
	path, err := exec.LookPath(containerCmd[0])
	if err != nil {
		logrus.Errorf("initProcess look path fails: %v", err)
		return err
	}

	// log path info
	// if you type "ls", it will be "/bin/ls"
	logrus.Infof("Find path: %v", path)
	if err := syscall.Exec(path, containerCmd, os.Environ()); err != nil {
		logrus.Errorf(err.Error())
	}

	return nil
}

func readCommand() []string {
	// 3 is the index of readPipe
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		logrus.Errorf("read pipe fails: %v", err)
		return nil
	}
	return strings.Split(string(msg), " ")
}
