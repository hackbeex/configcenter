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

type ReleaseOpType string

const (
	ReleaseOpNormal   ReleaseOpType = "normal"
	ReleaseOpRollback ReleaseOpType = "rollback"
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
	IsDelete    int    `json:"is_delete"`
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
		Where("id=?", req.Id).Scan(&resp)
	if db.Error != nil {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}

	return resp, nil
}

type ConfigListReq struct {
	NamespaceId string `json:"namespace_id"`
}

func (c *ConfigListReq) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.NamespaceId, validation.Required, validation.Length(32, 32)),
	)
}

type ConfigItemInfo struct {
	ConfigItem
	IsRelease int    `json:"is_release"`
	Status    OpType `json:"status"`
}

type ConfigListResp struct {
	List []ConfigItemInfo `json:"list"`
}

func (c *ConfigModel) List(req *ConfigListReq) (*ConfigListResp, error) {
	resp := &ConfigListResp{
		List: []ConfigItemInfo{},
	}

	if err := req.Validate(); err != nil {
		log.Warn(err)
		return resp, err
	}

	var release struct {
		UpdateTime int
		Config     []byte
	}
	db := database.Conn()
	db = db.Table("release").Select("config,update_time").Where("namespace_id=? AND is_delete=0", req.NamespaceId).
		Order("update_time DESC").Limit(1).Scan(&release)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}
	config := map[string]string{}
	if err := json.Unmarshal(release.Config, &config); err != nil {
		log.Error(err)
		return resp, err
	}

	db = database.Conn()
	db = db.Table("item").Select("*, 1 AS is_release").
		Where("namespace_id=? AND update_time<? AND is_delete=0", req.NamespaceId, release.UpdateTime).
		Find(&resp.List)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}

	//get not release items
	var unReleaseItems []ConfigItemInfo
	db = database.Conn()
	db = db.Table("item").Select("*, 0 AS is_release").
		Where("namespace_id=? AND update_time > ?", req.NamespaceId, release.UpdateTime).
		Find(&unReleaseItems)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}

	for _, item := range unReleaseItems {
		if _, ok := config[item.Key]; ok {
			if item.IsDelete == 1 {
				item.Status = OpDelete
			} else {
				item.Status = OpUpdate
			}
		} else if item.IsDelete == 1 {
			continue //unreleased item did't need to show
		} else {
			item.Status = OpCreate
		}
		resp.List = append(resp.List, item)
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
		Id       string
		IsDelete int
	}
	db = database.Conn()
	db = db.Table("item").Select("id,is_delete").Where("namespace_id=? AND key=?", req.NamespaceId, req.Key).Scan(&existItem)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}
	if existItem.Id != "" && existItem.IsDelete == 0 {
		return resp, errors.New("the config key exists")
	}

	id := existItem.Id
	if existItem.IsDelete == 0 {
		id = uuid.NewV1().String()
	}
	now := time.Now().Unix()
	tx := database.Conn().Begin()
	var itemOrderNum struct {
		MaxOrderNum int
	}
	tx = tx.Raw("SELECT MAX(order_num) max_order_num FROM item WHERE namespace_id=? FOR UPDATE").Scan(&itemOrderNum)
	if existItem.IsDelete == 1 {
		item := map[string]interface{}{
			"id":           id,
			"namespace_id": req.NamespaceId,
			"key":          req.Key,
			"value":        req.Value,
			"comment":      req.Comment,
			"order_num":    itemOrderNum.MaxOrderNum + 1,
			"is_delete":    0,
			"update_by":    req.UserId,
			"update_time":  now,
		}
		tx = database.Update(tx, "item", item, "id=?", id)
		tx = RecordTable(tx, "item", id, "", req.UserId, OpUpdate)
	} else {
		item := map[string]interface{}{
			"id":           id,
			"namespace_id": req.NamespaceId,
			"key":          req.Key,
			"value":        req.Value,
			"comment":      req.Comment,
			"order_num":    itemOrderNum.MaxOrderNum + 1,
			"is_delete":    0,
			"create_by":    req.UserId,
			"create_time":  now,
			"update_by":    req.UserId,
			"update_time":  now,
		}
		tx = database.Insert(tx, "item", item)
		tx = RecordTable(tx, "item", id, "", req.UserId, OpCreate)
	}
	if tx.Error != nil {
		tx.Rollback()
		log.Error(tx.Error)
		return resp, errors.Wrap(tx.Error, "db error")
	} else {
		tx.Commit()
	}

	c.recordCommit(id, OpCreate)
	resp.Id = id
	return resp, nil
}

