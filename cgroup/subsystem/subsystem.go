package subsystem

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

type ResourceConfig struct {
	MemoryLimit string
	CPUShare    string
	CPUSet      string
}

type Subsystem interface {
	// return the name of which type of subsystem
	Name() string
	// set a resource limit on a cgroup
	Set(cgroupPath string, res *ResourceConfig) error
	// add a processs with the pid to a group
	AddProcess(cgroupPath string, pid int) error
	// remove a cgroup
	RemoveCgroup(cgroupPath string) error
}

// instance of a subsystems
var SubsystemsInstance = []Subsystem{
	&CPU{},
	&CPUSet{},
	&Memory{},
}

// as the function name shows, find the root path of hierarchy
func FindHierarchyMountRootPath(subsystemName string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ")
		// find whether "subsystemName" appear in the last field
		// if so, then the fifth field is the path
		for _, opt := range strings.Split(fields[len(fields)-1], "/") {
			if opt == subsystemName {
				return fields[4]
			}
		}
	}
	return ""
}

// get the absolute path of a cgroup
func GetCgroupPath(subsystemName string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRootPath := FindHierarchyMountRootPath(subsystemName)
	expectedPath := path.Join(cgroupRootPath, cgroupPath)

	// find the cgroup or create a new cgroup
	if _, err := os.Stat(expectedPath); err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.Mkdir(expectedPath, 0755); err != nil {
				return "", fmt.Errorf("error when create cgroup: %v", err)
			}
		}
		return expectedPath, nil
	} else {
		return "", fmt.Errorf("cgroup path error: %v", err)
	}
}
