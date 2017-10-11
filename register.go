package owo

import (
	"encoding/json"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/wuciyou/dogo"
	"time"
)

var rootPath = "/owo_data"

type providerEvent int

const (
	PROVIDER_DELETE providerEvent = iota
	PROVIDER_ADD
)

type register struct {
	addr     string
	rootPath string
	zkConn   *zk.Conn
}

func initRegister(addr string) *register {
	var err error
	r := &register{rootPath: rootPath}
	r.zkConn, _, err = zk.Connect([]string{addr}, time.Second*10)
	if err != nil {
		dogo.Dglog.Error(err)
	}

	exists, _, err := r.zkConn.Exists(r.rootPath)

	if err != nil {
		dogo.Dglog.Error(err)
	}

	if !exists {
		if _, err := r.zkConn.Create(r.rootPath, []byte{}, 0, zk.WorldACL(zk.PermAll)); err != nil {
			dogo.Dglog.Errorf("Create path:'%s' returned error:%+v ", r.rootPath, err)
		}
	}

	return r
}

func (this *register) listenEvent(callback func([]byte, string, providerEvent), isInit bool) {
	var oldChilds = make(map[string][]byte)
	for {
		var curChilds = make(map[string][]byte)
		var allChilds = make(map[string][]byte)
		childs, _, event, err := this.zkConn.ChildrenW(this.rootPath)
		for _, childName := range childs {
			childData, _, childErr := this.zkConn.Get(this.rootPath + "/" + childName)
			if childErr != nil {
				dogo.Dglog.Warningf("Get path'%s' returned error:%+v\n", childName, childErr)
			} else {
				curChilds[childName] = childData
				allChilds[childName] = childData
			}
		}
		if len(oldChilds) > 0 {
			for childName, childData := range oldChilds {
				if _, exsits := curChilds[childName]; !exsits {
					callback(childData, childName, PROVIDER_DELETE)
				} else {
					delete(curChilds, childName)
				}
			}
		}
		for childName, childData := range curChilds {
			callback(childData, childName, PROVIDER_ADD)
		}

		if isInit {
			break
		}
		e := <-event

		dogo.Dglog.Infof("event:%+v, err:%+v \n", e, err)

		oldChilds = allChilds
	}
}

func (this *register) push(pNode *providerNode) {

	nodePath := fmt.Sprintf("%s/%s_%s", this.rootPath, pNode.Name, pNode.Addr)
	data, _ := json.Marshal(pNode)
	if _, err := this.zkConn.CreateProtectedEphemeralSequential(nodePath, data, zk.WorldACL(zk.PermAll)); err != nil {
		dogo.Dglog.Errorf("Create path:'%s' returned error:%+v ", nodePath, err)
	}
}
