package sample

import (
	"fmt"
	"net"
	"sync"

	"github.com/docker/libnetwork/datastore"
	"github.com/docker/libnetwork/discoverapi"
	"github.com/docker/libnetwork/osl"
	"github.com/docker/libnetwork/tcmanagerapi"
)

const tcmanagertype = "sample"

type tcmanager struct {
	types      string
	rate       uint64
	ifaddr     net.IP
	nets       map[string]*network
	handlePool *sync.Pool
	sync.Mutex
}

type network struct {
	id        uint16
	rate      uint64
	ceil      uint64
	eps       map[string]*endpoint
	classPool *sync.Pool
	sync.Mutex
}

type endpoint struct {
	id    uint16
	nid   string
	rate  uint64
	ceil  uint64
	caddr net.IP
	soft  bool
	sync.Mutex
}

func Init(dc tcmanagerapi.Callback, config map[string]string) error {
	c := tcmanagerapi.Capability{
		DataScope: datastore.LocalScope,
	}
	d := &tcmanager{
		types:  tcmanagertype,
		nets:   make(map[string]*network),
		ifaddr: net.ParseIP(config["ifaddr"]),
	}

	var i uint16 = 1
	d.handlePool = &sync.Pool{
		New: func() interface{} {
			i++
			return i
		},
	}

	return dc.RegisterTcManagerDriver(tcmanagertype, d, c)
}

func (m *tcmanager) Type() string {
	return tcmanagertype
}

func (m *tcmanager) CreateNetwork(id string, rate, ceil uint64) error {
	net := &network{
		id:   m.handlePool.Get().(uint16),
		rate: rate,
		ceil: ceil,
		eps:  make(map[string]*endpoint),
	}
	var i uint16 = 0
	net.classPool = &sync.Pool{
		New: func() interface{} {
			i++
			return i
		},
	}

	fmt.Println("TC: in tcmanager.createnetwork")

	if err := osl.ControlTc(osl.TC_CLASS_ADD, m.ifaddr, 1, net.id, 1, 0, 0, nil, rate, ceil); err != nil {
		return err
	}
	fmt.Println("TC: after add class for network")

	if err := osl.ControlTc(osl.TC_CGROUP_FILTER_ADD, m.ifaddr, 1, net.id, 1, net.id, 10, nil, 0, 0); err != nil {
		return err
	}
	fmt.Println("TC: after add cgroup filter for endpoints")

	if err := osl.ControlTc(osl.TC_NETWORK_FILTER_ADD, m.ifaddr, 1, net.id, 1, 0, 10, m.ifaddr, 0, 0); err != nil {
		return err
	}
	fmt.Println("TC: after add u32 filter for network")

	m.Lock()
	m.nets[id] = net
	m.Unlock()

	return nil
}

func (m *tcmanager) ChangeNetwork(id string, rate, ceil uint64) error {
	m.Lock()
	net := m.nets[id]
	m.Unlock()

	fmt.Println("TC: in tcmanager.changenetwork")
	if err := osl.ControlTc(osl.TC_CLASS_CHANGE, m.ifaddr, 1, net.id, 1, 0, 0, nil, rate, ceil); err != nil {
		return err
	}
	fmt.Println("TC: after change class for network")

	return nil
}

func (m *tcmanager) DeleteNetwork(id string) error {
	m.Lock()
	net := m.nets[id]
	m.Unlock()

	fmt.Println("TC: in tcmanager.deletenetwork")
	if err := osl.ControlTc(osl.TC_FILTER_DEL, m.ifaddr, 1, net.id, 1, 0, 10, m.ifaddr, 0, 0); err != nil {
		return err
	}
	fmt.Println("TC: after delete u32 filter for network")

	if err := osl.ControlTc(osl.TC_CGROUP_FILTER_DEL, m.ifaddr, 1, net.id, 1, net.id, 10, nil, 0, 0); err != nil {
		return err
	}
	fmt.Println("TC: after delete cgroup filter for endpoints")

	if err := osl.ControlTc(osl.TC_CLASS_DEL, m.ifaddr, 1, net.id, 1, 0, 10, m.ifaddr, net.rate, net.ceil); err != nil {
		return err
	}
	fmt.Println("TC: after delete class for network")

	m.handlePool.Put(net.id)

	m.Lock()
	delete(m.nets, id)
	m.Unlock()

	return nil
}

func (m *tcmanager) CreateEndpoint(nid, id string, caddr net.IP, rate, ceil uint64) error {
	m.Lock()
	net := m.nets[nid]
	m.Unlock()

	ep := &endpoint{
		id:    net.classPool.Get().(uint16),
		nid:   nid,
		rate:  rate,
		ceil:  ceil,
		caddr: caddr,
	}

	fmt.Println("TC: in tcmanager.createendpoint")
	if err := osl.ControlTc(osl.TC_CLASS_ADD, m.ifaddr, net.id, ep.id, 1, net.id, 10, nil, rate, ceil); err != nil {
		return err
	}
	fmt.Println("TC: after add class for endpoint")

	net.Lock()
	net.eps[id] = ep
	net.Unlock()

	return nil
}

func (m *tcmanager) ChangeEndpoint(nid, id string, rate, ceil uint64) error {
	m.Lock()
	net := m.nets[nid]
	m.Unlock()
	net.Lock()
	ep := net.eps[id]
	net.Unlock()

	fmt.Println("TC: in tcmanager.changeendpoint")
	if err := osl.ControlTc(osl.TC_CLASS_CHANGE, m.ifaddr, net.id, ep.id, 1, net.id, 10, ep.caddr, rate, ceil); err != nil {
		return err
	}
	fmt.Println("TC: after change class for endpoint")

	return nil
}

func (m *tcmanager) DeleteEndpoint(nid, id string) error {
	m.Lock()
	net := m.nets[nid]
	m.Unlock()
	net.Lock()
	ep := net.eps[id]
	net.Unlock()

	fmt.Println("TC: in tcmanager.deleteendpoint")
	if err := osl.ControlTc(osl.TC_CLASS_DEL, m.ifaddr, net.id, ep.id, 1, net.id, 10, ep.caddr, 0, 0); err != nil {
		return err
	}
	fmt.Println("TC: after delete class for endpoint")

	net.classPool.Put(ep.id)

	net.Lock()
	delete(net.eps, id)
	net.Unlock()

	return nil
}

func (m *tcmanager) GetEndpointClassid(nid, id string) uint32 {
	m.Lock()
	net := m.nets[nid]
	m.Unlock()
	net.Lock()
	ep := net.eps[id]
	net.Unlock()

	return uint32(net.id)<<16 + uint32(ep.id)
}

func (m *tcmanager) DiscoverNew(dType discoverapi.DiscoveryType, data interface{}) error {
	return nil
}

func (m *tcmanager) DiscoverDelete(dType discoverapi.DiscoveryType, data interface{}) error {
	return nil
}
