package etcd

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"strings"
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

func (e *EtcdBackend) GetVariable(group, variable string) (string, error) {
	response, err := e.client.Get(e.keyVariable(group, variable), false, false)
	if err != nil {
		return "", err
	}
	return response.Node.Value, nil
}

func (e *EtcdBackend) SetVariable(group, variable, value string) error {
	_, err := e.client.Set(e.keyVariable(group, variable), value, 0)
	return err
}

func (e *EtcdBackend) RemoveVariable(group, variable string) error {
	_, err := e.client.Delete(e.keyVariable(group, variable), false)
	return err
}

func (e *EtcdBackend) GetGroup(group string) (map[string]string, error) {
	key := e.keyGroup(group)
	response, err := e.client.Get(key, false, true)
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("/%s/", key)
	groupMap := make(map[string]string)
	for _, node := range response.Node.Nodes {
		groupMap[strings.TrimPrefix(node.Key, prefix)] = node.Value
	}

	return groupMap, nil
}

func (e *EtcdBackend) RemoveGroup(group string) error {
	_, err := e.client.Delete(e.keyGroup(group), true)
	return err
}
