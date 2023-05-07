package subsystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CPU struct{}

// return the name of the subsystem
func (c *CPU) Name() string {
	return "cpu"
}

// set the memory limit to this cgroup with cgroupPath
func (c *CPU) Set(cgroupPath string, res *ResourceConfig) error {	
	if subsystemCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, true); err != nil {
		return err
	} else {
		if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "cpu.shares"), []byte(res.MemoryLimit), 0644); err != nil {
			return fmt.Errorf("set cgroup cpu fail: %v", err)
		}
	}
	return nil
}

func (c *CPU) AddProcess(cgroupPath string, pid int) error {
	if subsystemCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, false); err != nil {
		return err
	} else {
		if err := ioutil.WriteFile(path.Join(subsystemCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("cgroup add process fail: %v", err)
		}
	}
	return nil
}

func (c *CPU) RemoveCgroup(cgroupPath string) error {
	if subsystemCgroupPath, err := GetCgroupPath(c.Name(), cgroupPath, false); err != nil {
		return err
	} else {
		return os.Remove(subsystemCgroupPath)
	}
}