func (c *ConfigModel) recordCommit(itemId string, opType OpType) {
	var item ConfigItem
	db := database.Conn()
	db = db.Table("item").Select("*").Scan(&item)
	if db.Error != nil {
		log.Error(db.Error)
		return
	}

	data, _ := json.Marshal(map[OpType]interface{}{
		opType: item,
	})
	now := time.Now().Unix()
	db = database.Conn()
	db = database.Insert(db, "commit", map[string]interface{}{
		"id":           uuid.NewV1().String(),
		"namespace_id": item.NamespaceId,
		"change_sets":  string(data),
		"create_by":    item.UpdateBy,
		"create_time":  now,
		"update_by":    item.UpdateBy,
		"update_time":  now,
	})
	if db.Error != nil {
		log.Error(db.Error)
		return
	}
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
		Value       string
		NamespaceId string
	}
	db := database.Conn()
	db = db.Table("item").Select("id,key,value,namespace_id").Where("id=? AND is_delete=0", req.Id).Scan(&oldItem)
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
	if oldItem.Value == req.Value {
		return errors.New("the value not change")
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
	tx = RecordTable(tx, "item", req.Id, "", req.UserId, OpUpdate)
	if tx.Error != nil {
		tx.Rollback()
		log.Error(tx.Error)
		return errors.Wrap(tx.Error, "db error")
	} else {
		tx.Commit()
	}

	c.recordCommit(req.Id, OpUpdate)
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
	tx = RecordTable(tx, "item", req.Id, "", req.UserId, OpDelete)
	if tx.Error != nil {
		tx.Rollback()
		log.Error(tx.Error)
		return errors.Wrap(tx.Error, "db error")
	} else {
		tx.Commit()
	}

	c.recordCommit(req.Id, OpDelete)
	return nil
}

