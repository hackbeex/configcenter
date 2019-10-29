package discover

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/hackbeex/configcenter/discover/client"
	"github.com/hackbeex/configcenter/discover/server"
	"github.com/hackbeex/configcenter/discover/store"
	"github.com/hackbeex/configcenter/local"
	"github.com/hackbeex/configcenter/util/log"
	"os"
	"time"
)

type Table struct {
	version string

	servers *server.Table
	clients *client.Table

	store *store.Store
}

func connectToEtcd() *clientv3.Client {
	config := local.Conf.Discover.Etcd

	cli, err := clientv3.New(
		clientv3.Config{
			Endpoints:            config.Endpoints,
			AutoSyncInterval:     time.Duration(config.AutoSyncInterval) * time.Second,
			DialTimeout:          time.Duration(config.DialTimeout) * time.Second,
			DialKeepAliveTime:    time.Duration(config.DialKeepAliveTime) * time.Second,
			DialKeepAliveTimeout: time.Duration(config.DialKeepAliveTimeout) * time.Second,
			Username:             config.Username,
			Password:             config.Password,
		},
	)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
	return cli
}

func initTable() *Table {
	sto := store.New(connectToEtcd())
	table := &Table{
		version: "1.0.0",
		store:   sto,
		servers: server.InitTable(sto),
		clients: client.InitTable(sto),
	}
	return table
}

func (t *Table) Version() string {
	return t.version
}

func (t *Table) GetStore() *store.Store {
	return t.store
}
