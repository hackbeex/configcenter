package watcher

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

type Watcher interface {
	Put(kv *mvccpb.KeyValue, isCreate bool) error
	Delete(kv *mvccpb.KeyValue) error
	GetWatchChan() clientv3.WatchChan
	Ctx() context.Context
	Refresh()
}
