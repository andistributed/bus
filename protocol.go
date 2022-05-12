package bus

import (
	"encoding/json"
	"fmt"
)

type JobSnapshot struct {
	Id         string `json:"id"`
	JobId      string `json:"jobId"`
	Name       string `json:"name"`
	Ip         string `json:"ip"`
	Group      string `json:"group"`
	Cron       string `json:"cron"`
	Target     string `json:"target"`
	Params     string `json:"params"`
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

func (s *JobSnapshot) Path() string {
	return fmt.Sprintf(jobSnapshotPrefix, s.Group, s.Ip)
}

func (s *JobSnapshot) NewExecuteSnapshot() *JobExecuteSnapshot {
	return &JobExecuteSnapshot{
		Id:         s.Id,
		JobId:      s.JobId,
		Name:       s.Name,
		Group:      s.Group,
		Ip:         s.Ip,
		Cron:       s.Cron,
		Target:     s.Target,
		Params:     s.Params,
		Status:     JobExecuteDoingStatus,
		CreateTime: s.CreateTime,
		Remark:     s.Remark,
		//StartTime:  now.Format("2006-01-02 15:04:05"),
		Times: 0,
	}
}

type JobExecuteSnapshot struct {
	Id         string `json:"id" db:"id"`
	JobId      string `json:"jobId" db:"job_id"`
	Name       string `json:"name" db:"name"`
	Ip         string `json:"ip" db:"ip"`
	Group      string `json:"group" db:"group"`
	Cron       string `json:"cron" db:"cron"`
	Target     string `json:"target" db:"target"`
	Params     string `json:"params" db:"params"`
	Remark     string `json:"remark" db:"remark"`
	CreateTime string `json:"createTime" db:"create_time"`
	StartTime  string `json:"startTime" db:"start_time"`
	FinishTime string `json:"finishTime" db:"finish_time"`
	Times      int    `json:"times" db:"times"`
	Status     int    `json:"status" db:"status"`
	Result     string `json:"result" db:"result"`
}

func (s *JobExecuteSnapshot) AsJSON() []byte {
	b, _ := json.Marshal(s)
	return b
}

func (s *JobExecuteSnapshot) String() string {
	return string(s.AsJSON())
}

func (s *JobExecuteSnapshot) Path() string {
	return fmt.Sprintf(jobExecuteSnapshotPrefix, s.Group, s.Ip)
}
