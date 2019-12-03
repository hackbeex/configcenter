package client

import (
	"encoding/json"
	"fmt"
	"github.com/hackbeex/configcenter/local"
	"github.com/hackbeex/configcenter/util"
	"github.com/hackbeex/configcenter/util/com"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/hackbeex/configcenter/util/response"
	"github.com/pkg/errors"
	"time"
)

type Configs struct {
	List map[string]string
}

func (c *Client) GetAllConfig() (*Configs, error) {
	var listResp = &Configs{
		List: map[string]string{},
	}
	c.config.Range(func(key string, val *Item) bool {
		listResp.List[key] = val.Value
		return true
	})
	return listResp, nil
}

func (c *Client) GetConfig(key string, defaultVal string) (string, bool) {
	item, ok := c.config.Load(key)
	if !ok {
		return defaultVal, ok
	}
	return item.Value, ok
}

//if key == "", listen all config change
func (c *Client) ListenConfig(key string, callback ListenCallback) {
	c.listens.AddCallback(key, callback)
}

func (c *Client) initConfig() error {
	res, err := c.fetchConfigList()
	if err != nil {
		if !local.Conf.Server.UseCache {
			log.Error(err)
			return err
		}
		log.Warn(err)

		config, err := c.cache.Load()
		if err != nil {
			return err
		}

		for key, val := range config {
			c.config.Store(key, &Item{
				Key:   key,
				Value: val,
			})
		}
	} else {
		config := map[string]string{}
		for _, item := range res.List {
			c.config.Store(item.Key, &Item{
				Key:   item.Key,
				Value: item.Value,
			})
			config[item.Key] = item.Value
		}

		if err := c.cache.Store(config); err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}

type ConfigListResp struct {
	List []Item `json:"list"`
}

func (c *Client) fetchConfigList() (*ConfigListResp, error) {
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

	data, _ := json.Marshal(map[string]string{
		"app":         c.App,
		"instance_id": c.instanceId,
	})
	url := fmt.Sprintf("http://%s:%d/api/v1/client/config/list", c.server.Host, c.server.Port)
	res, err := util.HttpPostJson(url, data)
	if err != nil {
		log.Error(err)
		return listResp, err
	}
	err = util.HttpParseResponseToJson(res, &fullResp)
	if err != nil {
		log.Error(err)
		return listResp, err
	}
	if fullResp.Code != 200 {
		log.Error(fullResp.Message)
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
}

func (c *Client) fetchConfigEvent() (*WatchConfigResp, error) {
	resp := &WatchConfigResp{}

	if c.server.Status != com.OnlineStatus {
		log.Warn("server is not online, can not watch config")
		return resp, nil
	}

	req, _ := json.Marshal(map[string]interface{}{
		"host":    c.Host,
		"port":    c.Port,
		"app":     c.App,
		"cluster": c.Cluster,
		"env":     c.Env,
	})
	url := fmt.Sprintf("http://%s:%d/api/v1/client/config/watch", c.server.Host, c.server.Port)
	res, err := util.HttpPostJson(url, req)
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
	if watchResp.Code != 200 {
		return resp, errors.New(watchResp.Message)
	}

	resp = &watchResp.Data
	c.instanceId = resp.InstanceId

	if resp.EventType != "" && resp.EventType != com.CwNothing {
		log.Infof("watch config event type: %s", resp.EventType)
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
			log.Info("watch config error: ", err.Error())
			if c.watchConfigInterval >= time.Minute*5 {
				c.watchConfigInterval = time.Minute * 5
			} else if c.watchConfigInterval > 0 {
				c.watchConfigInterval = c.watchConfigInterval * 2
			} else {
				c.watchConfigInterval = time.Second
			}
			log.Debugf("watch config error sleep: %ds", c.watchConfigInterval/time.Second)
			time.Sleep(c.watchConfigInterval)
			continue
		}
		c.watchConfigInterval = 0

		if cf.EventType == com.CwRefreshAll {
			if err := c.refreshConfig(); err != nil {
				continue
			}
		}
	}
}

func (c *Client) refreshConfig() error {
	res, err := c.fetchConfigList()
	if err != nil {
		log.Error(err)
		return err
	}
	old, err := c.GetAllConfig()
	if err != nil {
		log.Error(err)
		return err
	}

	var isChange = false
	newMap := map[string]string{}
	for _, item := range res.List {
		c.config.Store(item.Key, &Item{
			Key:   item.Key,
			Value: item.Value,
		})
		newMap[item.Key] = item.Value
	}
	for key, val := range old.List {
		if newVal, ok := newMap[key]; !ok {
			isChange = true
			c.listens.Call(key, &CallbackParam{
				Key:    key,
				NewVal: val,
				OldVal: val,
				OpType: com.OpDelete,
			}, true)
		} else if newVal != val {
			isChange = true
			c.listens.Call(key, &CallbackParam{
				Key:    key,
				NewVal: newVal,
				OldVal: val,
				OpType: com.OpUpdate,
			}, true)
		}
	}
	for key, newVal := range newMap {
		if _, ok := old.List[key]; !ok {
			isChange = true
			c.listens.Call(key, &CallbackParam{
				Key:    key,
				NewVal: newVal,
				OldVal: newVal,
				OpType: com.OpCreate,
			}, true)
		}
	}
	if isChange {
		if err := c.cache.Store(newMap); err != nil {
			log.Error(err)
		}
	}
	return nil
}

func (c *Client) timingPullConfig() {
	defer func() {
		if err := recover(); err != nil {
			log.Warn("timing pull config recover: ", err)
			c.timingPullConfig()
		}
	}()

	for {
		time.Sleep(time.Minute * 5)

		if err := c.refreshConfig(); err != nil {
			continue
		}
	}
}
