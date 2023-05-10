package utils

import (
	"fmt"
	"os/exec"
	"simple-docker/container"

	"github.com/sirupsen/logrus"
)

func commitContainer(containerName, imageName string) {
	mntURL := fmt.Sprintf(container.MntURL, containerName)
	mntURL += "/"
	imageTar := container.RootURL + "/" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		logrus.Errorf("tar folder %s error %v", mntURL, err)
	}
}
