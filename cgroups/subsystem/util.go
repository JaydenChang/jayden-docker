package subsystem

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRootPath, err := findCgroupMountPoint(subsystem)
	if err != nil {
		logrus.Errorf("find cgroup mount piunt, err: %s", err.Error())
		return "", err
	}
	cgroupTotalPath := path.Join(cgroupRootPath, cgroupPath)
	_, err = os.Stat(cgroupTotalPath)
	if err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(cgroupTotalPath, 0755); err != nil {
				return "", err
			}
		}
		return cgroupTotalPath, nil
	}
	return "", fmt.Errorf("cgroup path error")
}

func findCgroupMountPoint(subsystem string) (string, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ")
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem && len(fields) > 4 {
				return fields[4], nil
			}
		}
	}
	return "", scanner.Err()
}
