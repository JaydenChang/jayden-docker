package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

var (
	RUNNING             string = "running"
	STOP                string = "stopped"
	Exit                string = "exited"
	DefaultInfoLocation string = "/var/run/simple-docker/%s/"
	ConfigName          string = "config.json"
	ContainerLogFile    string = "container.log"
	RootURL             string = "/root/"
	MntURL              string = "/root/mnt/%s/"
	WriteLayerURL       string = "/root/writeLayer/%s"
)

type ContainerInfo struct {
	Pid         string `json:"pid"`        //容器的init进程在宿主机上的 PID
	Id          string `json:"id"`         //容器Id
	Name        string `json:"name"`       //容器名
	Command     string `json:"command"`    //容器内init运行命令
	CreatedTime string `json:"createTime"` //创建时间
	Status      string `json:"status"`     //容器的状态
	Volume      string `json:"volume"`
	PortMapping []string `json:"portmapping"` 
}

func volumeUrlExtract(volume string) []string {
	// divide volume by ":"
	return strings.Split(volume, ":")
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func NewWorkSpace(volume, imageName, containerName string) {
	CreateReadOnlyLayer(imageName)
	CreateWriteLayer(containerName)
	CreateMountPoint(containerName, imageName)
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(volumeURLs, containerName)
			logrus.Infof("%q", volumeURLs)
		} else {
			logrus.Infof("volume parameter input is not correct")
		}
	}
}

func CreateReadOnlyLayer(imageName string) error {
	unTarFolderURL := RootURL + "/" + imageName + "/"
	imageURL := RootURL + "/" + imageName + ".tar"
	exist, err := PathExists(unTarFolderURL)

	if err != nil {
		logrus.Infof("fail to judge whether dir %s exists. %v", unTarFolderURL, err)
		return err
	}
	if !exist {
		if err := os.MkdirAll(unTarFolderURL, 0777); err != nil {
			logrus.Errorf("mkdir dir %s error. %v", unTarFolderURL, err)
			return err
		}
		if _, err := exec.Command("tar", "-xvf", imageURL, "-C", unTarFolderURL).CombinedOutput(); err != nil {
			logrus.Errorf("unTar dir %s error %v", unTarFolderURL, err)
			return err
		}
	}
	return nil
}

func CreateWriteLayer(containerName string) {
	writeUrl := fmt.Sprintf(WriteLayerURL, containerName)
	if err := os.MkdirAll(writeUrl, 0777); err != nil {
		logrus.Infof("Mkdir write layer dir %s error. %v", writeUrl, err)
	}
}

func CreateMountPoint(containerName, imageName string) error {
	// create mnt folder as mount point
	mntURL := fmt.Sprintf(MntURL, containerName)
	if err := os.MkdirAll(mntURL, 0777); err != nil {
		logrus.Errorf("mkdir dir %s error %v", mntURL, err)
		return err
	}
	// mount 'writeLayer' and 'busybox' to 'mnt'
	tmpWriteLayer := fmt.Sprintf(WriteLayerURL, containerName)
	tmpImageLocation := RootURL + "/" + imageName
	dirs := "dirs=" + tmpWriteLayer + ":" + tmpImageLocation
	_, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL).CombinedOutput()
	if err != nil {
		logrus.Errorf("run command for creating mount point failed: %v", err)
		return err
	}
	return nil
}

func MountVolume(volumeURLs []string, containerName string) error {
	// create host file catalog
	parentURL := volumeURLs[0]
	if err := os.Mkdir(parentURL, 0777); err != nil {
		logrus.Infof("mkdir parent dir %s error. %v", parentURL, err)
	}
	// create mount point in container file system
	containerURL := volumeURLs[1]
	mntURL := fmt.Sprintf(MntURL, containerName)
	containerVolumeURL := mntURL + "/" + containerURL
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		logrus.Infof("mkdir container dir %s error. %v", containerVolumeURL, err)
	}
	// mount host file catalog to mount point in container
	dirs := "dirs=" + parentURL
	_, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL).CombinedOutput()
	if err != nil {
		logrus.Errorf("mount volume failed. %v", err)
		return err
	}
	return nil
}

func DeleteMountPointWithVolume(volumeURLs []string, containerName string) error {
	// umount volume point in container
	mntURL := fmt.Sprintf(MntURL, containerName)
	containerURL := mntURL + "/" + volumeURLs[1]
	if _, err := exec.Command("umount", containerURL).CombinedOutput(); err != nil {
		logrus.Errorf("umount volume failed. %v", err)
		return err
	}
	// umount the whole point of the container
	_, err := exec.Command("umount", mntURL).CombinedOutput()
	if err != nil {
		logrus.Errorf("umount mountpoint failed. %v", err)
		return err
	}
	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Infof("remove mountpoint dir %s error %v", mntURL, err)
	}
	return nil
}

func DeleteMountPoint(containerName string) error {
	mntURL := fmt.Sprintf(MntURL, containerName)
	_, err := exec.Command("umount", mntURL).CombinedOutput()
	if err != nil {
		logrus.Errorf("%v", err)
		return err
	}
	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Errorf("remove dir %s error %v", mntURL, err)
		return err
	}
	return nil
}

func DeleteWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerURL, containerName)
	if err := os.RemoveAll(writeURL); err != nil {
		logrus.Errorf("remove dir %s error %v", writeURL, err)
	}
}

func DeleteWorkSpace(volume, containerName string) {
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(volumeURLs, containerName)
		} else {
			DeleteMountPoint(containerName)
		}
	} else {
		DeleteMountPoint(containerName)
	}
	DeleteWriteLayer(containerName)
}

func NewParentProcess(tty bool, volume string, containerName, imageName string, envSlice []string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := os.Pipe()

	if err != nil {
		logrus.Errorf("New Pipe Error: %v", err)
		return nil, nil
	}
	// create a new command which run itself
	// the first arguments is `init` which is in the "container/init.go" file
	// so, the <cmd> will be interpret as "docker init <cmdArray>"
	cmd := exec.Command("/proc/self/exe", "init")

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	cmd.Stdin = os.Stdin
	if tty {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
		if err := os.MkdirAll(dirURL, 0622); err != nil {
			logrus.Errorf("NewParentProcess mkdir %s error %v", dirURL, err)
			return nil, nil
		}
		stdLogFilePath := dirURL + ContainerLogFile
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			logrus.Errorf("NewParentProcess create file %s error %v", stdLogFilePath, err)
			return nil, nil
		}
		cmd.Stdout = stdLogFile
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	cmd.Env = append(os.Environ(), envSlice...)
	NewWorkSpace(volume, imageName, containerName)
	cmd.Dir = fmt.Sprintf(MntURL, containerName)

	return cmd, writePipe
}
