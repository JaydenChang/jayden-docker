package subsystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CPUSet struct{}

// return the name of the subsystem
func (c *CPUSet) Name() string {
	return "CPUSet"
}

func (c *CPUSet) Set(cgroupPath string, res *ResourceConfig) error {
	if subsystemCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, true); err != nil {
		return err
	} else {
		if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpuset.cpus"), []byte(res.CPUSet), 0644); err != nil {
			return fmt.Errorf("set cgroup cpuset failed: %v", err)
		}
	}
	return nil
}

func (c *CPUSet) AddProcess(cgroupPath string, pid int) error {
	if subsystemCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, false); err != nil {
		return err
	} else {
		if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("cgroup add cpuset process failed: %v", err)
		}
	}
	return nil
}

func (c *CPUSet) RemoveCgroup(cgroupPath string) error {
	if subsystemCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, false); err != nil {
		return err
	} else {
		return os.Remove(path.Join(subsystemCgroupPath))
	}
}
