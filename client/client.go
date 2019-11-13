package client

import (
	"encoding/json"
	"fmt"
	"github.com/hackbeex/configcenter/util"
	"github.com/hackbeex/configcenter/util/com"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Client struct {
	InstanceId string
	Host       string
	Port       int
	Cluster    string
	Env        com.EnvType

	discover discoverInfo
	server   serverInfo
	config   *ConfigTable

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
		Host:    cf.ClientHost,
		Port:    cf.ClientPort,
		Env:     cf.ClientEnv,
		Cluster: cf.ClientCluster,
		discover: discoverInfo{
			Host: cf.DiscoverHost,
			Port: cf.DiscoverPort,
		},
		config: NewConfigTable(),
	}
}

func (c *Client) Register() error {
	if err := c.initServer(); err != nil {
		return err
	}
	if err := c.initConfig(); err != nil {
		return err
	}

	go c.watchServer()
	go c.watchConfig()

	return nil
}

func (c *Client) initServer() error {
	svr, err := c.fetchMatchedServer()
	if err != nil {
		return err
	}
	c.server = svr
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
		if item.Env == c.Env && item.Status == com.OnlineStatus {
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

func (c *Client) WatchServerExit() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Info("config client exiting ...")
	if err := c.DoServerExit(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func (c *Client) DoServerExit() error {
	if c.InstanceId == "" {
		log.Warn("config client do not have instance id")
		return nil
	}
	data, _ := json.Marshal(map[string]interface{}{
		"instance_id": c.InstanceId,
	})
	url := fmt.Sprintf("%s:%d/api/v1/instance/exit", c.server.Host, c.server.Port)
	_, err := util.HttpPostJson(url, data)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
