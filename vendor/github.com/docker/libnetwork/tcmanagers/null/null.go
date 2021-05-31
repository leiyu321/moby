package null

import (
	"net"
	"sync"

	"github.com/docker/libnetwork/datastore"
	"github.com/docker/libnetwork/discoverapi"
	"github.com/docker/libnetwork/tcmanagerapi"
)

const tcmanagertype = "null"

type tcmanager struct {
	types string
	sync.Mutex
}

func Init(dc tcmanagerapi.Callback, config map[string]string) error {
	c := tcmanagerapi.Capability{
		DataScope: datastore.LocalScope,
	}
	return dc.RegisterTcManagerDriver(tcmanagertype, &tcmanager{}, c)
}

func (m *tcmanager) Type() string {
	return tcmanagertype
}

func (m *tcmanager) CreateNetwork(id string, naddr net.IP, rate, ceil uint64) error {
	return nil
}

func (m *tcmanager) ChangeNetwork(id string, rate, ceil uint64) error {
	return nil
}

func (m *tcmanager) DeleteNetwork(id string) error {
	return nil
}

func (m *tcmanager) CreateEndpoint(nid, id string, caddr net.IP, rate, ceil uint64) error {
	return nil
}

func (m *tcmanager) ChangeEndpoint(nid, id string, rate, ceil uint64) error {
	return nil
}

func (m *tcmanager) DeleteEndpoint(nid, id string) error {
	return nil
}

func (m *tcmanager) GetEndpointClassid(nid, id string) uint32 {
	return 0
}

func (m *tcmanager) DiscoverNew(dType discoverapi.DiscoveryType, data interface{}) error {
	return nil
}

func (m *tcmanager) DiscoverDelete(dType discoverapi.DiscoveryType, data interface{}) error {
	return nil
}
