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
		c.config.Store(item.Key, &Item{
			Key:   item.Key,
			Value: item.Value,
		})
	}

	return nil
}

type ConfigItem struct {
	Id          string `json:"id"`
	NamespaceId string `json:"namespace_id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Comment     string `json:"comment"`
	OrderNum    int    `json:"order_num"`
	IsDelete    int    `json:"is_delete"`
	CreateBy    string `json:"create_by"`
	CreateTime  int    `json:"create_time"`
	UpdateBy    string `json:"update_by"`
	UpdateTime  int    `json:"update_time"`
	IsRelease   int    `json:"is_release"`
	Status      string `json:"status"`
}

type ConfigListResp struct {
	List []ConfigItem `json:"list"`
}

func (c *Client) GetConfigList() (*ConfigListResp, error) {
	type fullResp struct {
		response.BaseResult
		Data ConfigListResp `json:"data"`
	}
	var listResp fullResp

	url := fmt.Sprintf("%s:%d/api/v1/config/list", c.server.Host, c.server.Port)
	res, err := util.HttpPostJson(url, nil)
	if err != nil {
		log.Error(err)
		return &listResp.Data, err
	}
	err = util.HttpParseResponseToJson(res, &listResp)
	if err != nil {
		log.Error(err)
		return &listResp.Data, err
	}
	return &listResp.Data, nil
}

func (c *Client) fetchConfigChange() (config, error) {
	if c.server.Status != com.OnlineStatus {
		log.Warn("server is not online, can not watch config")
		return nil
	}

	url := fmt.Sprintf("%s:%d/api/v1/config/watch", c.server.Host, c.server.Port)
	res, err := util.HttpPostJson(url, nil)
	if err != nil {
		log.Error(err)
		return svr, err
	}

	var watchResp struct {
		response.BaseResult
		InstanceId string `json:"instance_id"`
	}
	err = util.HttpParseResponseToJson(res, &watchResp)
	if err != nil {
		log.Error(err)
		return svr, err
	}

	c.InstanceId = watchResp.InstanceId
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
