package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/admpub/log"
	"github.com/andistributed/bus"
	"github.com/andistributed/etcd"
	"github.com/andistributed/etcd/etcdconfig"
	"github.com/webx-top/com"
)

var (
	currentIP     = com.Getenv(`FOREST_JOB_IP`, `127.0.0.1`)
	etcdEndpoints = com.Getenv("ETCD_ENDPOINTS", "127.0.0.1:2379")
	etcdDialTime  = com.GetenvInt("ETCD_DIAL_TIMEOUT", 5)
	etcdUsername  = com.Getenv("ETCD_USERNAME")
	etcdPassword  = com.Getenv("ETCD_PASSWORD")
	etcdCertFile  = os.Getenv("ETCD_CERT_FILE") // ca.crt
	etcdKeyFile   = os.Getenv("ETCD_KEY_FILE")  // ca.key
	jobGroup      = `default`
)

func init() {
	flag.StringVar(&etcdCertFile, "etcd-cert", etcdCertFile, "--etcd-cert file (也可以通过环境变量ETCD_CERT_FILE来指定)")
	flag.StringVar(&etcdKeyFile, "etcd-key", etcdKeyFile, "--etcd-key file (也可以通过环境变量ETCD_KEY_FILE来指定)")
	flag.StringVar(&etcdEndpoints, "etcd-endpoints", etcdEndpoints, "--etcd-endpoints "+etcdEndpoints+" (也可以通过环境变量ETCD_ENDPOINTS来指定)")
	flag.IntVar(&etcdDialTime, "etcd-dialtimeout", etcdDialTime, "--etcd-dialtimeout "+strconv.Itoa(etcdDialTime)+" (也可以通过环境变量ETCD_DIAL_TIMEOUT来指定)")
	flag.StringVar(&etcdUsername, "etcd-username", etcdUsername, "--etcd-username root (也可以通过环境变量ETCD_USERNAME来指定)")
	flag.StringVar(&etcdPassword, "etcd-password", etcdPassword, "--etcd-password root (也可以通过环境变量ETCD_PASSWORD来指定)")
	flag.StringVar(&jobGroup, "group", jobGroup, "--group "+jobGroup)
	flag.StringVar(&currentIP, "current-ip", currentIP, "--current-ip "+currentIP+" (也可以通过环境变量FOREST_JOB_IP来指定)")
	flag.Parse()
}

func main() {
	log.SetLevel(`Info`)
	defer log.Close()
	startClient(currentIP)
}

func startClient(myIP string) error {
	endpoint := strings.Split(etcdEndpoints, ",")
	dialTime := time.Duration(etcdDialTime) * time.Second
	var etcdOpts []etcdconfig.Configer
	if len(etcdCertFile) > 0 && len(etcdKeyFile) > 0 {
		etcdOpts = append(etcdOpts, etcdconfig.TLSFile(etcdCertFile, etcdKeyFile))
	}
	if len(etcdUsername) > 0 {
		etcdOpts = append(etcdOpts, etcdconfig.Username(etcdUsername))
	}
	if len(etcdPassword) > 0 {
		etcdOpts = append(etcdOpts, etcdconfig.Password(etcdPassword))
	}
	etcd, err := etcd.New(endpoint, dialTime, etcdOpts...)
	if err != nil {
		return err
	}
	client := bus.NewClient(jobGroup, myIP, etcd)
	return client.Bootstrap()
}
