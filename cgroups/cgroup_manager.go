package cgroups

import (
	"simple-docker/cgroups/subsystem"

	"github.com/sirupsen/logrus"
)

type CgroupManager struct {
	Path string // relative path, relative to the root path of the hierarchy
	// so this may cause more than one cgroup in different hierarchies
	Resource *subsystem.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

// set the three resource config subsystems to the cgroup(will create if the cgroup path is not existed)
// this may generate more than one cgroup, because those subsystem may appear in different hierarchies
func (cm CgroupManager) Set(res *subsystem.ResourceConfig) error {
	for _, subsystem := range subsystem.SubsystemsInstance {
		if err := subsystem.Set(cm.Path, res); err != nil {
			logrus.Warnf("set resource fail: %v", err)
		}
	}
	return nil
}

// add process to the cgroup path
// why should we iterate all the subsystems? we have only one cgroup
// because those subsystems may appear at different hierarchies, which will then cause more than one cgroup, 1-3 in this case.
func (cm *CgroupManager) AddProcess(pid int) error {
	for _, subsystem := range subsystem.SubsystemsInstance {
		if err := subsystem.AddProcess(cm.Path, pid); err != nil {
			logrus.Warnf("app process fail: %v", err)
		}
	}
	return nil
}

// delete the cgroup(s)
func (cm *CgroupManager) Remove() error {
	for _, subsystem := range subsystem.SubsystemsInstance {
		if err := subsystem.RemoveCgroup(cm.Path); err != nil {
			return err
		}
	}
	return nil
}
