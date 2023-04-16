package container

import (
	"os"
	"os/exec"
	"simple-docker/common"
	"syscall"

	"github.com/sirupsen/logrus"
)

func NewParentProcess(tty bool,volume string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, _ := os.Pipe()
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	err := NewWorkSpace(common.RootPath, common.MntPath, volume)
	if err != nil {
		logrus.Errorf("new work space, err: %v", err)
	}
	cmd.Dir = common.MntPath
	return cmd, writePipe
}
