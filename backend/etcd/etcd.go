package etcd

import (
	"encoding/base64"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"strings"
	"time"
)

const (
	KeySeperator = "/"
)

type EtcdBackend struct {
	namespace, address string
	client             *etcd.Client
}

func New(namespace, protocol, address string) *EtcdBackend {

	e := &EtcdBackend{
		namespace: namespace,
		address:   address,
	}

	// Build the underlying etcd client.
	e.client = etcd.NewClient([]string{fmt.Sprintf("http://%s", e.address)})
	e.client.SyncCluster()

	// Build the Backend object.
	return e
}

func key(components ...string) string {
	return strings.Join(components, KeySeperator)
}

func (e *EtcdBackend) keyGroup(group string) string {
	return key(e.namespace, group)
}

func (e *EtcdBackend) keyVariable(group, variable string) string {
	return key(e.namespace, group, variable)
}

func (e *EtcdBackend) GetVariable(group, variable string) ([]byte, error) {
	response, err := e.client.Get(e.keyVariable(group, variable), false, false)
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(response.Node.Value)
}

func (e *EtcdBackend) setVariable(group, variable string, value []byte, ttl uint64) error {
	encodedValue := base64.StdEncoding.EncodeToString(value)
	_, err := e.client.Set(e.keyVariable(group, variable), encodedValue, ttl)
	return err
}

func (e *EtcdBackend) SetVariable(group, variable string, value []byte) error {
	return e.setVariable(group, variable, value, 0)
}

func (e *EtcdBackend) SetVariableTTL(group, variable string, value []byte, ttl time.Duration) error {
	return e.setVariable(group, variable, value, uint64(ttl.Seconds()))
}

func (e *EtcdBackend) RemoveVariable(group, variable string) error {
	_, err := e.client.Delete(e.keyVariable(group, variable), false)
	return err
}

func (e *EtcdBackend) GetGroup(group string) (map[string][]byte, error) {
	key := e.keyGroup(group)
	response, err := e.client.Get(key, false, true)
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("/%s/", key)
	groupMap := make(map[string][]byte)
	for _, node := range response.Node.Nodes {
		value, err := base64.StdEncoding.DecodeString(node.Value)
		if err != nil {
			return nil, err
		}
		groupMap[strings.TrimPrefix(node.Key, prefix)] = value
	}

	return groupMap, nil
}

func (e *EtcdBackend) RemoveGroup(group string) error {
	_, err := e.client.Delete(e.keyGroup(group), true)
	return err
}
