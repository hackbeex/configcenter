package core

import (
	"github.com/hackbeex/configcenter/util/com"
	"sync"
)

type ChangeConfig struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Instance struct {
	Id      string
	AppId   string
	Cluster string
	Host    string
	Port    int
	Status  com.RunStatus
	Life    int

	ChChange chan bool
	//ChReady bool
	//ChData  chan map[com.OpType][]ChangeConfig
}

type InstanceTable struct {
	table sync.Map
}

func NewInstanceTable() *InstanceTable {
	return &InstanceTable{
		table: sync.Map{},
	}
}

func (t *InstanceTable) Load(instanceId string) (*Instance, bool) {
	val, ok := t.table.Load(instanceId)
	return val.(*Instance), ok
}

func (t *InstanceTable) Store(instanceId string, val *Instance) {
	t.table.Store(instanceId, val)
}

func (t *InstanceTable) Delete(instanceId string) {
	t.table.Delete(instanceId)
}

func (t *InstanceTable) Range(f func(instanceId string, val *Instance) bool) {
	t.table.Range(func(k, v interface{}) bool {
		return f(k.(string), v.(*Instance))
	})
}

func (t *InstanceTable) LoadOrStore(instanceId string, val *Instance) (*Instance, bool) {
	res, loaded := t.table.LoadOrStore(instanceId, val)
	return res.(*Instance), loaded
}
