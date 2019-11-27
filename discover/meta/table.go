package meta

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/hackbeex/configcenter/discover/client"
	"github.com/hackbeex/configcenter/discover/server"
	"github.com/hackbeex/configcenter/discover/store"
	"github.com/hackbeex/configcenter/discover/watcher"
	"github.com/hackbeex/configcenter/local"
	"github.com/hackbeex/configcenter/util/log"
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

	clt, err := clientv3.New(
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
		log.Panic(err)
	}
	return clt
}

var tab *Table

func InitTable() {
	sto := store.New(connectToEtcd())
	clients := client.InitTable(sto)
	servers := server.InitTable(sto)
	tab = &Table{
		version: "1.0.0",
		store:   sto,
		servers: servers,
		clients: clients,
	}

	go runLoop()
	go tab.watch(watcher.NewServerWatcher(servers))
}

func GetTable() *Table {
	return tab
}

func GetStore() *store.Store {
	return tab.GetStore()
}

func (t *Table) Version() string {
	return t.version
}

func (t *Table) GetStore() *store.Store {
	return t.store
}

func (t *Table) Servers() *server.Table {
	return t.servers
}

func (t *Table) Clients() *client.Table {
	return t.clients
}
