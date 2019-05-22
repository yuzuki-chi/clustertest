package proxmoxve

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"github.com/levigross/grequests"
	"github.com/pkg/errors"
	"github.com/yuuki0xff/clustertest/cmdutils"
	"net"
	"net/http"
	"strings"
)

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
type VMID int
type NodeVMID struct {
	NodeID NodeID
	VMID   VMID
}

// todo: add some methods

// Ticket creates an authentication ticket.
func (c *PveClient) Ticket() error {
	return cmdutils.HandlePanic(func() error {
		query := struct {
			Username string
			Password string
		}{c.User, c.Password}
		token := &apiToken{}

		err := c.reqJSON("POST", "/api2/json/access/ticket", query, token)
		if err != nil {
			return err
		}
		c.token = token
		return nil
	})
}

// CloneVM creates a copy of virtual machine/template.
func (c *PveClient) CloneVM(from, to NodeVMID, name, description string) error {
	return cmdutils.HandlePanic(func() error {
		query := struct {
			NewVMID     VMID   `url:"newid"`
			TargetNode  NodeID `url:"target"`
			Node        NodeID
			VMID        VMID `url:"vmid"`
			Full        bool
			Name        string
			Description string
		}{
			NewVMID:     to.VMID,
			TargetNode:  to.NodeID,
			Node:        from.NodeID,
			VMID:        from.VMID,
			Full:        false,
			Name:        name,
			Description: description,
		}
		url := fmt.Sprintf("/api2/json/nodes/%s/qemu/%d/clone", from.NodeID, from.VMID)
		r, err := c.req("POST", url, query)
		if err != nil {
			return err
		}
		// todo
		_ = r
		return nil
	})
}

// ListNodes returns all NodeIDs in the cluster.
func (c *PveClient) ListNodes() ([]NodeID, error) {
	var nodes []NodeID
	err := cmdutils.HandlePanic(func() error {
		var nodesInfo []struct {
			Node NodeID
		}
		err := c.reqJSON("GET", "/api2/json/nodes", nil, &nodesInfo)
		if err != nil {
			return err
		}

		for _, n := range nodesInfo {
			nodes = append(nodes, n.Node)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// ListVMs returns all NodeVMIDs in the specified node.
func (c *PveClient) ListVMs(nodeID NodeID) ([]NodeVMID, error) {
	var ids []NodeVMID
	err := cmdutils.HandlePanic(func() error {
		url := fmt.Sprintf("/api2/json/nodes/%s/qemu", nodeID)
		var vmsInfo []struct {
			VMID VMID `json:"vmid"`
		}

		err := c.reqJSON("GET", url, nil, &vmsInfo)
		if err != nil {
			return err
		}

		for _, v := range vmsInfo {
			ids = append(ids, NodeVMID{
				NodeID: nodeID,
				VMID:   v.VMID,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// ListAllVMs returns all NodeVMIDs in the cluster.
func (c *PveClient) ListAllVMs() ([]NodeVMID, error) {
	var allvms []NodeVMID
	err := cmdutils.HandlePanic(func() error {
		nodes, err := c.ListNodes()
		if err != nil {
			return err
		}

		for _, node := range nodes {
			vms, err := c.ListVMs(node)
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

func (c *PveClient) req(method, path string, query interface{}) (*grequests.Response, error) {
	url, option := c.ro(path, query)
	r, err := grequests.DoRegularRequest(method, url, option)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request")
	}
	if !r.Ok {
		return nil, errors.Errorf("received unexpected status code: %d", r.StatusCode)
	}
	return r, nil
}
func (c *PveClient) reqJSON(method, path string, query, js interface{}) error {
	r, err := c.req(method, path, query)
	if err != nil {
		return err
	}

	err = r.JSON(js)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal")
	}
	return nil
}

// ro built the RequestOptions.
// If you don't need the query string, set query to nil.
func (c *PveClient) ro(path string, query interface{}) (string, *grequests.RequestOptions) {
	url := c.buildUrl(path)
	ro := &grequests.RequestOptions{
		QueryStruct: query,
		UserAgent:   "clustertest-proxmox-ve-provisioner",
		HTTPClient:  c.httpClient(),
	}
	return url, ro
}
func (c *PveClient) buildUrl(path string) string {
	urlL := strings.TrimRight(c.Address, "/")
	urlR := strings.TrimLeft(path, "/")
	return urlL + "/" + urlR
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
