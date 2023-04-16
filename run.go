package main

import (
	"os"
	"simple-docker/cgroups"
	"simple-docker/cgroups/subsystem"
	"simple-docker/container"
	"strings"

	"github.com/sirupsen/logrus"
)

func Run(cmdArray []string, tty bool, res *subsystem.ResourceConfig, volume string) {
	parent, writePipe := container.NewParentProcess(tty, volume)
	if parent == nil {
		logrus.Errorf("failed to new parent process")
		return
	}
	if err := parent.Start(); err != nil {
		logrus.Errorf("parent start failed: %v", err)
		return
	}
	cgroupManager := cgroups.NewCGroupManager("go-docker")
	defer cgroupManager.Destroy()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)
	sendInitCommand(cmdArray, writePipe)
	parent.Wait()
}

func sendInitCommand(cmdArray []string, writePipe *os.File) {
	command := strings.Join(cmdArray, " ")
	logrus.Infof("command all is %s", command)
	_, _ = writePipe.WriteString(command)
	_ = writePipe.Close()
}
