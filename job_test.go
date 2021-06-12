package bus

import (
	"fmt"
	"testing"
	"time"
)

func TestJobCancel(t *testing.T) {
	job := NewJobSession(&EchoJob{Sleep: 5})
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
