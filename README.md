# bus

> forest Go sdk

## 说明

[分布式调度平台](https://github.com/andistributed/forest)

## 快速开始

> go get github.com/andistributed/cat

### 定义一个任务

```go

type EchoJob struct {

}

func (*EchoJob) Execute(params string) (string, error) {
	time.Sleep(time.Second * 5)
	fmt.Println("参数:", params)
	return "ok", nil
}


```

### 配置客户端&启动

```go


etcd, _ := NewEtcd([]string{"127.0.0.1:2379"}, time.Second*10)
	forestClient := NewForestClient("trade", "127.0.0.1", etcd)

	forestClient.PushJob("com.busgo.cat.job.EchoJob",&EchoJob{})
	forestClient.Bootstrap()
	


```

### 控制台输出

```shell

=== RUN   TestForestClient_Bootstrap
2019/12/24 17:41:51 the forest client success registry to:/forest/client/trade/clients/127.0.0.1

```

## 联系方式

如有问题请联系 QQ:466862016 Email:466862016@qq.com 讨论QQ群:806735002 欢迎指点拍砖！