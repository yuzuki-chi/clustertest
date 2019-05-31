package proxmoxve

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/levigross/grequests"
	"github.com/pkg/errors"
	"github.com/yuuki0xff/clustertest/cmdutils"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const PveMaxVMID = 999999999

var reqLogger = log.New(ioutil.Discard, "", 0)

func init() {
	if os.Getenv("CLUSTERTEST_DEBUG") != "" {
		reqLogger = log.New(os.Stderr, "proxmox-ve: ", log.LstdFlags|log.Lshortfile)
	}
}

// See https://pve.proxmox.com/pve-docs/api-viewer/
type PveClient struct {
	Address     string
	User        string
	Password    string
	Fingerprint string
	token       *apiToken
	_httpClient *http.Client
}
type apiToken struct {
	CSRFPreventionToken string `json:"CSRFPreventionToken"`
	ClusterName         string
	Ticket              string
}
type NodeID string
type VMID string
type NodeVMID struct {
	NodeID NodeID
	VMID   VMID
}
type NodeInfo struct {
	ID NodeID `json:"node"`
	// Number of available CPUs.
	MaxCPU int `json:"maxcpu"`
	// Number of available memory in bytes.
	MaxMem int `json:"maxmem"`
	// Used memory in bytes.
	Mem int
	// Current node status
	//  - unknown
	//  - online
	//  - offline
	Status string
}
type VMInfo struct {
	ID   NodeVMID
	Name string
	// Maximum usable CPUs.
	Cpus int
	// Maximum memory in bytes.
	Mem int
	// Qemu process status.
	// - stopped
	// - running
	Status string
}
type Config struct {
	// CPU cores
	CPUCores int `url:"cores" json:"cores"`
	// CPU sockets
	CPUSockets int `url:"sockets" json:"sockets"`
	// Memory size in megabytes
	Memory int `url:"memory" json:"memory"`
	// Cloud-init: user name
	User string `url:"ciuser" json:"ciuser"`
	// Cloud-init: SSH public keys
	SSHKeys string `url:"sshkeys" json:"sshkeys"`
	// Cloud-init: static IP address configuration
	// format: gw=<ipv4>,ip=<ipv4>/<CIDR>
	IPAddress string `url:"ipconfig0" json:"ipconfig0"`
}

// todo: add some methods

// Ticket creates an authentication ticket.
func (c *PveClient) Ticket() error {
	return cmdutils.HandlePanic(func() error {
		query := struct {
			Username string `url:"username"`
			Password string `url:"password"`
		}{c.User, c.Password}
		token := &apiToken{}
		data := struct{ Data *apiToken }{token}

		err := c.reqJSON("POST", "/api2/json/access/ticket", query, nil, &data)
		if err != nil {
			return err
		}
		c.token = token
		return nil
	})
}

// IDFromName finds an VM with the given name and returns an ID.
func (c *PveClient) IDFromName(name string) (NodeVMID, error) {
	vms, err := c.ListAllVMs()
	if err != nil {
		return NodeVMID{}, err
	}
	for _, v := range vms {
		if v.Name == name {
			return v.ID, nil
		}
	}
	return NodeVMID{}, errors.Errorf("not found VM: name=%+v", name)
}

// RandomVMID returns an unused VMID.
func (c *PveClient) RandomVMID() (VMID, error) {
	usedIDs := map[VMID]struct{}{}

	// Load all VMIDs and fill usedIDs.
	vms, err := c.ListAllVMs()
	if err != nil {
		return VMID(""), err
	}
	for _, v := range vms {
		usedIDs[v.ID.VMID] = struct{}{}
	}

	// Find the unused VMID.
	maxTries := 10000
	for i := 0; i < maxTries; i++ {
		vmid := VMID(fmt.Sprint(rand.Int31n(PveMaxVMID)))
		_, used := usedIDs[vmid]
		if !used {
			return vmid, nil
		}
	}
	return VMID(""), errors.New("RandomVMID: not found the unused VMID")
}

