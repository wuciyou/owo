package owo

import (
	"errors"
	"fmt"
	"github.com/wuciyou/dogo"
	"math/rand"
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

	if pNode, err := this.doSelect(serviceMethod); err == nil {

		dogo.Dglog.Debugf("serverMethod:%s, pNode.Name:%s", serviceMethod, pNode.Name)

		client, err := rpc.DialHTTP("tcp", pNode.Addr)
		if err != nil {
			dogo.Dglog.Errorf("Can't connect server method '%s' error:%+v \n", pNode.Id, err)
			return err
		}
		err = client.Call(serviceMethod, args, reply)

		if err != nil {
			dogo.Dglog.Errorf("Can't call server method '%s' error:%+v \n", pNode.Id, err)
			return err
		}

	} else {
		return err
	}
	return nil
}

func (this *consumer) doSelect(name string) (*providerNode, error) {
	if providerNodes, exists := this.serviceMethodMaps[name]; exists {
		var totalWeight = 0
		var sameWeight = true
		for index, provider := range providerNodes {
			totalWeight += provider.Level
			if sameWeight && index > 0 && provider.Level != providerNodes[index-1].Level {
				sameWeight = false
			}
		}
		if totalWeight > 0 && !sameWeight {
			offset := rand.Intn(totalWeight)

			for _, provider := range providerNodes {
				offset -= provider.Level
				if offset < 0 {
					return provider, nil
				}
			}
		}
		if len(providerNodes) == 1 {
			return providerNodes[0], nil
		} else {
			return providerNodes[rand.Intn(len(providerNodes)-1)], nil
		}
	}

	errMesg := fmt.Sprintf("Can't find service method name:'%s'", name)
	dogo.Dglog.Errorf(errMesg)
	return nil, errors.New(errMesg)
}
