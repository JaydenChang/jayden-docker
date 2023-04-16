package container

import (
	"fmt"
	"os/exec"
	"path"
	"simple-docker/common"

	"github.com/sirupsen/logrus"
)

func CommitContainer(imageName, imagePath string) error {
	if imagePath == "" {
		imagePath = common.RootPath
	}
	imageTar := path.Join(imagePath, fmt.Sprintf("%s.tar", imageName))
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", common.MntPath, ".").CombinedOutput(); err != nil {
		logrus.Errorf("tar container imagem file name: %s, err: %v", imageTar, err)
		return err
	}
	return nil
}
