package model

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/hackbeex/configcenter/server/database"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type InstanceModel struct {
}

type InstanceListReq struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func (c *InstanceListReq) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Limit, validation.Max(100)),
		validation.Field(&c.Offset, validation.Min(0)),
	)
}

type InstanceItem struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Comment    string `json:"comment"`
	CreateBy   string `json:"create_by"`
	CreateTime int    `json:"create_time"`
	UpdateBy   string `json:"update_by"`
	UpdateTime int    `json:"update_time"`
}

type InstanceListResp struct {
	Offset int            `json:"offset"`
	Total  int            `json:"total"`
	List   []InstanceItem `json:"list"`
}

func (m *InstanceModel) List(req *InstanceListReq) (*InstanceListResp, error) {
	resp := &InstanceListResp{
		List:   []InstanceItem{},
		Offset: -1,
	}

	if err := req.Validate(); err != nil {
		log.Warn(err)
		return resp, err
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	//TODO instance op

	db := database.Conn()
	db = db.Table("app").Select("id,name,comment,create_by,create_time,update_by,update_time").
		Where("is_delete=0").Offset(req.Offset).Limit(req.Limit).Find(&resp.List)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}

	if len(resp.List) < req.Limit {
		resp.Offset = -1
	} else {
		resp.Offset = req.Offset + len(resp.List)
	}

	db = database.Conn()
	db = db.Table("app").Where("is_delete=0").Count(&resp.Total)
	if db.Error != nil {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}

	return resp, nil
}

type ExitInstanceReq struct {
	NamespaceId string `json:"namespace_id"`
}

func (c *ExitInstanceReq) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Limit, validation.Max(100)),
		validation.Field(&c.Offset, validation.Min(0)),
	)
}

func (m *InstanceModel) ExitInstance(req *ExitInstanceReq) error {
	if err := req.Validate(); err != nil {
		log.Warn(err)
		return err
	}

	return nil
}
