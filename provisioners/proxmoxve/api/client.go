package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/levigross/grequests"
	"github.com/pkg/errors"
	"github.com/yuuki0xff/clustertest/cmdutils"
	"golang.org/x/sync/semaphore"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const PveMaxVMID = 999999999
const timeout = 5 * time.Minute
const StoppedVMStatus = VMStatus("stopped")
const RunningVMStatus = VMStatus("running")
const MaxConn = 4
const MaxRetry = 30
const RetryInterval = 3 * time.Second

var reqLogger = log.New(ioutil.Discard, "", 0)

// TODO: 接続先のサーバ毎にコネクション数を制限
// sem limits max connection to API server.
var sem = semaphore.NewWeighted(MaxConn)

func init() {
	if os.Getenv("CLUSTERTEST_DEBUG") != "" {
		reqLogger = log.New(os.Stderr, "proxmox-ve: ", log.LstdFlags|log.Lshortfile)
	}
}

type PveClientOption struct {
	Address     string
	User        string
	Password    string
	Fingerprint string
}

// See https://pve.proxmox.com/pve-docs/api-viewer/
type PveClient struct {
	PveClientOption
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
type VMStatus string
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
	Status VMStatus
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

func NewPveClient(option PveClientOption) *PveClient {
	return &PveClient{
		PveClientOption: option,
	}
}

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
	var vmid VMID
	err := cmdutils.HandlePanic(func() error {
		data := struct{ Data *VMID }{&vmid}
		return c.reqJSON("GET", "/api2/json/cluster/nextid", nil, nil, &data)
	})
	return vmid, err
}

// CloneVM creates a copy of virtual machine/template.
func (c *PveClient) CloneVM(from, to NodeVMID, name, description, pool string) *Task {
	return NewTask(func(task *Task) error {
		return cmdutils.HandlePanic(func() error {
			query := struct {
				NewVMID     VMID   `url:"newid"`
				TargetNode  NodeID `url:"target"`
				Name        string `url:"name"`
				Description string `url:"description"`
				Pool        string `url:"pool"`
			}{
				NewVMID:     to.VMID,
				TargetNode:  to.NodeID,
				Name:        name,
				Description: description,
				Pool:        pool,
			}
			url := fmt.Sprintf("/api2/json/nodes/%s/qemu/%s/clone", from.NodeID, from.VMID)

			var taskID TaskID
			data := struct{ Data interface{} }{&taskID}
			err := c.reqJSON("POST", url, query, nil, &data)
			if err != nil {
				return err
			}
			task.NodeID = from.NodeID
			task.TaskID = taskID
			task.Client = c
			return err
		})
	})
}

func (c *PveClient) taskStatus(task *Task) (TaskStatus, error) {
	var status TaskStatus
	err := cmdutils.HandlePanic(func() error {
		data := struct{ Data *TaskStatus }{Data: &status}
		url := fmt.Sprintf("/api2/json/nodes/%s/tasks/%s/status", task.NodeID, task.TaskID)
		return c.reqJSON("GET", url, nil, nil, &data)
	})
	return status, err
}
func (c *PveClient) taskLog(task *Task, start, limit int) ([]string, error) {
	var lines []string
	err := cmdutils.HandlePanic(func() error {
		data := struct {
			Data []*struct {
				// Line number
				N int
				// Line text
				T string
			}
		}{}

		query := struct {
			Start int `url:"start"`
			Limit int `url:"limit"`
		}{start, limit}

		url := fmt.Sprintf("/api2/json/nodes/%s/tasks/%s/log", task.NodeID, task.TaskID)
		err := c.reqJSON("GET", url, &query, nil, &data)
		if err != nil {
			return err
		}

		for _, line := range data.Data {
			lines = append(lines, line.T)
		}
		return nil
	})
	return lines, err
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

// DeleteVM deletes the VM.
// If VM is running, it will be stop immediately and delete it.
func (c *PveClient) DeleteVM(id NodeVMID) *Task {
	return NewTask(func(task *Task) error {
		return cmdutils.HandlePanic(func() error {
			url := fmt.Sprintf("/api2/json/nodes/%s/qemu/%s", id.NodeID, id.VMID)
			taskID, err := c.reqTask("DELETE", url, nil, nil)
			if err != nil {
				return err
			}

			task.NodeID = id.NodeID
			task.TaskID = taskID
			task.Client = c
			return nil
		})
	})
}

// StartVM starts the VM.
func (c *PveClient) StartVM(id NodeVMID) *Task {
	return NewTask(func(task *Task) error {
		return cmdutils.HandlePanic(func() error {
			url := fmt.Sprintf("/api2/json/nodes/%s/qemu/%s/status/start", id.NodeID, id.VMID)

			taskID, err := c.reqTask("POST", url, nil, nil)
			if err != nil {
				return err
			}

			task.NodeID = id.NodeID
			task.TaskID = taskID
			task.Client = c
			return nil
		})
	})
}

// StopVM stops the VM immediately.  This operation is not safe.
// This is akin to pulling the power plug of a running computer and may cause VM data corruption.
func (c *PveClient) StopVM(id NodeVMID) *Task {
	return NewTask(func(task *Task) error {
		return cmdutils.HandlePanic(func() error {
			url := fmt.Sprintf("/api2/json/nodes/%s/qemu/%s/status/stop", id.NodeID, id.VMID)

			taskID, err := c.reqTask("POST", url, nil, nil)
			if err != nil {
				return err
			}

			task.NodeID = id.NodeID
			task.TaskID = taskID
			task.Client = c
			return nil
		})
	})
}
func (c *PveClient) req(method, path string, query interface{}, post interface{}) (r *grequests.Response, err error) {
	t := time.NewTicker(RetryInterval)
	for i := 0; i < MaxRetry; i++ {
		func() {
			sem.Acquire(context.Background(), 1)
			defer sem.Release(1)

			url, option := c.ro(path, query, post)
			reqLogger.Println(method, url, query, post)
			r, err = grequests.DoRegularRequest(method, url, option)
			if err != nil {
				reqLogger.Println(err)
				r = nil
				err = errors.Wrap(err, "failed to request")
				return
			}
			reqLogger.Println(r.StatusCode)
			if !r.Ok {
				status := r.StatusCode
				body := r.String()
				reqLogger.Println(body)
				r = nil
				err = NewStatusError(status, body)
				return
			}
			return
		}()

		if err != nil {
			if e, ok := err.(*StatusError); ok && (e.StatusCode == 403 || 500 <= e.StatusCode && e.StatusCode < 600) {
				// Wait for few seconds.
				reqLogger.Println("retrying ...")
				<-t.C
				continue
			}
		}
		return
	}
	return
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
func (c *PveClient) reqTask(method, path string, query, post interface{}) (TaskID, error) {
	var taskID TaskID
	data := struct{ Data interface{} }{&taskID}

	err := c.reqJSON(method, path, query, post, &data)
	return taskID, err
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
