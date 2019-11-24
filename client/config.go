package client

import (
	"fmt"
	"github.com/hackbeex/configcenter/util"
	"github.com/hackbeex/configcenter/util/com"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/hackbeex/configcenter/util/response"
	"time"
)

func (c *Client) initConfig() error {
	res, err := c.GetConfigList()
	if err != nil {
		log.Error(err)
		return err
	}

	for _, item := range res.List {
		c.config.Store(item.Key, &item)
	}

	return nil
}

type ConfigListResp struct {
	List []Item `json:"list"`
}

func (c *Client) GetConfigList() (*ConfigListResp, error) {
	type configListItem struct {
		Items []Item `json:"items"`
	}
	type configListByAppResp struct {
		List []configListItem `json:"list"`
	}
	type httpResp struct {
		response.BaseResult
		Data configListByAppResp `json:"data"`
	}
	var fullResp httpResp
	var listResp = &ConfigListResp{
		List: []Item{},
	}

	url := fmt.Sprintf("%s:%d/api/v1/client/config/list", c.server.Host, c.server.Port)
	res, err := util.HttpPostJson(url, nil)
	if err != nil {
		log.Error(err)
		return listResp, err
	}
	err = util.HttpParseResponseToJson(res, &fullResp)
	if err != nil {
		log.Error(err)
		return listResp, err
	}

	for _, item := range fullResp.Data.List {
		for _, v := range item.Items {
			listResp.List = append(listResp.List, v)
		}
	}

	return listResp, nil
}

type WatchConfigResp struct {
	InstanceId string                   `json:"instance_id"`
	EventType  com.ConfigWatchEventType `json:"event_type"`
	//Configs    map[com.OpType][]core.ChangeConfig `json:"configs"`
}

func (c *Client) fetchConfigEvent() (*WatchConfigResp, error) {
	resp := &WatchConfigResp{}

	if c.server.Status != com.OnlineStatus {
		log.Warn("server is not online, can not watch config")
		return resp, nil
	}

	url := fmt.Sprintf("%s:%d/api/v1/client/config/watch", c.server.Host, c.server.Port)
	res, err := util.HttpPostJson(url, nil)
	if err != nil {
		log.Error(err)
		return resp, err
	}
	var watchResp struct {
		response.BaseResult
		Data WatchConfigResp `json:"data"`
	}
	err = util.HttpParseResponseToJson(res, &watchResp)
	if err != nil {
		log.Error(err)
		return resp, err
	}

	resp = &watchResp.Data
	c.instanceId = resp.InstanceId

	if resp.EventType != com.CwNothing {
		log.Infof("watch config event type: %s", resp.EventType)
	}
	if resp.EventType == com.CwRefreshAll {
		config, err := c.GetConfigList()
		if err != nil {
			log.Error(err)
			return resp, err
		}

	}

	return resp, nil
}

func (c *Client) watchConfig() {
	defer func() {
		if err := recover(); err != nil {
			log.Warn("watch config recover: ", err)
			c.watchConfig()
		}
	}()

	for {
		cf, err := c.fetchConfigEvent()
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
