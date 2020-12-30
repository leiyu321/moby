package osl

import (
	"fmt"
	"net"

	"github.com/docker/libnetwork/ns"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
	"golang.org/x/sys/unix"
)

const (
	DEAFULT_HANDLE   = 0x10010
	DEFAULT_PARENT   = 0x10000
	DEFAULT_INTERVAL = 0x00010
)

var counter uint32

func init() {
	counter = 0
}

// type Qdisc struct{

// }

// // it's something wired!
// //Todo: change the way to use this method
// func FindIndexByAddr(nlh *netlink.Handle, addr net.IP) (int, error) {
// 	if addr == nil {
// 		return nil
// 	}
// 	ipAddr := &netlink.Addr{IPNet: addr, label: ""}
// 	return nlh.IndexByAddr(ipAddr)
// }

func AddTcBandwidth(addr net.IP, bandwidth uint64) (retErr error) {
	if addr == nil {
		return fmt.Errorf("address for TC is null! Please check it")
	}

	// path := "/proc/self/ns/net"
	// n, err := netns.GetFromPath(path)
	// if err != nil {
	// 	return fmt.Errorf("failed get network namespace %q: %v", path, err)
	// }
	// defer n.Close()

	nlh := ns.NlHandle()

	fmt.Println("TC:In AddTcBandwidth--Before IndexByAddr")

	// ifindex := FindIndexByAddr(n.nlHandle,addr)
	ipNet := &net.IPNet{IP: addr, Mask: net.CIDRMask(32, 32)}
	ifindex, err := nlh.IndexByAddr(&netlink.Addr{IPNet: ipNet})
	if err != nil {
		return err
	}

	// htb := &netlink.Htb{LinkIndex: ifindex, Handle: DEAFULT_HANDLE + DEFAULT_INTERVAL*counter, Parent: DEFAULT_PARENT, Rate2Quantum: bandwidth}
	// if err := n.nlHandle.QdiscAdd(htb); err != nil {
	// 	return err
	// }

	fmt.Printf("TC:After IndexByAddr--Before NewHtbClass--ifindex:%d\n", ifindex)
	// attrs := &netlink.ClassAttrs{LinkIndex: ifindex, Handle: DEAFULT_HANDLE + DEFAULT_INTERVAL*counter, Parent: DEFAULT_PARENT}
	// cattrs := &netlink.HtbClassAttrs{Rate: bandwidth, Ceil: bandwidth}
	htbclass := netlink.NewHtbClass(netlink.ClassAttrs{LinkIndex: ifindex, Handle: DEAFULT_HANDLE + DEFAULT_INTERVAL*counter, Parent: DEFAULT_PARENT},
		netlink.HtbClassAttrs{Rate: bandwidth, Ceil: bandwidth})

	counter += 1
	defer func() {
		if retErr != nil {
			counter -= 1
		}
	}()

	fmt.Println("TC:After NewHtbClass:Before ClassAdd")

	if err := nlh.ClassAdd(htbclass); err != nil {
		return err
	}

	fmt.Println("TC:ClassAdd successfully")
	return nil
}

const (
	TC_QDISC_ADD    = 1
	TC_QDISC_DEL    = 2
	TC_CLASS_ADD    = 3
	TC_CLASS_DEL    = 4
	TC_CLASS_CHANGE = 5
	TC_FILTER_ADD   = 6
	TC_FILTER_DEL   = 7
)

func ControlTc(flag int, ifaddr net.IP, major, minor uint16, pmajor, pminor uint16, priority uint16, caddr net.IP, rate, ceil uint64) error {
	if ifaddr == nil {
		return fmt.Errorf("addr for TC is null! Please check it")
	}

	ifindex, err := FindIndexByAddr(ifaddr)
	if err != nil {
		return err
	}

	switch flag {
	case TC_QDISC_ADD:
		return AddTcQdisc(ifindex, major, minor, pmajor, pminor)
	case TC_QDISC_DEL:
		return DeleteTcQdisc(ifindex, major, minor, pmajor, pminor)
	case TC_CLASS_ADD:
		return AddTcClass(ifindex, major, minor, pmajor, pminor, rate, ceil)
	case TC_CLASS_DEL:
		return DeleteTcClass(ifindex, major, minor)
	case TC_CLASS_CHANGE:
		return ChangeTcClass(ifindex, major, minor, pmajor, pminor, rate, ceil)
	case TC_FILTER_ADD:
		return AddTcFilter(ifindex, major, minor, pmajor, pminor, priority, caddr)
	case TC_FILTER_DEL:
		return DeleteTcFilter(ifindex, pmajor, pminor, caddr)
	default:
		return fmt.Errorf("Flag is error! No such function")
	}
}

func FindIndexByAddr(addr net.IP) (int, error) {
	if addr == nil {
		return -1, fmt.Errorf("address for finding ifindex is null! Please check it")
	}

	nlh := ns.NlHandle()
	ipNet := &net.IPNet{IP: addr, Mask: net.CIDRMask(32, 32)}
	ifindex, err := nlh.IndexByAddr(&netlink.Addr{IPNet: ipNet})
	if err != nil {
		return -1, err
	}

	return ifindex, nil
}

