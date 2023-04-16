package container

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"simple-docker/common"
	"strings"

	"github.com/sirupsen/logrus"
)

func NewWorkSpace(rootPath string, mntPath string, volume string) error {
	err := createReadOnlyLayer(rootPath)
	if err != nil {
		logrus.Errorf("create read only layer, err: %v", err)
		return err
	}
	err = createWriteLayer(rootPath)
	if err != nil {
		logrus.Errorf("create write layer, err: %v", err)
		return err
	}
	err = createMountPoint(rootPath, mntPath)
	if err != nil {
		logrus.Errorf("create mount point, err: %v", err)
		return err
	}
	mountVolume(rootPath, mntPath, volume)
	return nil
}

func createReadOnlyLayer(rootPath string) error {
	busyBoxPath := path.Join(rootPath, common.BusyBox)
	_, err := os.Stat(busyBoxPath)
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(busyBoxPath, os.ModePerm)
		if err != nil {
			logrus.Errorf("mkdir busybox, err: %v", err)
			return err
		}
	}
	busyBoxTarPath := path.Join(rootPath, common.BusyBoxTar)
	if _, err := exec.Command("tar", "-xvf", busyBoxTarPath, "-C", busyBoxPath).CombinedOutput(); err != nil {
		logrus.Errorf("tar busybox, err: %v", err)
		return err
	}
	return nil
}

func createWriteLayer(rootPath string) error {
	writeLayerPath := path.Join(rootPath, common.WriteLayer)
	_, err := os.Stat(writeLayerPath)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(writeLayerPath, os.ModePerm)
		if err != nil {
			logrus.Errorf("mkdir write layer, err: %v", err)
			return err
		}
	}
	return nil
}

func createMountPoint(rootPath string, mntPath string) error {
	_, err := os.Stat(mntPath)
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(mntPath, os.ModePerm)
		if err != nil {
			logrus.Errorf("mkdir mount point, err: %v", err)
			return err
		}
	}
	dirs := fmt.Sprintf("dirs=%s%s:%s%s", rootPath, mntPath, common.WriteLayer, rootPath, common.BusyBox)
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntPath)
	if err := cmd.Run(); err != nil {
		logrus.Errorf("mnt cmd run, err: %v", err)
		return err
	}
	return nil
}

func mountVolume(rootPath, mntPath, volume string) {
	if volume != "" {
		volumes := strings.Split(volume, ":")
		if len(volumes) > 1 {
			parentPath := volumes[0]
			if _, err := os.Stat(parentPath); err != nil && os.IsNotExist(err) {
				if err := os.MkdirAll(parentPath, os.ModePerm); err != nil {
					logrus.Errorf("mkdir parent path: %s, err: %v", parentPath, err)
				}
			}
			containerPath := volumes[1]
			containerVolumePath := path.Join(mntPath, containerPath)
			if _, err := os.Stat(containerVolumePath); err != nil && os.IsNotExist(err) {
				if err := os.MkdirAll(containerVolumePath, os.ModePerm); err != nil {
					logrus.Errorf("mkdir volume path: %s, err: %v", containerVolumePath, err)
				}
			}

			dirs := fmt.Sprintf("dirs=%s", parentPath)
			cmd := exec.Command("mount", "-c", "aufs", "-o", dirs, "none", containerVolumePath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				logrus.Errorf("mount cmd run, err: %v", err)
			}
		}

	}
}

func DeleteWorkSpace(rootPath, mntPath, volume string) error {
	err := unMountPoint(mntPath)
	if err != nil {
		return err
	}
	err = deleteWriteLayer(rootPath)
	if err != nil {
		return err
	}
	deleteVolume(mntPath, volume)
	return nil
}

func unMountPoint(mntPath string) error {
	if _, err := exec.Command("umount", mntPath).CombinedOutput(); err != nil {
		logrus.Errorf("umount mnt, err: %V", err)
		return err
	}
	err := os.RemoveAll(mntPath)
	if err != nil {
		logrus.Errorf("remove mnt path, err: %v", err)
		return err
	}
	return nil
}

func deleteWriteLayer(rootPath string) error {
	writeLayerPath := path.Join(rootPath, common.WriteLayer)
	return os.RemoveAll(writeLayerPath)
}

func deleteVolume(mntPath, volume string) {
	if volume != "" {
		volumes := strings.Split(volume, ":")
		if len(volumes) > 1 {
			containerPath := path.Join(mntPath, volumes[1])
			if _, err := exec.Command("umount", containerPath).CombinedOutput(); err != nil {
				logrus.Errorf("umount container path, err: %v", err)
			}
		}
	}
}
