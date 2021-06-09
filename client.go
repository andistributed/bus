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

const clientPathPrefix = "/forest/client/%s/clients/%s"
const (
	URegistryState = iota
	RegistryState
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
	txResponseWithChan *etcdresponse.TxResponseWithChan
	snapshotProcessor  *JobSnapshotProcessor
}

// NewClient new a client
func NewClient(group, ip string, etcd *etcd.Etcd) *Client {
	return &Client{
		etcd:  etcd,
		group: group,
		ip:    ip,
		jobs:  make(map[string]Job),
		quit:  make(chan bool),
		state: URegistryState,
	}
}

// Bootstrap bootstrap client
func (client *Client) Bootstrap() (err error) {
	if err = client.validate(); err != nil {
		return err
	}
	client.clientPath = fmt.Sprintf(clientPathPrefix, client.group, client.ip)
	client.snapshotPath = fmt.Sprintf(jobSnapshotPrefix, client.group, client.ip)
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
	if client.ip == "" {
		return errors.New("ip not allow null")
	}

	if client.group == "" {
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

	if client.state == RegistryState {
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
	client.state = RegistryState
	client.txResponseWithChan = txResponse

	select {
	case <-client.txResponseWithChan.StateChan:
		client.state = URegistryState
		log.Warnf("the forest client fail registry to----->: %s", client.clientPath)
		goto RETRY
	}

}

// look up
func (client *Client) lookup() {

	for {
		keys, values, err := client.etcd.GetWithPrefixKeyLimit(client.snapshotPath, 50)
		if err != nil {
			log.Debugf("the forest client load job snapshot error: %v", err)
			time.Sleep(time.Second * 3)
			continue
		}

		if client.state == URegistryState {
			time.Sleep(time.Second * 3)
			continue
		}

		if len(keys) == 0 || len(values) == 0 {
			log.Debugf("the forest client: %s load job snapshot is empty", client.clientPath)
			time.Sleep(time.Second * 3)
			continue
		}

		for i := 0; i < len(values); i++ {
			key := keys[i]
			if client.state == URegistryState {
				time.Sleep(time.Second * 3)
				break
			}

			if err := client.etcd.Delete(string(key)); err != nil {
				log.Debugf("the forest client: %s delete job snapshot fail: %v", client.clientPath, err)
				continue
			}

			value := values[i]
			if len(value) == 0 {
				log.Debugf("the forest client: %s found job snapshot value is nil", client.clientPath)
				continue
			}

			snapshot := new(JobSnapshot)
			err := json.Unmarshal(value, snapshot)
			if err != nil {
				log.Debugf("the forest client: %s found job snapshot value is cant not parse the json value：%v ", client.clientPath, err)
				continue
			}

			// push a job snapshot
			client.snapshotProcessor.pushJobSnapshot(snapshot)
		}
	}
}
