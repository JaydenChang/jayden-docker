package subsystem

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/sirupsen/logrus"
)

type CpuSetSubSystem struct{}

func (*CpuSetSubSystem) Name() string {
	return "cpuset"
}

func (c *CpuSetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	subsystemCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, true)
	if err != nil {
		logrus.Errorf("get %s path, err: %v", cgroupPath, err)
		return err
	}
	if res.CpuSet != "" {
		err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpuset.cpus"), []byte(res.CpuSet), 0644)
		if err != nil {
			logrus.Errorf("failed to write file cpuset.cpu, err: %+v", err)
			return err
		}
	}
	return nil
}

func (c *CpuSetSubSystem) Remove(cgroupPath string) error {
	subsystemCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, true)
	if err != nil {
		return err
	}
	return os.RemoveAll(subsystemCgroupPath)
}

func (c *CpuSetSubSystem) Apply(cgroupPath string, pid int) error {
	subsystemCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, true)
	if err != nil {
		return err
	}
	tasksPath := path.Join(subsystemCgroupPath, "tasks")
	err = ioutil.WriteFile(tasksPath, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		logrus.Errorf("write pid to tasks, path: %s, pid: %d, err: %v", tasksPath, pid, err)
		return err
	}
	return nil
}
