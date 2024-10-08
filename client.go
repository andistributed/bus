package bus

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/admpub/log"
	"github.com/andistributed/etcd"
	"github.com/andistributed/etcd/etcdresponse"
)

const (
	clientPathPrefix  = "/forest/client/%s/clients/%s"  // %s:client.group %s:client.ip
	jobInfoPathPrefix = "/forest/client/%s/jobs/%%s/%s" // %s:client.group %s:jobName %s:client.ip
)

const (
	UnregisteredState = iota
	RegisteredState
)

type Client struct {
	etcd               *etcd.Etcd
	jobs               map[string]Job
	group              string
	ip                 string
	running            bool
	quit               chan bool
	state              int
	clientPath         string
	snapshotPath       string
	jobClientPath      string
	txResponseWithChan *etcdresponse.TxResponseWithChan
	snapshotProcessor  *JobSnapshotProcessor
}

// NewClient new a client
func NewClient(group, ip string, etcd *etcd.Etcd) *Client {
	c := &Client{
		etcd:  etcd,
		group: group,
		ip:    ip,
		jobs:  make(map[string]Job),
		quit:  make(chan bool),
		state: UnregisteredState,
	}
	c.PushJob("cmd", DefaultCmdJob)
	return c
}

// Bootstrap bootstrap client
func (client *Client) Bootstrap() (err error) {
	if err = client.validate(); err != nil {
		return err
	}
	client.clientPath = fmt.Sprintf(clientPathPrefix, client.group, client.ip)
	client.snapshotPath = fmt.Sprintf(jobSnapshotPrefix, client.group, client.ip)
	client.jobClientPath = fmt.Sprintf(jobInfoPathPrefix, client.group, client.ip)
	client.snapshotProcessor = NewJobSnapshotProcessor(client.group, client.ip, client.etcd)
	client.addJobs()

	go client.registerNode()
	go client.lookup()

	client.running = true
	<-client.quit
	client.running = false
	return
}

// Stop stop client
func (client *Client) Stop() {
	if client.running {
		client.quit <- true
	}
}

func (client *Client) IsRunning() bool {
	return client.running
}

// add jobs
func (client *Client) addJobs() {
	if len(client.jobs) == 0 {
		return
	}
	for name, job := range client.jobs {
		client.snapshotProcessor.PushJob(name, job)
	}
}

// pre validate params
func (client *Client) validate() error {
	if len(client.ip) == 0 {
		return errors.New("ip not allow null")
	}

	if len(client.group) == 0 {
		return errors.New("group not allow null")
	}
	return nil
}

// PushJob push a new job to job list
func (client Client) PushJob(name string, job Job) error {
	if client.running {
		return fmt.Errorf("the forest client is running not allow push a job %s", name)
	}
	if _, ok := client.jobs[name]; ok {
		return fmt.Errorf("the job %s name exist", name)
	}
	client.jobs[name] = job
	return nil
}

func (client *Client) registerNode() {

RETRY:
	var (
		txResponse *etcdresponse.TxResponseWithChan
		err        error
	)

	if client.state == RegisteredState {
		log.Infof("the forest client has already registry to: %s", client.clientPath)
		return
	}
	if txResponse, err = client.etcd.TxKeepaliveWithTTLAndChan(client.clientPath, client.ip, 10); err != nil {
		log.Warnf("the forest client fail registry to: %s", client.clientPath)
		time.Sleep(time.Second * 3)
		goto RETRY
	}

	if !txResponse.Success {
		log.Warnf("the forest client fail registry to: %s", client.clientPath)
		time.Sleep(time.Second * 3)
		goto RETRY
	}

	log.Infof("the forest client success registry to: %s", client.clientPath)
	client.state = RegisteredState
	client.txResponseWithChan = txResponse

	select {
	case <-client.txResponseWithChan.StateChan:
		client.state = UnregisteredState
		log.Warnf("the forest client fail registry to----->: %s", client.clientPath)
		goto RETRY
	}

}

// look up
func (client *Client) lookup() {

	for {
		keys, values, err := client.etcd.GetWithPrefixKeyLimit(client.snapshotPath, 50)
		if err != nil {
			log.Debugf("the forest client load job snapshot %s error: %v", client.snapshotPath, err)
			time.Sleep(time.Second * 3)
			continue
		}

		if client.state == UnregisteredState {
			time.Sleep(time.Second * 3)
			continue
		}

		if len(keys) == 0 || len(values) == 0 {
			log.Debugf("the forest client: %s load job snapshot %s is empty", client.clientPath, client.snapshotPath)
			time.Sleep(time.Second * 3)
			continue
		}

		for index, value := range values {
			key := keys[index]
			if client.state == UnregisteredState {
				time.Sleep(time.Second * 3)
				break
			}

			if err := client.etcd.Delete(string(key)); err != nil {
				log.Debugf("the forest client: %s delete job snapshot %s fail: %v", client.clientPath, client.snapshotPath, err)
				continue
			}

			if len(value) == 0 {
				log.Debugf("the forest client: %s found job snapshot %s value is nil", client.clientPath, client.snapshotPath)
				continue
			}

			snapshot := new(JobSnapshot)
			err := json.Unmarshal(value, snapshot)
			if err != nil {
				log.Debugf("the forest client: %s found job snapshot %s value is can not parse the json value: %v ", client.clientPath, client.snapshotPath, err)
				continue
			}

			// push a job snapshot
			client.snapshotProcessor.pushJobSnapshot(snapshot)
		}
	}
}
