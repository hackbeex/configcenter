package server

import "sync"

type IdKey string

type Table struct {
	table sync.Map
}

func NewTable() *Table {
	return &Table{
		table: sync.Map{},
	}
}

func (t *Table) Load(key IdKey) (*Server, bool) {
	val, ok := t.table.Load(key)
	return val.(*Server), ok
}

func (t *Table) Store(key IdKey, val *Server) {
	t.table.Store(key, val)
}

func (t *Table) Delete(key IdKey) {
	t.table.Delete(key)
}

func (t *Table) Range(f func(key IdKey, val *Server) bool) {
	t.table.Range(func(k, v interface{}) bool {
		return f(k.(IdKey), v.(*Server))
	})
}

func (t *Table) LoadOrStore(key IdKey, val *Server) (*Server, bool) {
	res, loaded := t.table.LoadOrStore(key, val)
	return res.(*Server), loaded
}
