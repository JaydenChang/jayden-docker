package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"simple-docker/common"
	"strings"

	"github.com/sirupsen/logrus"
)

func ExecContainer(containerName string, cmdArray []string) {
	info, err := getContainerInfo(containerName)
	if err != nil {
		logrus.Errorf("get container info, err: %v", err)
	}
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = os.Setenv(common.EnvExecPid, info.Pid)
	if err != nil {
		logrus.Errorf("set env, err: %v", err)
	}
	err = os.Setenv(common.EnvExecCmd, strings.Join(cmdArray, " "))
	if err != nil {
		logrus.Errorf("set env, err: %v", err)
	}
	envs := getEnvsByPid(info.Pid)
	cmd.Env = append(os.Environ(), envs...)
	if err = cmd.Run(); err != nil {
		logrus.Errorf("exec cmd run, err: %v", err)
	}
}

func getEnvsByPid(pid string) []string {
	envFilePath := fmt.Sprintf("/proc/%s/environ", pid)
	file, err := os.Open(envFilePath)
	if err != nil {
		logrus.Errorf("open env file, path: %s, err: %v", envFilePath, err)
		return nil
	}
	bs, err := ioutil.ReadAll(file)
	if err != nil {
		logrus.Errorf("read env file, err: %v", err)
	}
	return strings.Split(string(bs), "\u0000")
}
