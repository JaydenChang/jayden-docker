package container

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

func setUpMount() {
	pwd, err := os.Getwd()
	if err != nil {
		logrus.Errorf("Get current location error %v", err)
		return
	}

	logrus.Infof("Current location is %s", pwd)

	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	pivotRoot(pwd)

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
}

func pivotRoot(root string) error {
	// remount the root dir, in order to make current root and old root in different file systems
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error: %v", err)
	}

	// create 'rootfs/.pivot_root' to store old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}

	// pivot_root mount on new rootfs, old_root mount on rootfs/.pivot_root
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}

	// change current work dir to root dir
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	// umount rootfs/.rootfs_root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("umount pivot_root dir %v", err)
	}

	// del the temporary dir
	return os.Remove(pivotDir)
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

// already in container
// initiate the container
func InitProcess() error {
	// read command from pipe, will plug if write side is not ready
	cmdArray := readCommand()
	if len(cmdArray) == 0 {
		return errors.New("init process failed, cmdArray is nil")
	}

	setUpMount()
	// look for the path of container command
	// so we don't need to type "/bin/ls", but "ls"
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		logrus.Errorf("initProcess look path failed: %v", err)
		return err
	}

	// log path info
	// if you type "ls", it will be "/bin/ls"
	logrus.Infof("Find path: %v", path)
	if err := syscall.Exec(path, cmdArray, os.Environ()); err != nil {
		logrus.Errorf(err.Error())
	}

	return nil
}
