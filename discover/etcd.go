package discover

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func (t *Table) getKeyValue(key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := t.etcd.Get(ctx, key, opts...)
	cancel()
	return resp, err
}

func (t *Table) getKeyValueWithPrefix(key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	opts = append(opts, clientv3.WithPrefix())
	return t.getKeyValue(key, opts...)
}

func (t *Table) putKeyValue(key, value string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := t.etcd.Put(ctx, key, value, opts...)
	cancel()
	return resp, err
}

func (t *Table) putKeyValueWithPrefix(key, value string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	opts = append(opts, clientv3.WithPrefix())
	return t.putKeyValue(key, value, opts...)
}

func (t *Table) putKeyValues(kvs map[string]string, opts ...clientv3.OpOption) error {
	if len(kvs) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	for k, v := range kvs {
		_, err := t.etcd.Put(ctx, k, v, opts...)
		if err != nil {
			cancel()
			return err
		}
	}
	cancel()
	return nil
}

func (t *Table) deleteKeyValue(key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := t.etcd.Delete(ctx, key, opts...)
	cancel()
	return resp, err
}

func (t *Table) deleteKeyValueWithPrefix(key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	opts = append(opts, clientv3.WithPrefix())
	return t.deleteKeyValue(key, opts...)
}

func (t *Table) deleteKeyValues(ks []string, opts ...clientv3.OpOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	for _, v := range ks {
		_, err := t.etcd.Delete(ctx, v, opts...)
		if err != nil {
			cancel()
			return err
		}
	}
	cancel()
	return nil
}
