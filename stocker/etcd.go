package stocker

import (
	"github.com/coreos/go-etcd/etcd"
	"sync"
)

var etcdClient = etcd.NewClient()

func Lock(c *etcd.Client) bool {
	for {
		_, success, _ := etcdClient.TestAndSet("lock", "unlock", "lock", 0)

		if success != true {
			fmt.Println("tried lock failed!")
		} else {
			return true
		}
	}
}

func Unlock(c *etcd.Client) {
	for {
		_, err := etcdClient.Set("lock", "unlock", 0)
		if err == nil {
			return
		}
		fmt.Println(err)
	}
}
