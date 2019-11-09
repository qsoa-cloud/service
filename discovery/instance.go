package discovery

const (
	InstanceStatusUnknown InstanceStatus = iota
	InstanceStatusReady
)

type Instance struct {
	Addr   string
	Status InstanceStatus
	Value  string
}

type InstanceStatus uint8
