package main

import (
	"context"
	"fmt"
	"time"

	"github.com/admpub/log"
	"github.com/andistributed/bus"
	"github.com/andistributed/etcd"
)

func main() {
	log.SetLevel(`Info`)
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

	client.PushJob("example.job", &EchoJob{}) // 注意：请确保job名称必须要在group内的所以机器上都存在
	client.Bootstrap()
}

func startClient2(myIP string) {
	etcd, err := etcd.New([]string{"127.0.0.1:2379"}, time.Second*10)
	if err != nil {
		panic(err)
	}
	client := bus.NewClient("trade", myIP, etcd)

	client.PushJob("example.job", &EchoJob{})
	client.Bootstrap()
}

type EchoJob struct {
}

func (*EchoJob) Execute(ctx context.Context, params string) (string, error) {
	time.Sleep(time.Second * 5)
	fmt.Println("参数:", params)
	return "ok", nil
}
