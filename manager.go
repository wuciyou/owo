package owo

import (
	"encoding/json"
	"flag"
	"github.com/wuciyou/dogo"
	"net"
	"net/http"
	"net/rpc"
	"strings"
)

const (
	DEFAULT_LISTEN_ADDR = ":0"
)

type owoApp struct {
	listenAddr string
	listen     net.Listener
	register   *register
	consumer   *consumer
	provider   *provider
}

var app = initOwo()

func initOwo() *owoApp {
	var err error
	listenAddr := flag.String("listen_addr", DEFAULT_LISTEN_ADDR, "服务监听地址，格式为[ip:port]")
	registerAddr := flag.String("register_addr", "127.0.0.1:2181", "注册中心地址 ，格式为[ip:port]")
	flag.Parse()
	app := &owoApp{
		listenAddr: *listenAddr,
	}

	app.listen, err = net.Listen("tcp", app.ListenAddr())
	dogo.Dglog.Infof("listen_addr:%s, register_addr:%s", app.listen.Addr().String(), *registerAddr)
	if err != nil {
		panic(err)
	}

	app.register = initRegister(*registerAddr)
	app.consumer = initConsumer()
	app.provider = initProvider(app.listen.Addr().String())
	app.register.listenEvent(app.refresh, true)
	go app.register.listenEvent(app.refresh, false)
	return app
}

func (this *owoApp) refresh(nodeData []byte, name string, eventType providerEvent) {
	dogo.Dglog.Infof("update %s, name:%s ,evenType:%d \n", string(nodeData), name, eventType)
	pNode := &providerNode{}
	err := json.Unmarshal(nodeData, pNode)
	if err != nil {
		dogo.Dglog.Errorf("Can't unmarshal data:'%s' to %+v \n", string(nodeData), pNode)
	}
	switch eventType {
	case PROVIDER_DELETE:
		this.consumer.deleteProvider(name)
	case PROVIDER_ADD:
		pNode.Id = name
		this.consumer.addProvider(name, pNode)
	}

}

func (this *owoApp) ListenAddr() string {
	this.listenAddr = strings.TrimSpace(this.listenAddr)
	if this.listenAddr == "" {
		return DEFAULT_LISTEN_ADDR
	}
	return this.listenAddr
}

func Call(serviceMethod string, args interface{}, reply interface{}) error {
	return app.consumer.call(serviceMethod, args, reply)
}

func Register(rcvr interface{}, name ...string) {
	providerNode, err := app.provider.Register(rcvr, name...)
	if err != nil {
		panic(err)
	}
	for _, node := range providerNode {
		app.register.push(node)
	}
}

func Run() {
	rpc.HandleHTTP()
	http.Serve(app.listen, nil)
}
