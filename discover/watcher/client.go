package watcher

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/hackbeex/configcenter/discover/client"
	"github.com/hackbeex/configcenter/discover/store"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/pkg/errors"
	"strings"
)

type ClientWatcher struct {
	table     *client.Table
	WatchChan clientv3.WatchChan
	ctx       context.Context
	store     *store.Store
	prefix    string
	attrs     []string
}

func NewClientWatcher(table *client.Table) *ClientWatcher {
	ctx := context.Background()
	sto := table.GetStore()
	ep := &ClientWatcher{
		table:  table,
		store:  sto,
		ctx:    ctx,
		prefix: client.KeyClientInstantPrefix,
		attrs:  []string{client.KeyClientAttrHost, client.KeyClientAttrPost, client.KeyClientAttrCluster, client.KeyClientAttrStatus, client.KeyClientAttrEnv},
	}
	ep.WatchChan = sto.Watch(ctx, ep.prefix, clientv3.WithPrefix())
	return ep
}

func (c *ClientWatcher) Ctx() context.Context {
	return c.ctx
}

func (c *ClientWatcher) GetWatchChan() clientv3.WatchChan {
	return c.WatchChan
}

func (c *ClientWatcher) Refresh() {
	c.ctx = context.Background()
	c.WatchChan = c.store.Watch(c.ctx, c.prefix, clientv3.WithPrefix())
}

func (c *ClientWatcher) Put(kv *mvccpb.KeyValue, isCreate bool) error {
	appId, _, err := c.fromKey(kv.Key)
	if err != nil {
		return err
	}

	log.Debugf("PUT EVENT[client], key: %s, value: %s", string(kv.Key), string(kv.Value))

	if isCreate {
		clientKey := c.prefix + appId
		ok, err := c.validKV(clientKey, c.attrs)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}
	if err := c.table.RefreshClientById(client.AppIdKey(appId)); err != nil {
		return err
	}
	return nil
}

func (c *ClientWatcher) Delete(kv *mvccpb.KeyValue) error {
	appId, _, err := c.fromKey(kv.Key)
	if err != nil {
		return err
	}

	log.Debugf("DELETE EVENT[client], key: %s", string(kv.Key))

	if err := c.table.DeleteClient(client.AppIdKey(appId)); err != nil {
		return err
	}
	return nil
}

func (c *ClientWatcher) fromKey(key []byte) (string, string, error) {
	path := strings.TrimPrefix(string(key), c.prefix)
	tmp := strings.Split(path, "/")
	if len(tmp) < 2 {
		err := errors.Errorf("invalid client key: %s", string(key))
		log.Error(err)
		return "", "", err
	}
	return tmp[0], tmp[1], nil
}

func (c *ClientWatcher) validKV(prefix string, attrs []string) (bool, error) {
	resp, err := c.store.GetKeyValueWithPrefix(prefix)
	if err != nil {
		return false, err
	}
	existAttrs := map[string]bool{}
	for _, kv := range resp.Kvs {
		_, attr, err := c.fromKey(kv.Key)
		if err != nil {
			continue
		}
		existAttrs[attr] = true
	}
	for _, attr := range attrs {
		if !existAttrs[attr] {
			return false, nil
		}
	}
	return true, nil
}
