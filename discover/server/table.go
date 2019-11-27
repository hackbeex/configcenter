package server

import (
	"bytes"
	"github.com/hackbeex/configcenter/discover/store"
	"github.com/hackbeex/configcenter/util/com"
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
	if !ok {
		return nil, ok
	}
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
	log.Debug("servers: ", servers)
	return servers
}

func (t *Table) GetStore() *store.Store {
	return t.store
}

func (t *Table) RefreshServerById(key IdKey) error {
	fullKey := KeyServerInstantPrefix + string(key) + "/"
	resp, err := t.store.GetKeyValueWithPrefix(fullKey)
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
		keyStr := string(bytes.TrimPrefix(kv.Key, []byte(fullKey)))
		switch keyStr {
		case KeyServerAttrHost:
			svr.Host = string(kv.Value)
		case KeyServerAttrPost:
			svr.Port, _ = strconv.Atoi(string(kv.Value))
		case KeyServerAttrStatus:
			svr.Status = com.RunStatus(string(kv.Value))
		case KeyServerAttrEnv:
			svr.Env = com.EnvType(string(kv.Value))
		default:
			err := errors.Errorf("unsupported server attr %s", keyStr)
			log.Error(err)
			return err
		}
	}
	//log.Debug("store server: ", svr)
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
	if status != com.OnlineStatus && status != com.OfflineStatus && status != com.BreakStatus {
		err := errors.Errorf("status is not support: %s", status)
		log.Error(err)
		return err
	}

	server, ok := t.Load(key)
	if !ok {
		err := errors.Errorf("server not exist: %s", key)
		log.Error(err)
		return err
	}
	if status == com.OnlineStatus {
		server.Life = serverMaxLife
	} else {
		log.Debug("server status set: ", status)
		server.Life = 0
	}

	if server.Status == status {
		t.Store(key, server)
		return nil
	}

	k := KeyServerInstantPrefix + key + "/" + KeyServerAttrStatus
	_, err := t.store.PutKeyValue(string(k), string(status))
	if err != nil {
		log.Error(err)
		return err
	}

	server.Status = status
	t.Store(key, server)
	return nil
}

type ServerInfo struct {
	Id     string        `json:"id"`
	Host   string        `json:"host"`
	Port   int           `json:"port"`
	Env    com.EnvType   `json:"env"`
	Status com.RunStatus `json:"status"`
}

func (t *Table) FetchServerList() ([]ServerInfo, error) {
	var list = make([]ServerInfo, 0)
	t.Range(func(key IdKey, val *Server) bool {
		list = append(list, ServerInfo{
			Id:     val.Id,
			Host:   val.Host,
			Port:   val.Port,
			Env:    val.Env,
			Status: val.Status,
		})
		return true
	})
	return list, nil
}
