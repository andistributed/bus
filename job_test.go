package bus

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestJobCancel(t *testing.T) {
	job := NewJobSession(&EchoJob{Sleep: time.Second * 5})
	go func() {
		result, err := job.Execute(`test`)
		if err != nil {
			panic(err)
		}
		fmt.Println(`result:`, result)
	}()
	time.Sleep(time.Second * 3)
	job.Cancel()
}

func TestJobCmd(t *testing.T) {
	ctx := context.Background()
	result, err := DefaultCmdJob.Execute(ctx, `ls -alh /`)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
	//panic(``)
}
