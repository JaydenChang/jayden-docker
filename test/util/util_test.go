package util

import (
	"os/exec"
	"path"
	"syscall"
	"testing"
)

func TestLookPath(t *testing.T) {
	path, err := exec.LookPath("ls")
	if err != nil {
		t.Error(err)
	}
	t.Logf("ls path: %s\n", path)
}

func TestChangeRunDir(t *testing.T) {
	err := syscall.Chdir("/root")
	if err != nil {
		t.Error(err)
	}
	cmd := exec.Command("pwd")
	bs, _ := cmd.CombinedOutput()
	t.Log(string(bs))
}

func TestPathJoin(t *testing.T) {
	newPath := path.Join("/root/", "busybox.tar")
	t.Log(newPath)
}
