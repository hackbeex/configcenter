package client

import "sync"

type AppIdKey string

type Table struct {
	table sync.Map
}

func NewTable() *Table {
	return &Table{
		table: sync.Map{},
	}
}

func (t *Table) Load(key AppIdKey) (*Client, bool) {
	val, ok := t.table.Load(key)
	return val.(*Client), ok
}

func (t *Table) Store(key AppIdKey, val *Client) {
	t.table.Store(key, val)
}

func (t *Table) Delete(key AppIdKey) {
	t.table.Delete(key)
}

func (t *Table) Range(f func(key AppIdKey, val *Client) bool) {
	t.table.Range(func(k, v interface{}) bool {
		return f(k.(AppIdKey), v.(*Client))
	})
}

func (t *Table) LoadOrStore(key AppIdKey, val *Client) (*Client, bool) {
	res, loaded := t.table.LoadOrStore(key, val)
	return res.(*Client), loaded
}
