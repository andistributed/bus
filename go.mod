module github.com/andistributed/bus

go 1.12

replace github.com/andistributed/etcd => ../etcd

require (
	github.com/admpub/log v0.3.1
	github.com/andistributed/etcd v0.1.0
	github.com/mattn/go-isatty v0.0.13 // indirect
	go.etcd.io/etcd/client/v3 v3.5.0-rc.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5 // indirect
	golang.org/x/sys v0.0.0-20210608053332-aa57babbf139 // indirect
	google.golang.org/genproto v0.0.0-20210608205507-b6d2f5bf0d7d // indirect
)
