package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"simple-docker/common"
	"simple-docker/container"
	"strings"
	"text/tabwriter"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

var (
	drivers  = map[string]NetworkDriver{}
	networks = map[string]*Network{}
)

type Network struct {
	Name    string
	IpRange *net.IPNet
	Driver  string
}

type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	Network     *Network
	PortMapping []string
}

type NetworkDriver interface {
	Name() string
	Create(subnet string, name string) (*Network, error)
	Delete(network Network) error
	Connect(network *Network, endpoint *Endpoint) error
	Disconnect(network *Network, endpoint *Endpoint) error
}

func (nw *Network) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(dumpPath, os.ModePerm)
		if err != nil {
			return err
		}
	}
	nwPath := path.Join(dumpPath, nw.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Errorln("error: ", err)
		return err
	}
	defer nwFile.Close()

	nwJson, _ := json.Marshal(nw)
	_, err = nwFile.Write(nwJson)
	if err != nil {
		logrus.Errorf("write network error: %v", err)
		return err
	}
	return nil
}

func (nw *Network) remove(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil && os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path.Join(dumpPath, nw.Name))
}

func (nw *Network) load(dumpPath string) error {
	nwConfigFile, err := os.Open(dumpPath)
	if err != nil {
		return err
	}
	defer nwConfigFile.Close()
	nwJson := make([]byte, 200)
	n, err := nwConfigFile.Read(nwJson)
	if err != nil {
		return err
	}
	err = json.Unmarshal(nwJson[:n], nw)
	if err != nil {
		logrus.Errorf("json unmarshal error: %v", err)
		return err
	}
	return nil
}

func Init() error {
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	if _, err := os.Stat(common.DefaultNetworkPath); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(common.DefaultNetworkPath, os.ModePerm); err != nil {
			return err
		}
	}

	err := filepath.Walk(common.DefaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(nwPath, "/") {
			return nil
		}
		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name: nwName,
		}
		if err := nw.load(nwPath); err != nil {
			logrus.Errorf("error loading network: %v", err)
		}
		networks[nwName] = nw
		return nil
	})

	if err != nil {
		logrus.Errorf("error loading network: %v", err)
		return err
	}
	logrus.Infof("networks: %v", networks)
	return nil
}

func CreateNetwork(driver, subnet, name string) error {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		logrus.Errorf("parse cidr, err: %v", err)
	}
	ip, err := ipAllocator.Allocate(ipNet)
	if err != nil {
		logrus.Errorf("allocate ip, err: %v", err)
	}
	ipNet.IP = ip

	nw, err := drivers[driver].Create(ipNet.String(), name)
	if err != nil {
		return err
	}
	err = nw.dump(common.DefaultNetworkPath)
	if err != nil {
		logrus.Errorf("dump network, err: %v", err)
		return err
	}
	return nil
}

func Connect(networkName string, containerInfo *container.ContainerInfo) error {
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}

	ip, err := ipAllocator.Allocate(network.IpRange)
	if err != nil {
		return err
	}

	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", containerInfo.Id, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: containerInfo.PortMapping,
	}
	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}
	if err = configEndpointIpAddressAndRoute(ep, containerInfo); err != nil {
		return err
	}
	err = configPortMapping(ep, containerInfo)
	if err != nil {
		logrus.Errorf("config port mapping, err: %v", err)
		return err
	}
	return nil
}

func configEndpointIpAddressAndRoute(ep *Endpoint, cinfo *container.ContainerInfo) error {
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		logrus.Errorf("fail config endpoint: %v", err)
		return err
	}
	defer enterContainerNetns(&peerLink, cinfo)()

	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IPAddress

	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("%v,%s", ep.Network, err)
	}

	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return err
	}

	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}

	_, cdir, _ := net.ParseCIDR("0.0.0.0/0")

	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cdir,
	}

	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}
	return nil
}

func Disconnect(network Network, endpoint Endpoint) error {
	return nil
}

func enterContainerNetns(enLink *netlink.Link, cinfo *container.ContainerInfo) func() {
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		logrus.Errorf("fail to get container net namespace: %v", err)
	}

	nsFD := f.Fd()
	runtime.LockOSThread()

	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		logrus.Errorf("set link netns, err: %v", err)
	}

	origins, err := netns.Get()
	if err != nil {
		logrus.Errorf("get current netns, err: %v", err)
	}

	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		logrus.Errorf("error set netns, err: %v", err)
	}
	return func() {
		netns.Set(origins)
		origins.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}

func configPortMapping(ep *Endpoint, cinfo *container.ContainerInfo) error {
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			logrus.Errorf("port mappingm format error, %v", pm)
			continue
		}
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			logrus.Errorf("iptables output: %v", output)
			continue
		}
	}
	return nil
}

func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, _ = fmt.Fprint(w, "NAME\tIpRange\tDriver\n")
	for _, nw := range networks {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", nw.Name, nw.IpRange.String(), nw.Driver)
	}
	if err := w.Flush(); err != nil {
		logrus.Errorf("Flush error %v", err)
		return
	}
}

func DeleteNetwork(networkName string) error {
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}

	if err := ipAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return fmt.Errorf("remove network gateway ip, err: %v", err)
	}

	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("remove network gateway ip, err: %v",err)
	}
	return nw.remove(common.DefaultNetworkPath)
}
