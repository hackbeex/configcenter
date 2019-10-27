package discover

import (
	"bytes"
	"github.com/coreos/etcd/clientv3"
	"github.com/hackbeex/configcenter/client"
	"github.com/hackbeex/configcenter/local"
	"github.com/hackbeex/configcenter/server"
	"github.com/hackbeex/configcenter/util/log"
	"os"
	"strconv"
	"time"
)

type Table struct {
	Version string
	Servers *server.Table
	Clients *client.Table

	etcd *clientv3.Client
}

const (
	Slash = "/"

	KeyConfigClientAppIdPrefix   = "/config-client/app-id/"
	KeyConfigClientInstantPrefix = "/config-client/instance/"
	KeyConfigClientAttrCluster   = "cluster"
	KeyConfigClientAttrHost      = "host"
	KeyConfigClientAttrPost      = "post"

	KeyConfigServerIdPrefix      = "/config-server/id/"
	KeyConfigServerInstantPrefix = "/config-server/instance/"
	KeyConfigServerAttrEnv       = "env"
	KeyConfigServerAttrHost      = "host"
	KeyConfigServerAttrPost      = "post"
)

func ConnectToEtcd() *clientv3.Client {
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
	table := &Table{
		Version: "1.0.0",
		etcd:    ConnectToEtcd(),
	}

	table.initConfigClients()
	table.initConfigServers()

	return table
}

func (t *Table) initConfigClients() {
	resp, err := t.getKeyValueWithPrefix(KeyConfigClientInstantPrefix)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
	t.Clients = client.NewTable()
	for _, kv := range resp.Kvs {
		key := bytes.TrimPrefix(kv.Key, []byte(KeyConfigClientInstantPrefix))
		segments := bytes.Split(key, []byte(Slash))
		if len(segments) != 2 {
			log.Warnf("invalid config client definition: %s", string(kv.Key))
			continue
		}
		id := client.AppIdKey(segments[0])
		attr := string(segments[1])

		instance, ok := t.Clients.Load(id)
		if !ok {
			instance = &client.Client{
				AppId: string(id),
			}
			t.Clients.Store(id, instance)
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
}

func (t *Table) initConfigServers() {
	resp, err := t.getKeyValueWithPrefix(KeyConfigServerInstantPrefix)
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
	t.Servers = server.NewTable()
	for _, kv := range resp.Kvs {
		key := bytes.TrimPrefix(kv.Key, []byte(KeyConfigServerInstantPrefix))
		segments := bytes.Split(key, []byte(Slash))
		if len(segments) != 2 {
			log.Warnf("invalid config server definition: %s", string(kv.Key))
			continue
		}
		id := server.IdKey(segments[0])
		attr := string(segments[1])

		instance, ok := t.Servers.Load(id)
		if !ok {
			instance = &server.Server{
				Id: string(id),
			}
			t.Servers.Store(id, instance)
		}

		switch attr {
		case KeyConfigServerAttrEnv:
			instance.Env = server.EnvType(attr)
		case KeyConfigServerAttrHost:
			instance.Host = attr
		case KeyConfigServerAttrPost:
			instance.Port, _ = strconv.Atoi(attr)
		default:
			log.Warn("invalid attr in server: ", attr)
		}
	}
}
