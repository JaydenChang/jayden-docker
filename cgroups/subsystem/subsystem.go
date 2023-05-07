package subsystem

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
