package server

import (
	"bytes"
	"github.com/hackbeex/configcenter/discover/store"
	"github.com/hackbeex/configcenter/util/log"
	"os"
	"strconv"
	"sync"
)

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

func InitTable(store *store.Store) *Table {
	resp, err := store.GetKeyValueWithPrefix(KeyConfigServerInstantPrefix)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
	servers := NewTable()
	for _, kv := range resp.Kvs {
		key := bytes.TrimPrefix(kv.Key, []byte(KeyConfigServerInstantPrefix))
		segments := bytes.Split(key, []byte("/"))
		if len(segments) != 2 {
			log.Warnf("invalid config server definition: %s", string(kv.Key))
			continue
		}
		id := IdKey(segments[0])
		attr := string(segments[1])

		instance, ok := servers.Load(id)
		if !ok {
			instance = &Server{
				Id: string(id),
			}
			servers.Store(id, instance)
		}

		switch attr {
		case KeyConfigServerAttrEnv:
			instance.Env = EnvType(attr)
		case KeyConfigServerAttrHost:
			instance.Host = attr
		case KeyConfigServerAttrPost:
			instance.Port, _ = strconv.Atoi(attr)
		default:
			log.Warn("invalid attr in server: ", attr)
		}
	}
	return servers
}
