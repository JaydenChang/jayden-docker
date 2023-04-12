package subsystem

type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

type Subsystem interface {
	Name() string
	Set(cgroupPath string, res *ResourceConfig) error
	Remove(cgroupPath string) error
	Apply(cgroupPath string, pid int) error
}

var (
	Subsystems = []Subsystem{
		&MemorySubSystem{},
		&CpuSubSystem{},
		&CpuSetSubSystem{},
	}
)