// CloneVM creates a copy of virtual machine/template.
func (c *PveClient) CloneVM(from, to NodeVMID, name, description string) error {
	return cmdutils.HandlePanic(func() error {
		query := struct {
			NewVMID     VMID   `url:"newid"`
			TargetNode  NodeID `url:"target"`
			Name        string `url:"name"`
			Description string `url:"description"`
		}{
			NewVMID:     to.VMID,
			TargetNode:  to.NodeID,
			Name:        name,
			Description: description,
		}
		url := fmt.Sprintf("/api2/json/nodes/%s/qemu/%s/clone", from.NodeID, from.VMID)
		_, err := c.req("POST", url, query, nil)
		return err
	})
}

// ResizeVolume changes size of the disk.
// The size parameter is in gigabytes.
func (c *PveClient) ResizeVolume(id NodeVMID, disk string, size int) error {
	return cmdutils.HandlePanic(func() error {
		query := struct {
			Disk string `url:"disk"`
			Size string `url:"size"`
		}{
			Disk: disk,
			Size: fmt.Sprintf("%dG", size),
		}
		url := fmt.Sprintf("/api2/json/nodes/%s/qemu/%s/resize", id.NodeID, id.VMID)
		_, err := c.req("PUT", url, query, nil)
		return err
	})
}

// UpdateConfig updates configuration of the specified VM.
func (c *PveClient) UpdateConfig(id NodeVMID, config *Config) error {
	return cmdutils.HandlePanic(func() error {
		url := fmt.Sprintf("/api2/json/nodes/%s/qemu/%s/config", id.NodeID, id.VMID)

		// Encode the SSHKeys field.
		// Server only accepts url encoded data.  We should encode the SSHKeys.
		if config.SSHKeys != "" {
			newConfig := &Config{}
			*newConfig = *config
			newConfig.SSHKeys = urlEncode(config.SSHKeys)
			config = newConfig
		}
		_, err := c.req("PUT", url, config, nil)
		return err
	})
}

// Config returns current configuration of the specified VM.
func (c *PveClient) Config(id NodeVMID) (*Config, error) {
	conf := &Config{}
	err := cmdutils.HandlePanic(func() error {
		data := struct{ Data *Config }{conf}
		url := fmt.Sprintf("/api2/json/nodes/%s/qemu/%s/config", id.NodeID, id.VMID)
		return c.reqJSON("GET", url, nil, nil, &data)
	})
	if err == nil {
		// SSHKeys field is url encoded.  We should decode it.
		conf.SSHKeys, err = urlDecode(conf.SSHKeys)
	}
	return conf, err
}

// VMInfo returns current status of the specified VM.
func (c *PveClient) VMInfo(id NodeVMID) (*VMInfo, error) {
	info := &VMInfo{}
	err := cmdutils.HandlePanic(func() error {
		data := struct{ Data *VMInfo }{info}
		url := fmt.Sprintf("/api2/json/nodes/%s/qemu/%s/status/current", id.NodeID, id.VMID)
		return c.reqJSON("GET", url, nil, nil, &data)
	})
	return info, err
}

