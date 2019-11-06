package model

import (
	"encoding/json"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/hackbeex/configcenter/server/database"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"strings"
	"time"
)

type ConfigModel struct {
}

type ConfigDetailReq struct {
	Id string `json:"id"`
}

func (c *ConfigDetailReq) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Id, validation.Required, validation.Length(32, 32)),
	)
}

type ConfigDetailResp struct {
	ConfigItem
}

type ConfigItem struct {
	Id          string `json:"id"`
	NamespaceId string `json:"namespace_id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Comment     string `json:"comment"`
	OrderNum    int    `json:"order_num"`
	CreateBy    string `json:"create_by"`
	CreateTime  int    `json:"create_time"`
	UpdateBy    string `json:"update_by"`
	UpdateTime  int    `json:"update_time"`
}

func (c *ConfigModel) Detail(req *ConfigDetailReq) (*ConfigDetailResp, error) {
	resp := &ConfigDetailResp{}

	if err := req.Validate(); err != nil {
		log.Warn(err)
		return resp, err
	}

	db := database.Conn()
	db = db.Table("item").Select("id,namespace_id,key,value,comment,order_num,create_by,create_time,update_by,update_time").
		Where("id=? AND is_delete=0", req.Id).Scan(&resp)
	if db.Error != nil {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}

	return resp, nil
}

type CreateConfigReq struct {
	NamespaceId string `json:"namespace_id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Comment     string `json:"comment"`
	UserId      string `json:"user_id"`
}

func (c *CreateConfigReq) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.NamespaceId, validation.Required, validation.Length(32, 32)),
		validation.Field(&c.Key, validation.Required, validation.Length(1, 128)),
		validation.Field(&c.Comment, validation.Length(1, 255)),
		validation.Field(&c.UserId, validation.Required, validation.Length(32, 32)),
	)
}

type CreateConfigResp struct {
	Id string `json:"id"`
}

func (c *ConfigModel) Create(req *CreateConfigReq) (*CreateConfigResp, error) {
	resp := &CreateConfigResp{}

	if err := req.Validate(); err != nil {
		log.Warn(err)
		return resp, err
	}

	req.Key = strings.TrimSpace(req.Key)

	var namespace struct {
		Id string
	}
	db := database.Conn()
	db = db.Table("namespace").Select("id").Where("id=? AND is_delete=0", req.NamespaceId).Scan(&namespace)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}
	if namespace.Id == "" {
		return resp, errors.New("the namespace not exists")
	}

	var existItem struct {
		Id string
	}
	db = database.Conn()
	db = db.Table("item").Select("id").Where("namespace_id=? AND key=? AND is_delete=0", req.NamespaceId, req.Key).Scan(&existItem)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}
	if existItem.Id != "" {
		return resp, errors.New("the config key exists")
	}

	now := time.Now().Unix()
	id := uuid.NewV1().String()
	tx := database.Conn().Begin()
	var itemOrderNum struct {
		MaxOrderNum int
	}
	tx = tx.Raw("SELECT MAX(order_num) max_order_num FROM item WHERE namespace_id=? FOR UPDATE").Scan(&itemOrderNum)
	item := map[string]interface{}{
		"id":           id,
		"namespace_id": req.NamespaceId,
		"key":          req.Key,
		"value":        req.Value,
		"comment":      req.Comment,
		"order_num":    itemOrderNum.MaxOrderNum + 1,
		"create_by":    req.UserId,
		"create_time":  now,
		"update_by":    req.UserId,
		"update_time":  now,
	}
	tx = database.Insert(tx, "item", item)
	tx = c.recordCommit(tx, req.NamespaceId, req.UserId, OpCreate, item)
	tx = RecordTable(tx, "item", id, "", req.UserId, OpCreate)
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

