package osl

import (
	"fmt"
	"net"

	"github.com/docker/libnetwork/ns"
	"github.com/vishvananda/netlink"
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
