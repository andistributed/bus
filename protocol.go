package bus

import (
	"encoding/json"

	"github.com/andistributed/etcd/etcdresponse"
)

type TxResponse struct {
	*etcdresponse.TxResponse
	StateChan chan bool
}

type JobSnapshot struct {
	Id         string `json:"id"`
	JobId      string `json:"jobId"`
	Name       string `json:"name"`
	Ip         string `json:"ip"`
	Group      string `json:"group"`
	Cron       string `json:"cron"`
	Target     string `json:"target"`
	Params     string `json:"params"`
	Mobile     string `json:"mobile"`
	Remark     string `json:"remark"`
	CreateTime string `json:"createTime"`
}

func (s *JobSnapshot) AsJSON() []byte {
	b, _ := json.Marshal(s)
	return b
}

func (s *JobSnapshot) String() string {
	return string(s.AsJSON())
}

type JobExecuteSnapshot struct {
	Id         string `json:"id" xorm:"pk"`
	JobId      string `json:"jobId" xorm:"job_id"`
	Name       string `json:"name" xorm:"name"`
	Ip         string `json:"ip" xorm:"ip"`
	Group      string `json:"group" xorm:"group"`
	Cron       string `json:"cron" xorm:"cron"`
	Target     string `json:"target" xorm:"target"`
	Params     string `json:"params" xorm:"params"`
	Mobile     string `json:"mobile" xorm:"mobile"`
	Remark     string `json:"remark" xorm:"remark"`
	CreateTime string `json:"createTime" xorm:"create_time"`
	StartTime  string `json:"startTime" xorm:"start_time"`
	FinishTime string `json:"finishTime" xorm:"finish_time"`
	Times      int    `json:"times" xorm:"times"`
	Status     int    `json:"status" xorm:"status"`
	Result     string `json:"result" xorm:"result"`
}

func (s *JobExecuteSnapshot) AsJSON() []byte {
	b, _ := json.Marshal(s)
	return b
}

func (s *JobExecuteSnapshot) String() string {
	return string(s.AsJSON())
}
