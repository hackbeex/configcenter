package client

import (
	"bytes"
	"github.com/hackbeex/configcenter/discover/store"
	"github.com/hackbeex/configcenter/util/log"
	"os"
	"strconv"
	"sync"
)

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

func InitTable(store *store.Store) *Table {
	resp, err := store.GetKeyValueWithPrefix(KeyConfigClientInstantPrefix)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
	clients := NewTable()
	for _, kv := range resp.Kvs {
		key := bytes.TrimPrefix(kv.Key, []byte(KeyConfigClientInstantPrefix))
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
		case KeyConfigClientAttrCluster:
			instance.Cluster = attr
		case KeyConfigClientAttrHost:
			instance.Host = attr
		case KeyConfigClientAttrPost:
			instance.Port, _ = strconv.Atoi(attr)
		default:
			log.Warn("invalid attr in client: ", attr)
		}
	}
	return clients
}
