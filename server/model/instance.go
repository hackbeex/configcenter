package model

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/hackbeex/configcenter/server/core"
	"github.com/hackbeex/configcenter/server/database"
	"github.com/hackbeex/configcenter/util/com"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type InstanceModel struct {
}

type InstanceListReq struct {
	AppId     string `json:"app_id"`
	ClusterId string `json:"cluster_id"`
}

func (c *InstanceListReq) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.AppId, validation.Required),
		validation.Field(&c.ClusterId, validation.Required),
	)
}

type InstanceItem struct {
	Id         string `json:"id"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	CreateTime int    `json:"create_time"`
	UpdateTime int    `json:"update_time"`
}

type InstanceListResp struct {
	List []InstanceItem `json:"list"`
}

func (m *InstanceModel) List(req *InstanceListReq) (*InstanceListResp, error) {
	resp := &InstanceListResp{
		List: []InstanceItem{},
	}

	if err := req.Validate(); err != nil {
		log.Warn(err)
		return resp, err
	}

	db := database.Conn()
	db = db.Table("instance").Select("id,host,port,create_time,update_time").
		Where("app_id=? AND cluster_id=? AND is_delete=0", req.AppId, req.ClusterId).Find(&resp.List)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}

	return resp, nil
}

type ExitInstanceReq struct {
	InstanceId string `json:"instance_id"`
}

func (c *ExitInstanceReq) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.InstanceId, validation.Required),
	)
}

func (m *InstanceModel) ExitInstance(req *ExitInstanceReq) error {
	if err := req.Validate(); err != nil {
		log.Warn(err)
		return err
	}

	instances := core.GetServer().Instances
	ins, ok := instances.Load(req.InstanceId)
	if !ok {
		return nil
	}
	ins.Life = 0
	ins.Status = com.OfflineStatus
	instances.Store(req.InstanceId, ins)

	return nil
}
