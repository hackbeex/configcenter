package client

import (
	"fmt"
	"github.com/hackbeex/configcenter/discover/com"
	"github.com/hackbeex/configcenter/discover/store"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/pkg/errors"
)

const (
	KeyClientAppIdPrefix   = "/config-client/app-id/"
	KeyClientInstantPrefix = "/config-client/instance/"
	KeyClientAttrCluster   = "cluster"
	KeyClientAttrHost      = "host"
	KeyClientAttrPost      = "post"
	KeyClientAttrEnv       = "env"
)

type Client struct {
	AppId   string
	Cluster string
	Host    string
	Port    int
	Env     com.EnvType
	Status  com.RunStatus
}

func (c *Client) Register(store *store.Store) error {
	if c.AppId == "" {
		err := errors.New("client appid require")
		log.Error(err)
		return err
	}

	prefix := KeyClientInstantPrefix + c.AppId + "/"
	kvs := map[string]string{
		KeyClientAppIdPrefix + c.AppId: c.AppId,
		prefix + KeyClientAttrHost:     c.Host,
		prefix + KeyClientAttrPost:     fmt.Sprintf("%d", c.Port),
		prefix + KeyClientAttrCluster:  c.Cluster,
		prefix + KeyClientAttrEnv:      string(c.Env),
	}

	if err := store.PutKeyValues(kvs); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
