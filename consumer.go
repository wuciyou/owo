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
	if providerNode, exsits := this.providerMaps[name]; exsits {
		methodName := providerNode.Name
		for index, serviceMethodNode := range this.serviceMethodMaps[methodName] {
			if serviceMethodNode.Id == name {
				this.serviceMethodMaps[methodName] = append(this.serviceMethodMaps[methodName][:index], this.serviceMethodMaps[methodName][index+1:]...)
			}
		}
		if len(this.serviceMethodMaps[methodName]) == 0 {
			delete(this.serviceMethodMaps, methodName)
		}
		delete(this.providerMaps, name)
		dogo.Dglog.Debugf("deleteProvider(name:%s),serviceMethodMaps:%+v \n", name, this.serviceMethodMaps[methodName])
	}
}

func (this *consumer) addProvider(name string, pNode *providerNode) {
	this.rwm.Lock()
	defer this.rwm.Unlock()

	if _, exsits := this.providerMaps[name]; !exsits {
		this.providerMaps[name] = pNode
		this.serviceMethodMaps[pNode.Name] = append(this.serviceMethodMaps[pNode.Name], pNode)
		dogo.Dglog.Debugf("addProvider(name:%s,pHode:%+v), serviceMethodMaps:%+v\n", name, pNode, this.serviceMethodMaps[pNode.Name])
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