func AddTcQdisc(ifindex int, major, minor uint16, pmajor, pminor uint16) error {
	handle := netlink.MakeHandle(major, minor)
	parent := netlink.MakeHandle(pmajor, pminor)
	htbqdisc := netlink.NewHtb(netlink.QdiscAttrs{LinkIndex: ifindex, Handle: handle, Parent: parent})

	if err := ns.NlHandle().QdiscAdd(htbqdisc); err != nil {
		return err
	}

	return nil
}

func DeleteTcQdisc(ifindex int, major, minor uint16, pmajor, pminor uint16) error {
	handle := netlink.MakeHandle(major, minor)
	parent := netlink.MakeHandle(pmajor, pminor)
	htbqdisc := netlink.NewHtb(netlink.QdiscAttrs{LinkIndex: ifindex, Handle: handle, Parent: parent})

	if err := ns.NlHandle().QdiscDel(htbqdisc); err != nil {
		return err
	}

	return nil
}

// func ChangeTcQdisc(ifindex int, major, minor uint16) error {
// 	handle := netlink.MakeHandle(major, minor)
// 	htbqdisc := netlink.NewHtb(netlink.QdiscAttrs{LinkIndex: ifindex, Handle: handle})

// 	if err := ns.NlHandle().QdiscChange(htbqidsc); err != nil {
// 		return err
// 	}

// 	return nil
// }

func AddTcClass(ifindex int, major, minor uint16, pmajor, pminor uint16, rate, ceil uint64) error {
	classid := netlink.MakeHandle(major, minor)
	parent := netlink.MakeHandle(pmajor, pminor)
	htbclass := netlink.NewHtbClass(netlink.ClassAttrs{LinkIndex: ifindex, Handle: classid, Parent: parent},
		netlink.HtbClassAttrs{Rate: rate * 8, Ceil: ceil * 8})

	if err := ns.NlHandle().ClassAdd(htbclass); err != nil {
		return err
	}

	return nil
}

func DeleteTcClass(ifindex int, major, minor uint16) error {
	classid := netlink.MakeHandle(major, minor)
	if err := ns.NlHandle().ClassDel(&netlink.HtbClass{ClassAttrs: netlink.ClassAttrs{LinkIndex: ifindex, Handle: classid}}); err != nil {
		return err
	}

	return nil
}

func ChangeTcClass(ifindex int, major, minor uint16, pmajor, pminor uint16, rate, ceil uint64) error {
	classid := netlink.MakeHandle(major, minor)
	parent := netlink.MakeHandle(pmajor, pminor)
	htbclass := netlink.NewHtbClass(netlink.ClassAttrs{LinkIndex: ifindex, Handle: classid, Parent: parent},
		netlink.HtbClassAttrs{Rate: rate * 8, Ceil: ceil * 8})

	if err := ns.NlHandle().ClassChange(htbclass); err != nil {
		return err
	}

	return nil
}

func AddTcFilter(ifindex int, cmajor, cminor uint16, pmajor, pminor uint16, priority uint16, addr net.IP) error {
	classid := netlink.MakeHandle(cmajor, cminor)
	parent := netlink.MakeHandle(pmajor, pminor)
	var keys []nl.TcU32Key
	keys = append(keys, nl.TcU32Key{Mask: 0x00ffffff, Val: uint32(addr[0])*65536 + uint32(addr[1])*256 + uint32(addr[2]), Off: 60})
	keys = append(keys, nl.TcU32Key{Mask: 0xff000000, Val: uint32(addr[3]), Off: 64})
	sel := &nl.TcU32Sel{Flags: nl.TC_U32_TERMINAL, Nkeys: 2,
		Keys: keys}

	u32filter := &netlink.U32{FilteraAttrs: netlink.FilterAttrs{LinkIndex: ifindex, Parent: parent, Priority: priority, Protocol: unix.ETH_P_IP},
		ClassId: classid, Sel: sel}

	if err := ns.NlHandle().FilterAdd(u32filter); err != nil {
		return err
	}

	return nil
}

func DeleteTcFilter(ifindex int, pmajor, pminor uint16, addr net.IP) error {
	device, err := ns.NlHandle().LinkByIndex(ifindex)
	if err != nil {
		return err
	}

	parent := netlink.MakeHandle(pmajor, pminor)
	handle, err := ns.NlHandle().HandleByAddr(device, parent, addr)
	if err != nil {
		return err
	}
	u32filter := &netlink.U32{FilterAttrs: netlink.FilterAttrs{LinkIndex: ifindex, Handle: handle, Parent: parent}}

	if err := ns.NlHandle().FilterDel(u32filter); err != nil {
		return err
	}

	return nil
}

func ChangeTcFilter(ifindex int, cmajor, cminor uint16, pmajor, pminor uint16) error {
	return nil
}
