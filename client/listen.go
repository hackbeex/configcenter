package client

import (
	"github.com/hackbeex/configcenter/util/com"
	"sync"
)

type CallbackParam struct {
	Key    string
	NewVal string
	OldVal string
	OpType com.OpType
}
type ListenCallback func(param *CallbackParam)

type Listen struct {
	key       string
	callbacks []ListenCallback
}

type ListenTable struct {
	table sync.Map
}

func NewListenTable() *ListenTable {
	return &ListenTable{
		table: sync.Map{},
	}
}

func (t *ListenTable) Load(key string) (*Listen, bool) {
	val, ok := t.table.Load(key)
	return val.(*Listen), ok
}

func (t *ListenTable) Store(key string, val *Listen) {
	t.table.Store(key, val)
}

func (t *ListenTable) Delete(key string) {
	t.table.Delete(key)
}

func (t *ListenTable) Range(f func(key string, val *Listen) bool) {
	t.table.Range(func(k, v interface{}) bool {
		return f(k.(string), v.(*Listen))
	})
}

func (t *ListenTable) LoadOrStore(key string, val *Listen) (*Listen, bool) {
	res, loaded := t.table.LoadOrStore(key, val)
	return res.(*Listen), loaded
}

func (t *ListenTable) AddCallback(key string, callback ListenCallback) {
	val, ok := t.Load(key)
	if !ok {
		t.Store(key, &Listen{
			key:       key,
			callbacks: []ListenCallback{callback},
		})
		return
	}
	val.callbacks = append(val.callbacks, callback)
	t.Store(key, val)
}

func (t *ListenTable) Call(key string, param *CallbackParam, global bool) {
	if listen, ok := t.Load(param.Key); ok {
		for _, callback := range listen.callbacks {
			callback(param)
		}
	}
	if global && key != "" {
		if listen, ok := t.Load(""); ok {
			for _, callback := range listen.callbacks {
				callback(param)
			}
		}
	}
}