// ListNodes returns all NodeIDs in the cluster.
func (c *PveClient) ListNodes() ([]*NodeInfo, error) {
	var nodes []*NodeInfo
	err := cmdutils.HandlePanic(func() error {
		data := struct{ Data interface{} }{&nodes}

		err := c.reqJSON("GET", "/api2/json/nodes", nil, nil, &data)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// ListVMs returns information of all VMs in the specified node.
func (c *PveClient) ListVMs(nodeID NodeID) ([]*VMInfo, error) {
	var ids []*VMInfo
	err := cmdutils.HandlePanic(func() error {
		url := fmt.Sprintf("/api2/json/nodes/%s/qemu", nodeID)
		var vmsInfo []struct {
			VMID   VMID `json:"vmid"`
			Name   string
			Cpus   int
			Maxmem int
		}
		data := struct{ Data interface{} }{&vmsInfo}

		err := c.reqJSON("GET", url, nil, nil, &data)
		if err != nil {
			return err
		}

		for _, v := range vmsInfo {
			ids = append(ids, &VMInfo{
				ID: NodeVMID{
					NodeID: nodeID,
					VMID:   v.VMID,
				},
				Name: v.Name,
				Cpus: v.Cpus,
				Mem:  v.Cpus,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// ListAllVMs returns information of all VMs in the cluster.
func (c *PveClient) ListAllVMs() ([]*VMInfo, error) {
	var allvms []*VMInfo
	err := cmdutils.HandlePanic(func() error {
		nodes, err := c.ListNodes()
		if err != nil {
			return err
		}

		for _, node := range nodes {
			vms, err := c.ListVMs(node.ID)
			if err != nil {
				return err
			}
			allvms = append(allvms, vms...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return allvms, nil
}

func (c *PveClient) req(method, path string, query interface{}, post interface{}) (*grequests.Response, error) {
	url, option := c.ro(path, query, post)
	reqLogger.Println(method, url, query, post)
	r, err := grequests.DoRegularRequest(method, url, option)
	if err != nil {
		reqLogger.Println(err)
		return nil, errors.Wrap(err, "failed to request")
	}
	reqLogger.Println(r.StatusCode)
	if !r.Ok {
		return nil, errors.Errorf("received unexpected status code: %d", r.StatusCode)
	}
	return r, nil
}
func (c *PveClient) reqJSON(method, path string, query, post, js interface{}) error {
	r, err := c.req(method, path, query, post)
	if err != nil {
		return err
	}

	reqLogger.Println(r.String())
	err = r.JSON(js)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal")
	}
	return nil
}

// ro built the RequestOptions.
// If you don't need the query string, set query to nil.
func (c *PveClient) ro(path string, query interface{}, post interface{}) (string, *grequests.RequestOptions) {
	url := c.buildUrl(path)
	ro := &grequests.RequestOptions{
		QueryStruct: query,
		Data:        interface2mapString(post),
		UserAgent:   "clustertest-proxmox-ve-provisioner",
		Cookies:     c.cookies(),
		Headers:     c.headers(),
		HTTPClient:  c.httpClient(),
	}
	return url, ro
}
func (c *PveClient) buildUrl(path string) string {
	urlL := strings.TrimRight(c.Address, "/")
	urlR := strings.TrimLeft(path, "/")
	return urlL + "/" + urlR
}
func (c *PveClient) cookies() []*http.Cookie {
	if c.token == nil {
		return nil
	}
	return []*http.Cookie{
		{Name: "PVEAuthCookie", Value: c.token.Ticket},
	}
}
func (c *PveClient) headers() map[string]string {
	if c.token == nil {
		return nil
	}
	return map[string]string{
		"CSRFPreventionToken": c.token.CSRFPreventionToken,
	}
}
func (c *PveClient) httpClient() *http.Client {
	if c._httpClient == nil {
		c._httpClient = &http.Client{}
		if c.Fingerprint != "" {
			binaryFingerprint, err := hex2bin(c.Fingerprint)
			if err != nil {
				panic(errors.Wrap(err, "invalid fingerprint"))
			}
			c._httpClient = &http.Client{
				Transport: &http.Transport{
					DialTLS: makeDialer(binaryFingerprint, true),
				},
			}
		}
	}
	return c._httpClient
}

type Dialer func(network, addr string) (net.Conn, error)

func makeDialer(fingerprint []byte, skipCAVerification bool) Dialer {
	return func(network, addr string) (net.Conn, error) {
		c, err := tls.Dial(network, addr, &tls.Config{InsecureSkipVerify: skipCAVerification})
		if err != nil {
			return nil, err
		}
		connstate := c.ConnectionState()
		for _, peercert := range connstate.PeerCertificates {
			hash := sha256.Sum256(peercert.Raw)
			if bytes.Compare(hash[0:], fingerprint) == 0 {
				// Pinned key found.
				return c, nil
			}
		}
		return nil, fmt.Errorf("pinned key not found: %s", fingerprint)
	}
}
func hex2bin(s string) ([]byte, error) {
	s = strings.ReplaceAll(s, ":", "")
	return hex.DecodeString(s)
}

func (i *VMInfo) String() string {
	return fmt.Sprintf(`<VMInfo id=%+v name="%s">`, i.ID, i.Name)
}

func interface2mapString(i interface{}) map[string]string {
	var tmp map[string]interface{}
	jsonCast(i, &tmp)

	m := map[string]string{}
	for key := range tmp {
		value := fmt.Sprint(tmp[key])
		m[key] = value
	}
	return m
}

func jsonCast(from interface{}, to interface{}) {
	b, err := json.Marshal(from)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(b, to)
	if err != nil {
		panic(err)
	}
}

func urlEncode(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}
func urlDecode(s string) (string, error) {
	return url.QueryUnescape(s)
}
