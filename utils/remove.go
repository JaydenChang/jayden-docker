package utils

import (
	"fmt"
	"os"
	"simple-docker/container"

	"github.com/sirupsen/logrus"
)

func removeContainer(containerName string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		logrus.Errorf("get container %s info failed: %v", containerName, err)
		return
	}
	// only remove the stopped container
	if containerInfo.Status != container.STOP {
		logrus.Errorf("cannot remove running container %s", containerName)
		return
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	fmt.Println("################", dirURL)
	// remove all the info including sub dir
	if err := os.RemoveAll(dirURL); err != nil {
		logrus.Errorf("cannot remove dir %s error: %v", dirURL, err)
		return
	}
	container.DeleteWorkSpace(containerInfo.Volume, containerName)
}
