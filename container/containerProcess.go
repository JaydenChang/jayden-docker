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

func NewParentProcess(tty bool) (*exec.Cmd, *os.File) {

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
	}

	cmd.ExtraFiles = []*os.File{readPipe}
	rootURL := "/root/"
	mntURL := "/root/mnt/"
	NewWorkSpace(rootURL, mntURL)
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

func DeleteMountPoint(rootURL string, mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("%v", err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Errorf("remove dir %s error %v", mntURL, err)
	}
}

func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		logrus.Errorf("remove dir %s error %v", writeURL, err)
	}
}

func DeleteWorkSpace(rootURL, mntURL string) {
	DeleteMountPoint(rootURL, mntURL)
	DeleteWriteLayer(rootURL)
}

// extra tar to 'busybox', used as the read only layer for container
func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"
	exist, err := PathExists(busyboxURL)

	if err != nil {
		logrus.Infof("fail to judge whether dir %s exists. %v", busyboxURL, err)
	}
	if !exist {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			logrus.Errorf("mkdir dir %s error. %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			logrus.Errorf("unTar dir %s error %v", busyboxTarURL, err)
		}
	}
}

// create a unique folder as writeLayer
func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.Mkdir(writeURL, 0777); err != nil {
		logrus.Errorf("mkdir dir %s error %v", writeURL, err)
	}
}

func CreateMountPoint(rootURL string, mntURL string) {
	// create mnt folder as mount point
	if err := os.Mkdir(mntURL, 0777); err != nil {
		logrus.Errorf("mkdir dir %s error %v", mntURL, err)
	}
	// mount 'writeLayer' and 'busybox' to 'mnt'
	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	fmt.Println(dirs)
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("%v", err)
	}
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

func NewWorkSpace(rootURL, mntURL string) {
	CreateReadOnlyLayer(rootURL)
	CreateWriteLayer(rootURL)
	CreateMountPoint(rootURL, mntURL)
}
