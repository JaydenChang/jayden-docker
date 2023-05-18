package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"simple-docker/container"
	"text/tabwriter"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

var (
	defaultNetworkPath = "/var/run/simple-docker/network/network"
	drivers            = map[string]NetworkDriver{}
	networks           = map[string]*Network{}
)

type Network struct {
	Name    string
	IpRange *net.IPNet
	Driver  string
}

type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"device"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	PortMapping []string         `json:"portmapping"`
	Network     *Network
}

type NetworkDriver interface {
	Name() string
	Create(subnet, name string) (*Network, error)
	Delete(network Network) error
	Connect(network *Network, endpoint *Endpoint) error
	Disconnect(network Network, endpoint *Endpoint) error
}

// create network
func CreateNetwork(driver, subnet, name string) error {
	_, cidr, _ := net.ParseCIDR(subnet)
	// allocate gateway ip by IPAM
	gatewayIP, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = gatewayIP

	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}
	// save network info
	return nw.dump(defaultNetworkPath)
}

func (nw *Network) dump(dumpPath string) error {
	// check if the path exists
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dumpPath, 0644)
		} else {
			return err
		}
	}
	// filename: network name
	nwPath := path.Join(dumpPath, nw.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Errorf("error: %v", err)
		return err
	}
	defer nwFile.Close()
	// jsonify network
	nwJson, err := json.Marshal(nw)
	if err != nil {
		logrus.Errorf("jsonify error: %v", err)
		return err
	}

	_, err = nwFile.Write(nwJson)
	if err != nil {
		logrus.Errorf("write json error: %v", err)
		return err
	}
	return nil
}

func (nw *Network) load(dumpPath string) error {
	nwConfigFile, err := os.Open(dumpPath)
	if err != nil {
		return err
	}
	defer nwConfigFile.Close()
	// read network config
	nwJson := make([]byte, 2000)
	n, err := nwConfigFile.Read(nwJson)
	if err != nil {
		return err
	}
	// unmarshal json network
	err = json.Unmarshal(nwJson[:n], nw)
	if err != nil {
		logrus.Errorf("unmarshal json error: %v", err)
		return err
	}
	return nil
}

func Connect(networkName string, cinfo *container.ContainerInfo) error {
	// get network info from networks dictionary
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("network %s not found", networkName)
	}
	// get usable ip addresses from IPAM
	ip, err := ipAllocator.Allocate(network.IpRange)
	if err != nil {
		fmt.Println("!!!! allocate ip address error")
		return err
	}
	// create endpoint
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.Id, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: cinfo.PortMapping,
	}
	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}
	if err = configEndpointIpAddressAndRoute(ep, cinfo); err != nil {
		return err
	}
	return configPortMapping(ep, cinfo)
}

func Init() error {
	// load driver
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver
	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(defaultNetworkPath, 0644)
		} else {
			return err
		}
	}
	// check all files in network config catalog
	filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name: nwName,
		}
		if err := nw.load(nwPath); err != nil {
			logrus.Errorf("error loading network %s", err)
		}
		networks[nwName] = nw
		return nil
	})
	return nil
}

func ListNetworks() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprintf(w, "NAME\tIPRange\tDriver\n")
	for _, nw := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\n", nw.Name, nw.IpRange.String(), nw.Driver)
	}
	if err := w.Flush(); err != nil {
		logrus.Errorf("flush error: %v", err)
	}
}

func DeleteNetwork(networkName string) error {
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("network %s not found", networkName)
	}
	if err := ipAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return fmt.Errorf("error remove network gateway ip: %s", err)
	}
	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("error remove network driver: %s", err)
	}
	return nw.remove(defaultNetworkPath)
}

func (nw *Network) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		return os.Remove(path.Join(dumpPath, nw.Name))
	}
}

func Disconnect(networkName string, cinfo *container.ContainerInfo) error {
	return nil
}
