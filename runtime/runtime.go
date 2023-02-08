package runtime

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/coreos/go-iptables/iptables"
	"github.com/zbh888/classifier-runtime/lib/slicing"
	"github.com/zbh888/classifier-runtime/lib/u32"
)

func runTC(args ...string) error {
	cmd := exec.Command("/sbin/tc", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if nil != err {
		log.Println("Error running /sbin/tc:", err)
		return err
	}
	return nil
}

var runtime *Runtime
var classificator *Classification

func FindNetwork(ipV4 string) (string, error) {

	ipV4 = fmt.Sprintf("%s/24", ipV4)
	_, ipnet, err := net.ParseCIDR(ipV4)
	if err != nil {
		log.Println("Impossible to parse the address")
		return "", err
	}

	cmd := fmt.Sprintf("ip route | grep %s | awk '{print $3}'", ipnet.String())
	output, _ := exec.Command("bash", "-c", cmd).Output()

	ipV4 = fmt.Sprintf("%s/24", strings.TrimSpace(string(output)))
	_, ipnet, err = net.ParseCIDR(ipV4)
	if err != nil {
		log.Println("Impossible to parse the address")
		return "", err
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Error reading interfaces: %+v\n", err.Error())
		return "", err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Printf("localAddresses: %+v\n", err.Error())
			return "", err
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPNet:
				if ipnet.Contains(v.IP.To4()) {
					return v.IP.To4().String(), nil
				}
			}

		}
	}

	return "", nil
}

func FindInterface(ipV4 string) (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Error reading interfaces: %+v\n", err.Error())
		return "", err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Printf("localAddresses: %+v\n", err.Error())
			return "", err
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPNet:
				if v.IP.To4().String() == ipV4 {
					return i.Name, nil
				}
			}

		}
	}

	return "", errors.New("ERROR interface not found")
}

type ClassID struct {
	Root int
	VoIP int
	Rest int
}

type Classification struct {
	Classes map[uint8]ClassID
	Queues  map[string]int
	Ifaces  []string
}

func (c *Classification) calculateBurst(rate int) string {
	// var b float64 = 18.0 / (30.0 / float64(rate))
	// var burst float64 = math.Ceil(b)
	// return fmt.Sprintf("%dk", int(burst*4))
	return "30k"
}

func (c *Classification) queueIsSet(iface string) bool {
	for _, q := range c.Ifaces {
		if q == iface {
			return true
		}
	}
	return false
}

func (c *Classification) buildTree(Controls []ADMControl) error {

	for k, adm := range Controls {
		log.Println(k, adm.Endpoint, adm.SliceID, adm.Throughput)
		iface, err := FindInterface(adm.Endpoint)
		if err != nil {
			return err
		}

		if !c.queueIsSet(iface) {
			runTC("qdisc", "add", "dev", iface, "root", "handle", "1:", "htb", "r2q", "10")
			c.Ifaces = append(c.Ifaces, iface)
			c.Queues[iface] = 1
		}

		voipClass := c.Queues[iface]*10 + 2
		restClass := c.Queues[iface]*10 + 1
		c.Classes[adm.SliceID] = ClassID{
			VoIP: voipClass,
			Rest: restClass,
			Root: c.Queues[iface],
		}

		rootClass := fmt.Sprintf("1:%d", c.Queues[iface])
		runTC("class", "add", "dev", iface, "parent", "1:", "classid", rootClass, "htb", "rate", fmt.Sprintf("%dmbit", adm.Throughput), "burst", c.calculateBurst(adm.Throughput), "cburst", c.calculateBurst(adm.Throughput), "mtu", "1500")

		voipClasss := fmt.Sprintf("1:%d", voipClass)
		restClasss := fmt.Sprintf("1:%d", restClass)
		runTC("class", "add", "dev", iface, "parent", rootClass, "classid", voipClasss, "htb", "rate", "2mbit", "prio", "0", "burst", "3k", "cburst", "3k", "mtu", "1500")
		runTC("qdisc", "add", "dev", iface, "parent", voipClasss, "sfq", "perturb", "10")
		runTC("class", "add", "dev", iface, "parent", rootClass, "classid", restClasss, "htb", "rate", fmt.Sprintf("%dmbit", adm.Throughput-1), "ceil", fmt.Sprintf("%dmbit", adm.Throughput-1), "prio", "1", "burst", c.calculateBurst(adm.Throughput-1), "cburst", c.calculateBurst(adm.Throughput-1), "mtu", "1500")
		runTC("qdisc", "add", "dev", iface, "parent", restClasss, "sfq", "perturb", "10")

		c.Queues[iface]++
	}

	return nil
}

