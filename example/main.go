package main

import (
	"fmt"
	"time"

	"github.com/andistributed/bus"
	"github.com/andistributed/etcd"
)

func main() {
	go startClient("127.0.0.1")
	go startClient("127.0.0.2")
	startClient("127.0.0.3")
}

func startClient(myIP string) {
	etcd, _ := etcd.New([]string{"127.0.0.1:2379"}, time.Second*10)
	forestClient := bus.NewForestClient("trade", myIP, etcd)

	forestClient.PushJob("test.job", &EchoJob{})
	forestClient.Bootstrap()
}

func startClient2(myIP string) {
	etcd, _ := etcd.New([]string{"127.0.0.1:2379"}, time.Second*10)
	forestClient := bus.NewForestClient("trade", myIP, etcd)

	forestClient.PushJob("test.job2", &EchoJob{})
	forestClient.Bootstrap()
}

type EchoJob struct {
}

func (*EchoJob) Execute(params string) (string, error) {

	time.Sleep(time.Second * 5)
	fmt.Println("参数:", params)
	return "ok", nil
}
