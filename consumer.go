package owo

import (
	"errors"
	"fmt"
	"github.com/wuciyou/dogo"
	"net/rpc"
	"sync"
)

type consumer struct {
	serviceMethodMaps map[string][]*providerNode
	providerMaps      map[string]*providerNode
	rwm               *sync.RWMutex
}

func initConsumer() *consumer {
	c := &consumer{}
	c.rwm = &sync.RWMutex{}
	c.providerMaps = make(map[string]*providerNode)
	c.serviceMethodMaps = make(map[string][]*providerNode)
	return c
}

func (this *consumer) deleteProvider(name string) {
	this.rwm.Lock()
	defer this.rwm.Unlock()
	if _, exsits := this.providerMaps[name]; exsits {
		dogo.Dglog.Debugf("deleteProvider(name:%s)", name)
		delete(this.providerMaps, name)
	}
}

func (this *consumer) addProvider(name string, pNode *providerNode) {
	this.rwm.Lock()
	defer this.rwm.Unlock()

	if _, exsits := this.providerMaps[name]; !exsits {
		dogo.Dglog.Debugf("addProvider(name:%s,pHode:%+v)", name, pNode)
		this.providerMaps[name] = pNode
	}
}

func (this *consumer) call(serviceMethod string, args interface{}, reply interface{}) error {

	for realServiceName, pNode := range this.providerMaps {

		dogo.Dglog.Debugf("serverMethod:%s, pNode.Name:%s", serviceMethod, pNode.Name)
		if pNode.Name == serviceMethod {
			client, err := rpc.DialHTTP("tcp", pNode.Addr)
			if err != nil {
				dogo.Dglog.Errorf("Can't connect server method '%s' error:%+v \n", realServiceName, err)
				return err
			}
			err = client.Call(serviceMethod, args, reply)

			if err != nil {
				dogo.Dglog.Errorf("Can't call server method '%s' error:%+v \n", realServiceName, err)
				return err
			}

			return nil
		}
	}
	return errors.New(fmt.Sprintf("Not find serverMedthod:'%s'", serviceMethod))
}