func (*Classification) getQueue() error {
	return nil
}

type Runtime struct {
	iptable *iptables.IPTables
	Aware   bool
}

func (r *Runtime) AdmissionControl(adm *ADM) error {

	if adm.Aware {
		log.Println("Setting up the data-plane for Slice Awareness")
		classificator.buildTree(adm.Controls)
		r.Aware = true
		return nil
	}

	log.Println("Slice Awareness Deactivated")
	return nil
}

func (r *Runtime) NewPDU(pdu *PDU) error {

	if !r.Aware {
		return nil
	}

	label := slicing.NewLabel(pdu.SliceID, true, 0x2e)

	iface, err := FindInterface(pdu.Ingress)
	if err != nil {
		return err
	}
	oface, err := FindInterface(pdu.Endpoint)
	if err != nil {
		return err
	}

	var dscp5GString string = fmt.Sprintf("0x%s", hex.EncodeToString([]byte{0x2e}))
        
	print("BOHAN MADE A CHANGE")
	
	u32GTP := u32.BuildMatchGTP(pdu.TEID, pdu.IPv4, pdu.IsRAN)

	// Add marking to incoming packets
	runtime.iptable.AppendUnique(MANGLE_TABLE, PIPE_CHAIN, "-i", iface, "-p", "udp", "--dport", "2152", "--sport", "2152", "-j", "MARK", "--set-xmark", label.GeneratePipe())
	runtime.iptable.AppendUnique(MANGLE_TABLE, QOS_CHAIN, "-i", iface, "-p", "udp", "--dport", "2152", "--sport", "2152", "-m", "dscp", "--dscp", dscp5GString, "-j", "MARK", "--set-xmark", label.GenerateDSCP())
	runtime.iptable.AppendUnique(MANGLE_TABLE, SLICE_CHAIN, "-i", iface, "-m", "u32", "--u32", u32GTP, "-j", "MARK", "--set-xmark", label.GenerateSliceID())

	voipClass := classificator.Classes[pdu.SliceID].VoIP
	restClass := classificator.Classes[pdu.SliceID].Rest
	voipClasss := fmt.Sprintf("1:%d", voipClass)
	restClasss := fmt.Sprintf("1:%d", restClass)

	runtime.iptable.AppendUnique(MANGLE_TABLE, POSTROUTING_CHAIN, "-o", oface, "-m", "mark", "--mark", label.GeneratePipeSliceID(), "-j", "CLASSIFY", "--set-class", restClasss)
	runtime.iptable.AppendUnique(MANGLE_TABLE, POSTROUTING_CHAIN, "-o", oface, "-m", "mark", "--mark", label.Generate(), "-j", "CLASSIFY", "--set-class", voipClasss)

	return nil
}

func (*Runtime) Run() {
	lis := fmt.Sprintf("%s:%d", RuntimeConfig.Sbi.RegisterIPv4, RuntimeConfig.Sbi.Port)
	Router.Run(lis)
}

func InitRuntime(flush bool, config string) error {
	err := ParseConf(config)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	table, err := iptables.New()
	if err != nil {
		return err
	}

	table.ClearAndDeleteChain(MANGLE_TABLE, PIPE_CHAIN)
	table.ClearAndDeleteChain(MANGLE_TABLE, SLICE_CHAIN)
	table.ClearAndDeleteChain(MANGLE_TABLE, SLICE_CHAIN)
	table.ClearAndDeleteChain(NAT_TABLE, PREROUTING_CHAIN)

	table.NewChain(MANGLE_TABLE, PIPE_CHAIN)
	table.NewChain(MANGLE_TABLE, SLICE_CHAIN)
	table.NewChain(MANGLE_TABLE, QOS_CHAIN)

	table.AppendUnique(MANGLE_TABLE, PREROUTING_CHAIN, "-j", PIPE_CHAIN)
	table.AppendUnique(MANGLE_TABLE, PREROUTING_CHAIN, "-j", SLICE_CHAIN)
	table.AppendUnique(MANGLE_TABLE, PREROUTING_CHAIN, "-j", QOS_CHAIN)

	classificator = &Classification{
		Classes: make(map[uint8]ClassID),
		Queues:  make(map[string]int),
	}

	runtime = &Runtime{
		iptable: table,
		Aware:   false,
	}
	InitRouter(false, false)
	return nil
}

func Run() {
	runtime.Run()
}
