package client

import (
	"fmt"
	"github.com/hackbeex/configcenter/util"
	"github.com/hackbeex/configcenter/util/com"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/pkg/errors"
	"time"
)

type Client struct {
	host    string
	port    int
	cluster string
	env     com.EnvType

	discover discoverInfo
	server   serverInfo

	watchConfigInterval time.Duration
}

type discoverInfo struct {
	Host string
	Port int
}

type serverInfo struct {
	Id     string        `json:"id"`
	Host   string        `json:"host"`
	Port   int           `json:"port"`
	Env    com.EnvType   `json:"env"`
	Status com.RunStatus `json:"status"`
}

type Config struct {
	ClientHost    string
	ClientPort    int
	ClientCluster string
	ClientEnv     com.EnvType
	DiscoverHost  string
	DiscoverPort  int
}

func New(cf *Config) *Client {
	return &Client{
		host: cf.ClientHost,
		port: cf.ClientPort,
		env:  cf.ClientEnv,
		discover: discoverInfo{
			Host: cf.DiscoverHost,
			Port: cf.DiscoverPort,
		},
	}
}

func (c *Client) Register() error {
	svr, err := c.fetchMatchedServer()
	if err != nil {
		return err
	}
	c.server = svr

	go c.watchServer()
	go c.watchConfig()

	return nil
}

func (c *Client) watchServer() {
	defer func() {
		if err := recover(); err != nil {
			log.Warn("watch server recover: ", err)
			c.watchServer()
		}
	}()

	for {
		time.Sleep(time.Minute * 5)

		svr, err := c.fetchMatchedServer()
		if err != nil {
			log.Error(err)
			continue
		}
		if svr == c.server {
			continue
		}
		c.server = svr
		log.Info("matched server change:", svr)
	}
}

func (c *Client) fetchMatchedServer() (serverInfo, error) {
	var svr serverInfo
	url := fmt.Sprintf("%s:%d/api/v1/discover/server/fetch", c.discover.Host, c.discover.Port)
	res, err := util.HttpPostJson(url, nil)
	if err != nil {
		log.Error(err)
		return svr, err
	}

	var listResp struct {
		List []serverInfo `json:"list"`
	}
	err = util.HttpParseResponseToJson(res, &listResp)
	if err != nil {
		log.Error(err)
		return svr, err
	}

	for _, item := range listResp.List {
		if item.Env == c.env && item.Status == com.OnlineStatus {
			svr = item
			break
		}
	}
	if svr.Id == "" {
		err := errors.New("no online server find")
		log.Warn(err)
		return svr, err
	}
	return svr, nil
}

func (c *Client) fetchConfigChange() (config, error) {
	if c.server.Status != com.OnlineStatus {
		log.Warn("server is not online, can not watch config")
		return nil
	}
	var svr serverInfo
	url := fmt.Sprintf("%s:%d/api/v1/config/watch", c.server.Host, c.server.Port)
	res, err := util.HttpPostJson(url, nil)
	if err != nil {
		log.Error(err)
		return svr, err
	}

	var watchResp struct {
	}
	err = util.HttpParseResponseToJson(res, &watchResp)
	if err != nil {
		log.Error(err)
		return svr, err
	}

	//todo

	return svr, nil
}

func (c *Client) watchConfig() {
	defer func() {
		if err := recover(); err != nil {
			log.Warn("watch config recover: ", err)
			c.watchConfig()
		}
	}()

	for {
		cf, err := c.fetchConfigChange()
		if err != nil {
			log.Error(err)
			if c.watchConfigInterval >= time.Minute*5 {
				c.watchConfigInterval = time.Minute * 5
			} else if c.watchConfigInterval > 0 {
				c.watchConfigInterval = c.watchConfigInterval * 2
			} else {
				c.watchConfigInterval = time.Second
			}
			time.Sleep(c.watchConfigInterval)
			continue
		}

		c.watchConfigInterval = 0

		//todo do config something
		log.Info(cf)
	}
}
