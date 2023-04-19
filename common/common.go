package common

const (
	RootPath   = "/root/"
	MntPath    = "/root/mnt/"
	BusyBox    = "busybox"
	BusyBoxTar = "busybox.tar"
	WriteLayer = "writeLayer"
)

const (
	Running = "running"
	Stop    = "stoped"
	Exit    = "exited"
)

const (
	DefaultContainerInfoPath = "/var/run/simple-docker"
	ContainerInfoFileName    = "config.json"
	ContainerLogFileName     = "container.log"
)

const (
	EnvExecPid = "docker_pid"
	EnvExecCmd = "docker_cmd"
)

const (
	DefaultNetworkPath   = "/var/run/simple-docker/network/network"
	DefaultAllocatorPath = "/var/run/simple-docker/ipam/subnet.json"
)
