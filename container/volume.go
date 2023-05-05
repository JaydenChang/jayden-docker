package container

import (
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

func volumeUrlExtract(volume string) []string {
	// divide volume by ":"
	return strings.Split(volume, ":")
}

func MountVolume(rootURL, mntURL string, volumeURLs []string) {
	// create host file catalog
	parentURL := volumeURLs[0]
	if err := os.Mkdir(parentURL, 0777); err != nil {
		logrus.Infof("mkdir parent dir %s error. %v", parentURL, err)
	}
	// create mount point in container file system
	containerURL := volumeURLs[1]
	containerVolumeURL := mntURL + containerURL
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		logrus.Infof("mkdir container dir %s error. %v", containerVolumeURL, err)
	}
	// mount host file catalog to mount point in container
	dirs := "dirs=" + parentURL
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("mount volume failed. %v", err)
	}
}

func DeleteMountPointWithVolume(rootURL, mntURL string, volumeURLs []string) {
	// umount volume point in container
	containerURL := mntURL + volumeURLs[1]
	cmd := exec.Command("umount", containerURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("umount volume failed. %v", err)
	}
	// umount the whole point of the container
	cmd = exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("umount mountpoint failed. %v", err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Infof("remove mountpoint dir %s error %v", mntURL, err)
	}
}
