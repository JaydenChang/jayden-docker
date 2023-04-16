package subsystem

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestGetCgroupPath(t *testing.T) {
	logrus.Infof(findCgroupMountPoint("memory"))
	logrus.Infof(findCgroupMountPoint("cpu"))
	logrus.Infof(findCgroupMountPoint("cpuset"))
}