type ConfigHistoryReq struct {
	NamespaceId string `json:"namespace_id"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

func (c *ConfigHistoryReq) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.NamespaceId, validation.Required, validation.Length(32, 32)),
		validation.Field(&c.Limit, validation.Max(100)),
		validation.Field(&c.Offset, validation.Min(0)),
	)
}

type CommitItem map[OpType]ConfigItem

type ConfigHistoryResp struct {
	List   []CommitItem `json:"list"`
	Offset int          `json:"offset"`
	Total  int          `json:"total"`
}

func (c *ConfigModel) GetHistory(req *ConfigHistoryReq) (*ConfigHistoryResp, error) {
	resp := &ConfigHistoryResp{
		List:   []CommitItem{},
		Offset: -1,
	}

	if err := req.Validate(); err != nil {
		log.Warn(err)
		return resp, err
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	var commits []struct {
		Id         string
		changeSets []byte
	}
	db := database.Conn()
	db = db.Table("commit").Select("id,change_sets").Where("namespace_id=? AND is_delete=0", req.NamespaceId).
		Order("update_time DESC").Limit(req.Limit).Offset(req.Offset).Find(&commits)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}

	for _, cm := range commits {
		var sets CommitItem
		if err := json.Unmarshal(cm.changeSets, &sets); err != nil {
			log.Error(err)
			continue
		}
		resp.List = append(resp.List, sets)
	}

	if len(resp.List) < req.Limit {
		resp.Offset = -1
	} else {
		resp.Offset = req.Offset + len(resp.List)
	}

	db = database.Conn()
	db = db.Table("commit").Where("namespace_id=? AND is_delete=0", req.NamespaceId).Count(&resp.Total)
	if db.Error != nil {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}

	return resp, nil
}

type ReleaseConfigReq struct {
	NamespaceId string `json:"namespace_id"`
	Name        string `json:"name"`
	Comment     string `json:"comment"`
	UserId      string `json:"user_id"`
}

func (c *ReleaseConfigReq) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.NamespaceId, validation.Required, validation.Length(32, 32)),
		validation.Field(&c.Name, validation.Required, validation.Length(1, 64)),
		validation.Field(&c.Comment, validation.Length(1, 255)),
		validation.Field(&c.UserId, validation.Required, validation.Length(32, 32)),
	)
}

func (c *ConfigModel) Release(req *ReleaseConfigReq) error {
	if err := req.Validate(); err != nil {
		log.Warn(err)
		return err
	}

	var namespace struct {
		Id        string
		AppId     string
		ClusterId string
	}
	db := database.Conn()
	db = db.Table("namespace").Select("id,app_id,cluster_id").Where("id=? AND is_delete=0", req.NamespaceId).Scan(&namespace)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
	}
	if namespace.Id == "" {
		return errors.New("the namespace not exists")
	}

	var preRelease struct {
		Id string
	}
	db = database.Conn()
	db = db.Table("release").Select("id").Where("namespace_id=? AND is_delete=0", req.NamespaceId).
		Order("update_time DESC").Limit(1).Scan(&preRelease)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
	}

	var items []struct {
		Key   string
		Value string
	}
	db = database.Conn()
	db = db.Table("item").Select("key,value").Where("namespace_id=? AND is_delete=0", req.NamespaceId).Find(&items)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
	}
	if len(items) == 0 {
		return errors.New("no new configs to release")
	}

	itemMap := map[string]string{}
	for _, item := range items {
		itemMap[item.Key] = item.Value
	}
	config, _ := json.Marshal(itemMap)

	now := time.Now().Unix()
	id := uuid.NewV1().String()
	release := map[string]interface{}{
		"id":           id,
		"name":         req.Name,
		"comment":      req.Comment,
		"app_id":       namespace.AppId,
		"cluster_id":   namespace.ClusterId,
		"namespace_id": req.NamespaceId,
		"config":       config,
		"create_by":    req.UserId,
		"create_time":  now,
		"update_by":    req.UserId,
		"update_time":  now,
	}
	historyId := uuid.NewV1().String()
	releaseHistory := map[string]interface{}{
		"id":             historyId,
		"app_id":         namespace.AppId,
		"namespace_id":   req.NamespaceId,
		"cluster_id":     namespace.ClusterId,
		"release_id":     id,
		"pre_release_id": preRelease.Id,
		"op_type":        ReleaseOpNormal,
		"create_by":      req.UserId,
		"create_time":    now,
		"update_by":      req.UserId,
		"update_time":    now,
	}

	tx := database.Conn().Begin()
	tx = database.Insert(tx, "release", release)
	tx = database.Insert(tx, "release_history", releaseHistory)
	tx = RecordTable(tx, "release", id, "", req.UserId, OpCreate)
	tx = RecordTable(tx, "release_history", historyId, "", req.UserId, OpCreate)
	if tx.Error != nil {
		tx.Rollback()
		log.Error(tx.Error)
		return errors.Wrap(tx.Error, "db error")
	} else {
		tx.Commit()
	}

	return nil
}

type ConfigReleaseHistoryReq struct {
	NamespaceId string `json:"namespace_id"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

func (c *ConfigReleaseHistoryReq) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.NamespaceId, validation.Required, validation.Length(32, 32)),
		validation.Field(&c.Limit, validation.Max(100)),
		validation.Field(&c.Offset, validation.Min(0)),
	)
}

type ChangeItem struct {
	Key      string
	Type     OpType
	NewValue string
	OldValue string
}

