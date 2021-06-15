package bus

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/admpub/log"
	"github.com/andistributed/etcd"
	"github.com/andistributed/etcd/etcdevent"
)

const jobSnapshotPrefix = "/forest/client/snapshot/%s/%s/"                // %s:client.group %s:client.ip
const jobExecuteSnapshotPrefix = "/forest/client/execute/snapshot/%s/%s/" // %s:client.group %s:client.ip
const jobKillerPrefix = "/forest/client/killer/snapshot/%s/%s/"           // %s:client.group %s:client.ip  +snapshot.id

const (
	// JobExecuteDoingStatus 执行中
	JobExecuteDoingStatus = 1
	// JobExecuteSuccessStatus 执行成功
	JobExecuteSuccessStatus = 2
	// JobExecuteUnknownStatus 未知
	JobExecuteUnknownStatus = 3
	// JobExecuteErrorStatus 执行失败
	JobExecuteErrorStatus = -1
)

type JobSnapshotProcessor struct {
	etcd                *etcd.Etcd
	snapshotPath        string
	snapshotExecutePath string
	snapshotKillerPath  string
	snapshots           chan *JobSnapshot
	jobs                map[string]Job
	jobSessions         *JobSessions
	lk                  *sync.RWMutex
}

// NewJobSnapshotProcessor new a job snapshot processor
func NewJobSnapshotProcessor(group, ip string, etcd *etcd.Etcd) *JobSnapshotProcessor {
	processor := &JobSnapshotProcessor{
		etcd:        etcd,
		snapshots:   make(chan *JobSnapshot, 100),
		jobs:        make(map[string]Job),
		jobSessions: NewJobSessions(),
		lk:          &sync.RWMutex{},
	}
	processor.snapshotPath = fmt.Sprintf(jobSnapshotPrefix, group, ip)
	processor.snapshotExecutePath = fmt.Sprintf(jobExecuteSnapshotPrefix, group, ip)
	processor.snapshotKillerPath = fmt.Sprintf(jobKillerPrefix, group, ip)
	go processor.lookup()
	go processor.watchKiller()
	return processor
}

// lookup the job snapshot
func (processor *JobSnapshotProcessor) lookup() {
	for {
		select {
		case snapshot := <-processor.snapshots:
			go processor.handleSnapshot(snapshot)
		}
	}
}

func (processor *JobSnapshotProcessor) pushJobSnapshot(snapshot *JobSnapshot) {
	processor.snapshots <- snapshot
}

func (processor *JobSnapshotProcessor) targetList() []string {
	targets := make([]string, len(processor.jobs))
	var i int
	for target := range processor.jobs {
		targets[i] = target
		i++
	}
	sort.Strings(targets)
	return targets
}

// handle the snapshot
func (processor *JobSnapshotProcessor) handleSnapshot(snapshot *JobSnapshot) {
	target := snapshot.Target
	now := time.Now()
	executeSnapshot := &JobExecuteSnapshot{
		Id:         snapshot.Id,
		JobId:      snapshot.JobId,
		Name:       snapshot.Name,
		Group:      snapshot.Group,
		Ip:         snapshot.Ip,
		Cron:       snapshot.Cron,
		Target:     snapshot.Target,
		Params:     snapshot.Params,
		Status:     JobExecuteDoingStatus,
		CreateTime: snapshot.CreateTime,
		Remark:     snapshot.Remark,
		Mobile:     snapshot.Mobile,
		StartTime:  now.Format("2006-01-02 15:04:05"),
		Times:      0,
	}

	var errResult error
	if len(target) == 0 {
		errResult = fmt.Errorf("the snapshot: %#v target is nil", snapshot)
		log.Error(errResult)
		executeSnapshot.Result = `target is nil`
		executeSnapshot.Status = JobExecuteUnknownStatus
	}
	var job Job
	if errResult == nil {
		var ok bool
		job, ok = processor.jobs[target]
		if !ok || job == nil {
			targets := processor.targetList()
			errResult = fmt.Errorf("the snapshot: %#v target %s is not found in the job list %v", snapshot, snapshot.Target, targets)
			log.Error(errResult)
			executeSnapshot.Result = `target "` + snapshot.Target + `" is not found in the job list [` + strings.Join(targets, `, `) + `]`
			executeSnapshot.Status = JobExecuteUnknownStatus
		}
	}
	value, err := json.Marshal(executeSnapshot)
	if err != nil {
		log.Errorf("[1] json.Marshal(executeSnapshot) %#v: %s", executeSnapshot, err.Error())
		return
	}

	key := processor.snapshotExecutePath + executeSnapshot.Id
	if err := processor.etcd.Put(key, string(value)); err != nil {
		log.Errorf("the snapshot:\n\t%v\nput snapshot %s execute fail: %v\n", executeSnapshot, key, err)
		return
	}

	if executeSnapshot.Status != JobExecuteDoingStatus {
		return
	}
	session := processor.jobSessions.Add(executeSnapshot.Id, job)
	result, err := session.Execute(snapshot.Params)
	processor.jobSessions.Done(executeSnapshot.Id)
	after := time.Now()
	executeSnapshot.Status = JobExecuteSuccessStatus
	executeSnapshot.Result = result
	if err != nil {
		executeSnapshot.Status = JobExecuteErrorStatus
	}

	duration := after.Sub(now)

	times := duration / time.Second
	executeSnapshot.Times = int(times)
	executeSnapshot.FinishTime = after.Format("2006-01-02 15:04:05")
	log.Debugf("the execute snapshot: %#v execute success", executeSnapshot)

	// store the execute job snapshot
	value, err = json.Marshal(executeSnapshot)
	if err != nil {
		log.Errorf("[2] json.Marshal(executeSnapshot) %#v: %s", executeSnapshot, err.Error())
		return
	}
	if err := processor.etcd.Put(key, string(value)); err != nil {
		log.Errorf("the snapshot:\n\t%v\nput update snapshot execute fail: %v", executeSnapshot, err)
		return
	}
}

func (processor *JobSnapshotProcessor) cancelExecute(snapshotId string) {
	processor.jobSessions.Cancel(snapshotId)
}

func (processor *JobSnapshotProcessor) watchKiller() {
	keyChangeEventResponse := processor.etcd.WatchWithPrefixKey(processor.snapshotKillerPath)
	for ch := range keyChangeEventResponse.Event {
		switch ch.Type {
		case etcdevent.KeyCreateChangeEvent:
		case etcdevent.KeyUpdateChangeEvent:
			log.Infof("kill job snapshot %+v", ch)
			key := string(ch.Key)
			p := strings.LastIndex(key, `/`)
			snapshotId := key[p+1:]
			processor.cancelExecute(snapshotId)
			if err := processor.etcd.Delete(key); err != nil {
				log.Errorf(`kill job snapshot failed: %s: %s`, key, err.Error())
			}
		case etcdevent.KeyDeleteChangeEvent:
		}
	}
}

// PushJob push a job to job list
func (processor *JobSnapshotProcessor) PushJob(name string, job Job) {
	processor.lk.Lock()
	defer processor.lk.Unlock()
	if _, ok := processor.jobs[name]; ok {
		log.Errorf("the job %s: %v has exist!", name, job)
		return
	}
	processor.jobs[name] = job
}
