package owo

import (
	"encoding/json"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/wuciyou/dogo"
	"sync"
	"time"
)

var rootPath = "/owo_data"

type providerEvent int

const (
	PROVIDER_DELETE providerEvent = iota
	PROVIDER_ADD
)

type hooksMethod func(r *register)
type register struct {
	addr                     string
	rootPath                 string
	zkConn                   *zk.Conn
	cacheSelfProviderNodeMap map[string]*providerNode
	rwm                      *sync.RWMutex
	zkConnEvent              <-chan zk.Event
	hooksMethod              map[string][]hooksMethod
}

func initRegister(addr string) *register {
	var err error
	r := &register{rootPath: rootPath}
	r.cacheSelfProviderNodeMap = make(map[string]*providerNode)
	r.rwm = &sync.RWMutex{}
	r.zkConn, r.zkConnEvent, err = zk.Connect([]string{addr}, time.Second*10)
	r.hooksMethod = make(map[string][]hooksMethod)
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
	r.listenEvent(r.zkConnEvent)

	// 断线重新连接成功后，重新注册服务
	r.listenEventFunc(r.disRePush, zk.EventSession, zk.StateHasSession)

	return r
}

func (this *register) listenEvent(event <-chan zk.Event) {
	go func() {
		for {
			e := <-event
			for _, hookFunc := range this.hooksMethod[fmt.Sprintf("%d_%d", e.Type, e.State)] {
				hookFunc(this)
			}

			for _, hookFunc := range this.hooksMethod[fmt.Sprintf("%d_", e.Type)] {
				hookFunc(this)
			}
		}
	}()
}

func (this *register) listenEventFunc(hook hooksMethod, eventType zk.EventType, eventStat ...zk.State) {
	hooksKey := fmt.Sprintf("%d_", eventType)
	if len(eventStat) > 0 {
		hooksKey = fmt.Sprintf("%d_%d", eventType, eventStat[0])
	}
	this.rwm.Lock()
	this.hooksMethod[hooksKey] = append(this.hooksMethod[hooksKey], hook)
	this.rwm.Unlock()
}

func (this *register) listenNode(callback func([]byte, string, providerEvent), isInit bool) {
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

func (this *register) disRePush(r *register) {
	for path, providerNode := range this.cacheSelfProviderNodeMap {
		if isExists, _, err := this.zkConn.Exists(path); !isExists || err != nil {
			dogo.Dglog.Warningf("repush provider node old path:'%s', err:%+v", path, err)
			this.push(providerNode)
		}
	}
}
func (this *register) push(pNode *providerNode) {

	nodePath := fmt.Sprintf("%s/%s_%s", this.rootPath, pNode.Name, pNode.Addr)
	data, _ := json.Marshal(pNode)
	if path, err := this.zkConn.CreateProtectedEphemeralSequential(nodePath, data, zk.WorldACL(zk.PermAll)); err != nil {
		dogo.Dglog.Errorf("Create path:'%s' returned error:%+v ", nodePath, err)
	} else {
		this.rwm.Lock()
		pNode.Id = path
		this.cacheSelfProviderNodeMap[path] = pNode
		this.rwm.Unlock()
	}

}