type ReleaseItem struct {
	Id           string
	ReleaseId    string
	PreReleaseId string
	OpType       ReleaseOpType
	CreateBy     string
	CreateTime   int
	UpdateBy     string
	UpdateTime   int
	Name         string
	Comment      string
	Config       map[string]string
	Change       map[string]ChangeItem
}

type ConfigReleaseHistoryResp struct {
	List   []ReleaseItem `json:"list"`
	Offset int           `json:"offset"`
	Total  int           `json:"total"`
}

func (c *ConfigModel) GetReleaseHistory(req *ConfigReleaseHistoryReq) (*ConfigReleaseHistoryResp, error) {
	resp := &ConfigReleaseHistoryResp{
		List:   []ReleaseItem{},
		Offset: -1,
	}

	if err := req.Validate(); err != nil {
		log.Warn(err)
		return resp, err
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	var releaseHistories []struct {
		Id           string
		ReleaseId    string
		PreReleaseId string
		OpType       ReleaseOpType
		CreateBy     string
		CreateTime   int
		UpdateBy     string
		UpdateTime   int
		Name         string
		Comment      string
		Config       []byte
		PreConfig    []byte
	}
	db := database.Conn()
	db = db.Table("release_history t1").
		Select("t1.id,t1.release_id,t1.pre_release_id,t1.op_type,t1.create_by,t1.create_time,"+
			"t1.update_by,t1.update_time,t2.name,t2.comment,t2.config,t3.config AS pre_config").
		Joins("JOIN release t2 ON t2.id=t1.release_id AND t2.is_delete=0").
		Joins("LEFT JOIN release t3 ON t3.id=t1.pre_release_id AND t3.is_delete=0").
		Where("t1.namespace_id=? AND t1.is_delete=0", req.NamespaceId).
		Order("update_time DESC").Limit(req.Limit).Offset(req.Offset).
		Find(&releaseHistories)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}

	for _, history := range releaseHistories {
		config := map[string]string{}
		if err := json.Unmarshal(history.Config, &config); err != nil {
			log.Error(err)
			continue
		}
		//get changed config items
		change := map[string]ChangeItem{}
		if history.PreReleaseId != "" {
			preConfig := map[string]string{}
			if err := json.Unmarshal(history.PreConfig, &preConfig); err != nil {
				log.Error(err)
				continue
			}
			for key, val := range config {
				if pre, ok := preConfig[key]; ok {
					if pre != val {
						change[key] = ChangeItem{
							Key:      key,
							Type:     OpUpdate,
							NewValue: val,
							OldValue: pre,
						}
					}
				} else {
					change[key] = ChangeItem{
						Key:      key,
						Type:     OpCreate,
						NewValue: val,
						OldValue: "",
					}
				}
			}
			for key, pre := range preConfig {
				if _, ok := config[key]; !ok {
					change[key] = ChangeItem{
						Key:      key,
						Type:     OpDelete,
						NewValue: "",
						OldValue: pre,
					}
				}
			}
		}
		resp.List = append(resp.List, ReleaseItem{
			Id:           history.Id,
			ReleaseId:    history.ReleaseId,
			PreReleaseId: history.PreReleaseId,
			OpType:       history.OpType,
			CreateBy:     history.CreateBy,
			CreateTime:   history.CreateTime,
			UpdateBy:     history.UpdateBy,
			UpdateTime:   history.UpdateTime,
			Name:         history.Name,
			Comment:      history.Comment,
			Config:       config,
			Change:       change,
		})
	}

	if len(resp.List) < req.Limit {
		resp.Offset = -1
	} else {
		resp.Offset = req.Offset + len(resp.List)
	}

	db = database.Conn()
	db = db.Table("release_history t1").
		Joins("JOIN release t2 ON t2.id=t1.release_id AND t2.is_delete=0").
		Where("t1.namespace_id=? AND t1.is_delete=0", req.NamespaceId).Count(&resp.Total)
	if db.Error != nil {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}

	return resp, nil
}
