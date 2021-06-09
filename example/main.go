package main

import (
	"fmt"
	"time"

	"github.com/admpub/log"
	"github.com/andistributed/bus"
	"github.com/andistributed/etcd"
)

func main() {
	defer log.Close()
	go startClient("127.0.0.1")
	go startClient2("127.0.0.2")
	startClient("127.0.0.3")
}

func startClient(myIP string) {
	etcd, err := etcd.New([]string{"127.0.0.1:2379"}, time.Second*10)
	if err != nil {
		panic(err)
	}
	client := bus.NewClient("trade", myIP, etcd)

	client.PushJob("test.job", &EchoJob{})
	client.Bootstrap()
}

func startClient2(myIP string) {
	etcd, err := etcd.New([]string{"127.0.0.1:2379"}, time.Second*10)
	if err != nil {
		panic(err)
	}
	client := bus.NewClient("trade", myIP, etcd)

	client.PushJob("test.job2", &EchoJob{})
	client.Bootstrap()
}

type EchoJob struct {
}

func (*EchoJob) Execute(params string) (string, error) {
	time.Sleep(time.Second * 5)
	fmt.Println("参数:", params)
	return "ok", nil
}
