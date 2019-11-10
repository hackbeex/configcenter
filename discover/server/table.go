package server

import (
	"bytes"
	"github.com/hackbeex/configcenter/discover/com"
	"github.com/hackbeex/configcenter/discover/store"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/pkg/errors"
	"strconv"
	"sync"
)

type IdKey string

type Table struct {
	table sync.Map
	store *store.Store
}

func NewTable(store *store.Store) *Table {
	return &Table{
		table: sync.Map{},
		store: store,
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
	resp, err := store.GetKeyValueWithPrefix(KeyServerInstantPrefix)
	if err != nil {
		log.Panic(err)
	}
	servers := NewTable(store)
	for _, kv := range resp.Kvs {
		key := bytes.TrimPrefix(kv.Key, []byte(KeyServerInstantPrefix))
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
		case KeyServerAttrEnv:
			instance.Env = com.EnvType(attr)
		case KeyServerAttrHost:
			instance.Host = attr
		case KeyServerAttrPost:
			instance.Port, _ = strconv.Atoi(attr)
		case KeyServerAttrStatus:
			instance.Status = com.RunStatus(attr)
		default:
			log.Warn("invalid attr in server: ", attr)
		}
	}
	return servers
}

func (t *Table) GetStore() *store.Store {
	return t.store
}

func (t *Table) RefreshServerById(key IdKey) error {
	resp, err := t.store.GetKeyValueWithPrefix(KeyServerInstantPrefix + string(key))
	if err != nil {
		log.Error(err)
		return err
	}

	svr, ok := t.Load(key)
	if !ok {
		svr = &Server{
			Id: string(key),
		}
	}

	for _, kv := range resp.Kvs {
		keyStr := string(bytes.TrimPrefix(kv.Key, []byte(key)))
		switch keyStr {
		case KeyServerAttrHost:
			svr.Host = keyStr
		case KeyServerAttrPost:
			svr.Port, _ = strconv.Atoi(keyStr)
		case KeyServerAttrStatus:
			svr.Status = com.RunStatus(keyStr)
		case KeyServerAttrEnv:
			svr.Env = com.EnvType(keyStr)
		default:
			err := errors.Errorf("unsupported server attr %s", keyStr)
			log.Error(err)
			return err
		}
	}
	t.Store(key, svr)

	//todo notify clients which uses this server to update server list

	return nil
}

func (t *Table) DeleteServer(key IdKey) error {
	_, ok := t.Load(key)
	if !ok {
		return nil
	}

	//todo notify clients which uses this server to update server list

	t.Delete(key)

	return nil
}

func (t *Table) UpdateStatus(key IdKey, status com.RunStatus) error {
	server, ok := t.Load(key)
	if !ok {
		err := errors.Errorf("server not exist: %s", key)
		log.Error(err)
		return err
	}
	if server.Status == status {
		return nil
	}

	k := KeyServerInstantPrefix + key + "/" + KeyServerAttrStatus
	_, err := t.store.PutKeyValue(string(k), string(com.OnlineStatus))
	if err != nil {
		log.Error(err)
		return err
	}

	server.Status = status
	return nil
}

func (t *Table) FetchServerList() ([]Server, error) {
	var list = make([]Server, 0)
	t.Range(func(key IdKey, val *Server) bool {
		list = append(list, *val)
		return true
	})
	return list, nil
}
