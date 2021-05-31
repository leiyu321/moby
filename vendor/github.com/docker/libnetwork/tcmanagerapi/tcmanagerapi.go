package tcmanagerapi

import (
	"net"

	"github.com/docker/docker/pkg/plugingetter"
	"github.com/docker/libnetwork/discoverapi"
)

const (
	DefaultTCManager = "sample"
	NullTCManager    = "null"
)

type TcManager interface {
	discoverapi.Discover

	Type() string
	CreateNetwork(id string, naddr net.IP, rate, ceil uint64) error
	ChangeNetwork(id string, rate, ceil uint64) error
	DeleteNetwork(id string) error
	CreateEndpoint(nid, id string, caddr net.IP, rate, ceil uint64) error
	ChangeEndpoint(nid, id string, rate, ceil uint64) error
	DeleteEndpoint(nid, id string) error
	GetEndpointClassid(nid, id string) uint32
}

type Callback interface {
	GetPluginGetter() plugingetter.PluginGetter
	RegisterTcManagerDriver(name string, driver TcManager, capability Capability) error
}

type Capability struct {
	DataScope         string
	ConnectivityScope string
}
