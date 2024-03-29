package client

import (
	"sync"
)

type Item struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ConfigTable struct {
	table sync.Map
}

func NewConfigTable() *ConfigTable {
	return &ConfigTable{
		table: sync.Map{},
	}
}

func (t *ConfigTable) Load(key string) (*Item, bool) {
	val, ok := t.table.Load(key)
	if !ok {
		return nil, ok
	}
	return val.(*Item), ok
}

func (t *ConfigTable) Store(key string, val *Item) {
	t.table.Store(key, val)
}

func (t *ConfigTable) Delete(key string) {
	t.table.Delete(key)
}

func (t *ConfigTable) Range(f func(key string, val *Item) bool) {
	t.table.Range(func(k, v interface{}) bool {
		return f(k.(string), v.(*Item))
	})
}
