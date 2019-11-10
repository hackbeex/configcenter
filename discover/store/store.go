package store

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type Store struct {
	client *clientv3.Client
}

func New(client *clientv3.Client) *Store {
	return &Store{
		client: client,
	}
}

func (s *Store) GetKeyValue(key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := s.client.Get(ctx, key, opts...)
	cancel()
	return resp, err
}

func (s *Store) GetKeyValueWithPrefix(key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	opts = append(opts, clientv3.WithPrefix())
	return s.GetKeyValue(key, opts...)
}

func (s *Store) PutKeyValue(key, value string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := s.client.Put(ctx, key, value, opts...)
	cancel()
	return resp, err
}

func (s *Store) PutKeyValues(kvs map[string]string, opts ...clientv3.OpOption) error {
	if len(kvs) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	for k, v := range kvs {
		_, err := s.client.Put(ctx, k, v, opts...)
		if err != nil {
			cancel()
			return err
		}
	}
	cancel()
	return nil
}

func (s *Store) DeleteKeyValue(key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	resp, err := s.client.Delete(ctx, key, opts...)
	cancel()
	return resp, err
}

func (s *Store) deleteKeyValueWithPrefix(key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	opts = append(opts, clientv3.WithPrefix())
	return s.DeleteKeyValue(key, opts...)
}

func (s *Store) DeleteKeyValues(ks []string, opts ...clientv3.OpOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	for _, v := range ks {
		_, err := s.client.Delete(ctx, v, opts...)
		if err != nil {
			cancel()
			return err
		}
	}
	cancel()
	return nil
}

func (s *Store) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	return s.client.Watch(ctx, key, opts...)
}

func (s *Store) FromKeyToValue(prefix string, key []byte) (string, string, error) {
	path := strings.TrimPrefix(string(key), prefix)
	tmp := strings.Split(path, "/")
	if len(tmp) < 2 {
		err := errors.Errorf("invalid server key: %s", string(key))
		return "", "", err
	}
	return tmp[0], tmp[1], nil
}

func (s *Store) IsValidKV(prefixKey string, attrs []string) (bool, error) {
	resp, err := s.GetKeyValueWithPrefix(prefixKey)
	if err != nil {
		return false, err
	}
	existAttrs := map[string]bool{}
	for _, kv := range resp.Kvs {
		attr := strings.TrimPrefix(string(kv.Key), prefixKey)
		if attr == "" {
			log.Warnf("%s: attr is empty", string(kv.Key))
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
