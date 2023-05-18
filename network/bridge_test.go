package network

import (
	"log"
	"net"
	"simple-docker/container"
	"strings"
	"testing"

	"github.com/vishvananda/netlink"
)

func TestBridgeInit(t *testing.T) {
	d := BridgeNetworkDriver{}
	_, err := d.Create("192.168.0.1/24", "testbridge")
	t.Logf("err: %v", err)
}

func TestBridgeConnect(t *testing.T) {
	ep := Endpoint{
		ID: "testcontainer",
	}
	n := Network{
		Name: "testbridge",
	}
	d := BridgeNetworkDriver{}
	err := d.Connect(&n, &ep)
	t.Logf("err: %v", err)
}

func TestNetworkConnect(t *testing.T) {
	d := BridgeNetworkDriver{}
	n, err := d.Create("192.168.0.1/24", "testbridge")
	t.Logf("network: %v", n)
	t.Logf("err: %v", err)
	Init()
	networks[n.Name] = n
	cInfo := &container.ContainerInfo{
		Id:  "testcontainer",
		Pid: "15438",
	}
	err = Connect(n.Name, cInfo)
	t.Logf("err: %v", err)
}

func TestLoad(t *testing.T) {
	n := Network{
		Name: "testbridge",
	}
	n.load("/var/run/simple-docker/network/networks/testbridge")
	t.Logf("network loaded: %v", n)
}

func TestAddBridge(t *testing.T) {
	bridgeName := "testbridge"
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		log.Printf("error:%v\n", err)
	}
	// create *netlink.Bridge object
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName

	br := &netlink.Bridge{LinkAttrs: la}
	// 等于 ip link add name testbridge type bridge
	if err := netlink.LinkAdd(br); err != nil {
		t.Errorf("Bridge creation failed for bridge %s: %v", bridgeName, err)
	}
}
