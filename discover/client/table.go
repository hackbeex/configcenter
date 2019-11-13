package client

import (
	"bytes"
	"github.com/hackbeex/configcenter/discover/store"
	"github.com/hackbeex/configcenter/util/com"
	"github.com/hackbeex/configcenter/util/log"
	"strconv"
	"sync"
)

type AppIdKey string

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

func InitTable(store *store.Store) *Table {
	resp, err := store.GetKeyValueWithPrefix(KeyClientInstantPrefix)
	if err != nil {
		log.Panic(err)
	}
	clients := NewTable(store)
	for _, kv := range resp.Kvs {
		key := bytes.TrimPrefix(kv.Key, []byte(KeyClientInstantPrefix))
		segments := bytes.Split(key, []byte("/"))
		if len(segments) != 2 {
			log.Warnf("invalid config client definition: %s", string(kv.Key))
			continue
		}
		id := AppIdKey(segments[0])
		attr := string(segments[1])

		instance, ok := clients.Load(id)
		if !ok {
			instance = &Client{
				AppId: string(id),
			}
			clients.Store(id, instance)
		}

		switch attr {
		case KeyClientAttrCluster:
			instance.Cluster = attr
		case KeyClientAttrHost:
			instance.Host = attr
		case KeyClientAttrPost:
			instance.Port, _ = strconv.Atoi(attr)
		case KeyClientAttrEnv:
			instance.Env = com.EnvType(attr)
		default:
			log.Warn("invalid attr in client: ", attr)
		}
	}
	return clients
}

type ClientInfo struct {
	AppId   string        `json:"app_id"`
	Cluster string        `json:"cluster"`
	Host    string        `json:"host"`
	Port    int           `json:"port"`
	Env     com.EnvType   `json:"env"`
	Status  com.RunStatus `json:"status"`
}

func (t *Table) FetchClientList() ([]ClientInfo, error) {
	var list = make([]ClientInfo, 0)
	t.Range(func(key AppIdKey, val *Client) bool {
		list = append(list, ClientInfo{
			AppId:   val.AppId,
			Cluster: val.Cluster,
			Host:    val.Host,
			Port:    val.Port,
			Env:     val.Env,
			Status:  val.Status,
		})
		return true
	})
	return list, nil
}