func (c *ConfigModel) recordCommit(db *gorm.DB, namespaceId, userId string, opType OpType, changeSet interface{}) *gorm.DB {
	data, _ := json.Marshal(map[OpType]interface{}{
		opType: changeSet,
	})
	now := time.Now().Unix()
	return database.Insert(db, "commit", map[string]interface{}{
		"id":           uuid.NewV1().String(),
		"namespace_id": namespaceId,
		"change_sets":  string(data),
		"create_by":    userId,
		"create_time":  now,
		"update_by":    userId,
		"update_time":  now,
	})
}

type UpdateConfigReq struct {
	Id      string `json:"id"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	Comment string `json:"comment"`
	UserId  string `json:"user_id"`
}

func (c *UpdateConfigReq) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Id, validation.Required, validation.Length(32, 32)),
		validation.Field(&c.Key, validation.Required, validation.Length(1, 128)),
		validation.Field(&c.Comment, validation.Length(1, 255)),
		validation.Field(&c.UserId, validation.Required, validation.Length(32, 32)),
	)
}

func (c *ConfigModel) Update(req *UpdateConfigReq) error {
	if err := req.Validate(); err != nil {
		log.Warn(err)
		return err
	}

	req.Key = strings.TrimSpace(req.Key)

	var oldItem struct {
		Id          string
		Key         string
		NamespaceId string
	}
	db := database.Conn()
	db = db.Table("item").Select("id,key,namespace_id").Where("id=? AND is_delete=0", req.Id).Scan(&oldItem)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
	}
	if oldItem.Id == "" {
		return errors.New("the item not exists")
	}
	if oldItem.Key != req.Key {
		return errors.New("the key not the same as before")
	}

	var existItem struct {
		Id string
	}
	db = database.Conn()
	db = db.Table("item").Select("id").
		Where("namespace_id=? AND key=? AND id!=? AND is_delete=0", oldItem.NamespaceId, req.Key, req.Id).
		Scan(&existItem)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
	}
	if existItem.Id != "" {
		return errors.New("the config key exists")
	}

	now := time.Now().Unix()
	item := map[string]interface{}{
		"id":          req.Id,
		"key":         req.Key,
		"value":       req.Value,
		"comment":     req.Comment,
		"update_by":   req.UserId,
		"update_time": now,
	}
	tx := database.Conn().Begin()
	tx = database.Update(tx, "item", item, "id=?", req.Id)
	tx = c.recordCommit(tx, oldItem.NamespaceId, req.UserId, OpUpdate, item)
	tx = RecordTable(tx, "item", req.Id, "", req.UserId, OpUpdate)
	if tx.Error != nil {
		tx.Rollback()
		log.Error(tx.Error)
		return errors.Wrap(tx.Error, "db error")
	} else {
		tx.Commit()
	}

	return nil
}

type DeleteConfigReq struct {
	Id     string `json:"id"`
	UserId string `json:"user_id"`
}

func (c *DeleteConfigReq) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Id, validation.Required, validation.Length(32, 32)),
		validation.Field(&c.UserId, validation.Required, validation.Length(32, 32)),
	)
}

func (c *ConfigModel) Delete(req *DeleteConfigReq) error {
	if err := req.Validate(); err != nil {
		log.Warn(err)
		return err
	}

	var oldItem struct {
		Id          string
		NamespaceId string
	}
	db := database.Conn()
	db = db.Table("item").Select("id,namespace_id").Where("id=? AND is_delete=0", req.Id).Scan(&oldItem)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
	}
	if oldItem.Id == "" {
		return errors.New("the item not exists")
	}

	now := time.Now().Unix()
	item := map[string]interface{}{
		"id":          req.Id,
		"is_delete":   1,
		"update_by":   req.UserId,
		"update_time": now,
	}
	tx := database.Conn().Begin()
	tx = database.Update(tx, "item", item, "id=?", req.Id)
	tx = c.recordCommit(tx, oldItem.NamespaceId, req.UserId, OpDelete, item)
	tx = RecordTable(tx, "item", req.Id, "", req.UserId, OpDelete)
	if tx.Error != nil {
		tx.Rollback()
		log.Error(tx.Error)
		return errors.Wrap(tx.Error, "db error")
	} else {
		tx.Commit()
	}

	return nil
}
