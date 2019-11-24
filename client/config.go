package client

import (
	"fmt"
	"github.com/hackbeex/configcenter/util"
	"github.com/hackbeex/configcenter/util/com"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/hackbeex/configcenter/util/response"
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

		log.Info(cf)

		if cf.EventType == com.CwRefreshAll {
			res, err := c.fetchConfigList()
			if err != nil {
				log.Error(err)
				continue
			}
			old, err := c.GetAllConfig()
			if err != nil {
				log.Error(err)
				continue
			}
			type updateItem struct {
				OldVal string
				NewVal string
			}
			newMap := map[string]string{}
			for _, item := range res.List {
				c.config.Store(item.Key, &item)
				newMap[item.Key] = item.Value
			}
			for key, val := range old.List {
				if newVal, ok := newMap[key]; !ok {
					c.listens.Call(key, &CallbackParam{
						Key:    key,
						NewVal: val,
						OldVal: val,
						OpType: com.OpDelete,
					}, true)
				} else if newVal != val {
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
					c.listens.Call(key, &CallbackParam{
						Key:    key,
						NewVal: newVal,
						OldVal: newVal,
						OpType: com.OpCreate,
					}, true)
				}
			}
		}
	}
}
