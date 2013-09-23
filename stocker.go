package main

import (
	// "flag"
	"fmt"
	// "github.com/coreos/etcd/store"
	"github.com/coreos/go-etcd/etcd"
	"log"
	// "strconv"
	"os"
	"os/signal"
	"time"
)

func Lock(c *etcd.Client) bool {
	for {
		_, success, err := c.TestAndSet("lock", "unlock", "lock", 20)

		if success != true {
			fmt.Println(err)
			return false
		} else {
			return true
		}
	}
}

func Unlock(c *etcd.Client) {
	for {
		_, err := c.Set("lock", "unlock", 0)
		if err == nil {
			return
		}
		fmt.Println(err)
	}
}

func main() {

	// flag.BoolVar(&daemon, "d", false, "run daemon")
	// flag.StringVar(&container, "c", "", "container")
	// flag.IntVar(&instances, "i", 0, "number of instances")

	// flag.Parse()

	c := etcd.NewClient()

	// if container != "" {
	// 	directory := "stocker/containers/" + container

	// 	c.Set(directory+"/instances", strconv.Itoa(instances), 0)

	// 	values, err := c.Get(directory + "/running")
	// }

	// values, err := c.Get("stocker/containers/my_app")

	// log.Println(err)

	// for i, res := range values { // .. and print them out
	// 	fmt.Printf("[%d] get response: %+v\n", i, res)
	// }

	// log.Println(daemon)
	// if daemon {
	// 	log.Println("hello")
	// }

	Lock(c)

	// ch := make(chan *store.Response, 10)
	// stop := make(chan bool, 1)

	// go func() {
	// 	for {
	// 		if result, err := c.Watch("stocker/containers", 0, ch, stop); err != nil {
	// 			log.Println(err)
	// 		} else {
	// 			log.Println(result)
	// 		}
	// 	}

	// }()
	// time.Sleep(10)

	// c.Set("stocker/containers/one", "sdfasdf", 4)

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	killSignal := make(chan os.Signal, 1)
	signal.Notify(killSignal, os.Interrupt, os.Kill)

	// Block until a signal is received.

	s := <-killSignal
	log.Println("Got signal:", s)

	// Try and shut things down
	time.Sleep(1)
}
