package main

import (
	"fmt"
	"time"

	"github.com/andistributed/bus"
)

func main() {
	etcd, _ := bus.NewEtcd([]string{"127.0.0.1:2379"}, time.Second*10)
	forestClient := bus.NewForestClient("trade", "127.0.0.1", etcd)

	forestClient.PushJob("com.busgo.cat.job.EchoJob", &EchoJob{})
	forestClient.Bootstrap()
}

type EchoJob struct {
}

func (*EchoJob) Execute(params string) (string, error) {

	time.Sleep(time.Second * 5)
	fmt.Println("参数:", params)
	return "ok", nil
}
