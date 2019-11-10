package watcher

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/hackbeex/configcenter/discover/server"
	"github.com/hackbeex/configcenter/discover/store"
	"github.com/hackbeex/configcenter/util/log"
)

type ServerWatcher struct {
	table     *server.Table
	WatchChan clientv3.WatchChan
	ctx       context.Context
	store     *store.Store
	prefix    string
	attrs     []string
}

func NewServerWatcher(table *server.Table) *ServerWatcher {
	ctx := context.Background()
	sto := table.GetStore()
	ep := &ServerWatcher{
		table:  table,
		store:  sto,
		ctx:    ctx,
		prefix: server.KeyServerInstantPrefix,
		attrs:  []string{server.KeyServerAttrHost, server.KeyServerAttrPost, server.KeyServerAttrEnv, server.KeyServerAttrStatus},
	}
	ep.WatchChan = sto.Watch(ctx, ep.prefix, clientv3.WithPrefix())
	return ep
}

func (c *ServerWatcher) Ctx() context.Context {
	return c.ctx
}

func (c *ServerWatcher) GetWatchChan() clientv3.WatchChan {
	return c.WatchChan
}

func (c *ServerWatcher) Refresh() {
	c.ctx = context.Background()
	c.WatchChan = c.store.Watch(c.ctx, c.prefix, clientv3.WithPrefix())
}

func (c *ServerWatcher) Put(kv *mvccpb.KeyValue, isCreate bool) error {
	appId, _, err := c.store.FromKeyToValue(c.prefix, kv.Key)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("PUT EVENT[server], key: %s, value: %s", string(kv.Key), string(kv.Value))

	if isCreate {
		serverKey := c.prefix + appId
		ok, err := c.store.IsValidKV(serverKey, c.attrs)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}
	if err := c.table.RefreshServerById(server.IdKey(appId)); err != nil {
		return err
	}
	return nil
}

func (c *ServerWatcher) Delete(kv *mvccpb.KeyValue) error {
	appId, _, err := c.store.FromKeyToValue(c.prefix, kv.Key)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("DELETE EVENT[server], key: %s", string(kv.Key))

	if err := c.table.DeleteServer(server.IdKey(appId)); err != nil {
		return err
	}
	return nil
}
