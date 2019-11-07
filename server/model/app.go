package model

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/hackbeex/configcenter/server/database"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"time"
)

type AppModel struct {
}

type CreateAppReq struct {
	Appid   string `json:"appid"`
	Name    string `json:"name"`
	Comment string `json:"comment"`
	UserId  string `json:"user_id"`
}

func (c *CreateAppReq) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Appid, validation.Required, validation.Length(1, 64)),
		validation.Field(&c.Name, validation.Required, validation.Length(1, 64)),
		validation.Field(&c.Comment, validation.Length(1, 255)),
		validation.Field(&c.UserId, validation.Required, validation.Length(32, 32)),
	)
}

type CreateAppResp struct {
	Id string `json:"id"`
}

func (a *AppModel) Create(req *CreateAppReq) (*CreateAppResp, error) {
	resp := &CreateAppResp{}

	if err := req.Validate(); err != nil {
		log.Warn(err)
		return resp, err
	}

	var existApp struct {
		Id string
	}
	db := database.Conn()
	db = db.Table("app").Select("id").Where("appid=? AND is_delete=0", req.Appid).Scan(&existApp)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}
	if existApp.Id != "" {
		return resp, errors.New("the appid exists")
	}

	now := time.Now().Unix()
	id := uuid.NewV1().String()
	tx := database.Conn().Begin()
	tx = database.Insert(tx, "app", map[string]interface{}{
		"id":          id,
		"appid":       req.Appid,
		"name":        req.Name,
		"comment":     req.Comment,
		"create_by":   req.UserId,
		"create_time": now,
		"update_by":   req.UserId,
		"update_time": now,
	})
	tx = RecordTable(tx, "app", id, "", req.UserId, OpCreate)
	if tx.Error != nil {
		tx.Rollback()
		log.Error(tx.Error)
		return resp, errors.Wrap(tx.Error, "db error")
	} else {
		tx.Commit()
	}

	resp.Id = id
	return resp, nil
}

type AppListReq struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func (c *AppListReq) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Limit, validation.Max(100)),
		validation.Field(&c.Offset, validation.Min(0)),
	)
}

type AppItem struct {
	Id         string `json:"id"`
	Appid      string `json:"appid"`
	Name       string `json:"name"`
	Comment    string `json:"comment"`
	CreateBy   string `json:"create_by"`
	CreateTime int    `json:"create_time"`
	UpdateBy   string `json:"update_by"`
	UpdateTime int    `json:"update_time"`
}

type AppListResp struct {
	Offset int       `json:"offset"`
	Total  int       `json:"total"`
	List   []AppItem `json:"list"`
}

func (a *AppModel) List(req *AppListReq) (*AppListResp, error) {
	resp := &AppListResp{
		List:   []AppItem{},
		Offset: -1,
	}

	if err := req.Validate(); err != nil {
		log.Warn(err)
		return resp, err
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	db := database.Conn()
	db = db.Table("app").Select("id,appid,name,comment,create_by,create_time,update_by,update_time").
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

type AppDetailReq struct {
	Id string `json:"id"`
}

func (c *AppDetailReq) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Id, validation.Required, validation.Length(32, 32)),
	)
}

type AppDetailResp struct {
	AppItem
}

func (a *AppModel) Detail(req *AppDetailReq) (*AppDetailResp, error) {
	resp := &AppDetailResp{}

	if err := req.Validate(); err != nil {
		log.Warn(err)
		return resp, err
	}

	db := database.Conn()
	db = db.Table("app").Select("id,appid,name,comment,create_by,create_time,update_by,update_time").
		Where("id=? AND is_delete=0", req.Id).Scan(&resp)
	if db.Error != nil {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}

	return resp, nil
}
