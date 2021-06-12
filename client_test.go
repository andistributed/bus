package bus

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/andistributed/etcd"
	"github.com/andistributed/forest"
	"github.com/webx-top/com"
)

var testSnapshot = JobSnapshot{
	CreateTime: time.Now().Format(`2006-01-02 15:04:05`),
	Cron:       "0 0 * * * ?",
	JobId:      "110",
	Group:      `trade`,
	Id:         forest.GenerateSerialNo(),
	Ip:         `127.0.0.8`,
	Mobile:     ``,
	Name:       `第一个任务`,
	Params:     `我是参数`,
	Remark:     `我是备注`,
	Target:     `test.job.EchoJob`,
}

func TestNewClient(t *testing.T) {

}

func _TestClient_Bootstrap(t *testing.T) {
	etcd, err := etcd.New([]string{"127.0.0.1:2379"}, time.Second*10)
	if err != nil {
		panic(err)
	}
	client := NewClient("trade", "127.0.0.8", etcd)

	client.PushJob("test.job.EchoJob", &EchoJob{})
	go func() {
		err = client.Bootstrap()
		if err != nil {
			fmt.Println(err)
		}
	}()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {

		for i := 0; i < 10; i++ {
			snapshotPath := fmt.Sprintf(jobSnapshotPrefix, "trade", "127.0.0.8")
			send := testSnapshot
			send.Id = forest.GenerateSerialNo()
			send.CreateTime = time.Now().Format(`2006-01-02 15:04:05`)
			iStr := strconv.Itoa(i)
			send.Name = `第` + iStr + `个任务`
			send.Params = `我是参数: ` + iStr
			value, err := json.Marshal(send)
			if err != nil {
				panic(err)
			}
			success, oldValue, err := etcd.PutNotExist(snapshotPath+send.Id, string(value))
			if err != nil {
				panic(err)
			}
			if !success {
				panic("手动执行任务失败: oldValue: " + string(oldValue))
			}
			time.Sleep(time.Second * time.Duration(com.RandRangeInt(1, 5)))
		}
		wg.Done()
	}()
	wg.Wait()
	client.Stop()
}

type EchoJob struct {
	Sleep time.Duration
}

func (j *EchoJob) Execute(ctx context.Context, params string) (string, error) {
	sleep := j.Sleep
	if sleep == 0 {
		sleep = time.Second * 2
	}
	time.Sleep(sleep)
	fmt.Println("参数:", params)
	return "ok", nil
}
