package bus

import (
	"fmt"
	"testing"
	"time"
)

/**
 * json :

  {
  "createTime": "2019-12-24 17:43:04",
  "cron": "0 0 * * * ?",
  "group": "trade",
  "id": "d176488a-865f-4840-91c8-757ec6da20bf",
  "ip": "127.0.0.1",
  "jobId": "110",
  "mobile": "18758586911",
  "name": "第一个任务",
  "params": "我是参数",
  "remark": "备注",
  "target": "com.busgo.cat.job.EchoJob"
}


*/
func TestNewForestClient(t *testing.T) {

}

func TestForestClient_Bootstrap(t *testing.T) {

	etcd, _ := NewEtcd([]string{"127.0.0.1:2379"}, time.Second*10)
	forestClient := NewForestClient("trade", "127.0.0.1", etcd)

	forestClient.PushJob("com.busgo.cat.job.EchoJob", &EchoJob{})
	forestClient.Bootstrap()
}

type EchoJob struct {
}

func (*EchoJob) execute(params string) (string, error) {

	time.Sleep(time.Second * 5)
	fmt.Println("参数:", params)
	return "ok", nil
}
