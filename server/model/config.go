package model

import (
	"encoding/json"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/hackbeex/configcenter/server/core"
	"github.com/hackbeex/configcenter/server/database"
	"github.com/hackbeex/configcenter/util/com"
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
	return validation.ValidateStruct(c,
		validation.Field(&c.Id, validation.Required, validation.Length(36, 36)),
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
	db = db.Table("item").Select("id,namespace_id,`key`,value,comment,order_num,create_by,create_time,update_by,update_time").
		Where("id=? AND is_delete=0", req.Id).Scan(&resp)
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
	return validation.ValidateStruct(c,
		validation.Field(&c.NamespaceId, validation.Required, validation.Length(36, 36)),
	)
}

type ConfigItemInfo struct {
	ConfigItem
	IsRelease int        `json:"is_release"`
	Status    com.OpType `json:"status"`
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

	//get current release config
	release, err := c.getLastRelease(req.NamespaceId)
	if err != nil {
		return resp, err
	}

	db := database.Conn()
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
		if _, ok := release.Config[item.Key]; ok {
			if item.IsDelete == 1 {
				item.Status = com.OpDelete
			} else {
				item.Status = com.OpUpdate
			}
		} else if item.IsDelete == 1 {
			continue //unreleased item do not need to show
		} else {
			item.Status = com.OpCreate
		}
		resp.List = append(resp.List, item)
	}

	return resp, nil
}

type lastRelease struct {
	Id         string
	UpdateTime int
	Config     map[string]string
}

func (c *ConfigModel) getLastRelease(namespaceId string) (*lastRelease, error) {
	var resp = &lastRelease{}
	var release struct {
		Id         string
		UpdateTime int
		Config     []byte
	}
	db := database.Conn()
	db = db.Table("release_history t1").Select("t1.id,t2.config,t1.update_time").
		Joins("JOIN `release` t2 ON t1.release_id=t2.id AND t2.is_delete=0").
		Where("t1.namespace_id=? AND t1.op_type=? AND t1.is_delete=0", namespaceId, ReleaseOpNormal).
		Order("t1.update_time DESC").Limit(1).Scan(&release)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}

	if len(release.Config) > 0 {
		if err := json.Unmarshal(release.Config, &resp.Config); err != nil {
			log.Error(err)
			return resp, err
		}
	}
	resp.Id = release.Id
	resp.UpdateTime = release.UpdateTime
	return resp, nil
}

type ConfigListByAppReq struct {
	App        string `json:"app"`
	InstanceId string `json:"instance_id"`
}

func (c *ConfigListByAppReq) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.App, validation.Required, validation.Length(1, 64)),
	)
}

type ItemSimple struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type ConfigListByAppItem struct {
	Namespace NamespaceItem `json:"namespace"`
	Items     []ItemSimple  `json:"items"`
}

type ConfigListByAppResp struct {
	List []ConfigListByAppItem `json:"list"`
}

func (c *ConfigModel) ListByApp(req *ConfigListByAppReq) (*ConfigListByAppResp, error) {
	resp := &ConfigListByAppResp{
		List: []ConfigListByAppItem{},
	}

	var app struct {
		Id string
	}
	db := database.Conn()
	db = db.Table("app").Select("id").Where("name=? AND is_delete=0", req.App).Scan(&app)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}
	if app.Id == "" {
		return resp, errors.New("app name not exists")
	}

	if req.InstanceId != "" {
		var instance struct {
			Id string
		}
		db := database.Conn()
		db = db.Table("instance").Select("id").Where("id=? AND app_id=? AND is_delete=0", req.InstanceId, app.Id).Scan(&instance)
		if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
			log.Error(db.Error)
			return resp, errors.Wrap(db.Error, "db error")
		}
		if instance.Id == "" {
			log.Warnf("instance id[%s] not matches app name[%s]", req.InstanceId, req.App)
			return resp, errors.New("instance id not matches app id")
		}
	}

	appMdl := AppModel{}
	appDetail, err := appMdl.Detail(&AppDetailReq{
		AppId: app.Id,
	})
	if err != nil {
		return resp, err
	}

	var lastReleaseId string
	var lastUpdateTime int
	for _, namespace := range appDetail.Namespaces {
		release, err := c.getLastRelease(namespace.Id)
		if err != nil {
			return resp, err
		}
		if release.UpdateTime > lastUpdateTime {
			lastReleaseId = release.Id
		}
		items := make([]ItemSimple, 0, len(release.Config))
		for k, v := range release.Config {
			items = append(items, ItemSimple{
				Key:   k,
				Value: v,
			})
		}
		resp.List = append(resp.List, ConfigListByAppItem{
			Namespace: namespace,
			Items:     items,
		})
	}

	if req.InstanceId != "" {
		now := time.Now().Unix()
		var release struct {
			Id               string `json:"id"`
			ReleaseHistoryId string `json:"release_history_id"`
		}
		db := database.Conn()
		db = db.Table("instance_release").Select("id,release_history_id").Where("instance_id=? AND is_delete=0", req.InstanceId).Scan(&release)
		if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
			log.Error(db.Error)
			return resp, errors.Wrap(db.Error, "db error")
		}
		tx := database.Conn().Begin()
		if release.Id != "" {
			if lastReleaseId != release.ReleaseHistoryId {
				tx = tx.Exec("UPDATE instance_release SET release_history_id=?, update_time=? WHERE id=?", lastReleaseId, now, release.Id)
				tx = RecordTable(tx, "instance_release", "", "", com.OpUpdate, release.Id)
			}
		} else {
			id := uuid.NewV1().String()
			tx = database.Insert(tx, "instance_release", map[string]interface{}{
				"id":                 id,
				"instance_id":        req.InstanceId,
				"release_history_id": lastReleaseId,
				"create_time":        now,
				"update_time":        now,
			})
			tx = RecordTable(tx, "instance_release", "", "", com.OpCreate, id)
		}
		if tx.Error != nil {
			tx.Rollback()
			log.Error(tx.Error)
			return resp, errors.Wrap(tx.Error, "db error")
		} else {
			tx.Commit()
		}
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
	return validation.ValidateStruct(c,
		validation.Field(&c.NamespaceId, validation.Required, validation.Length(36, 36)),
		validation.Field(&c.Key, validation.Required, validation.Length(1, 128)),
		validation.Field(&c.Comment, validation.Length(1, 255)),
		validation.Field(&c.UserId, validation.Required, validation.Length(36, 36)),
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
	db = db.Table("item").Select("id,is_delete").Where("namespace_id=? AND `key`=?", req.NamespaceId, req.Key).Scan(&existItem)
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
	tx = tx.Raw("SELECT MAX(order_num) max_order_num FROM item WHERE namespace_id=? FOR UPDATE", req.NamespaceId).Scan(&itemOrderNum)
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
		tx = RecordTable(tx, "item", "", req.UserId, com.OpUpdate, id)
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
		tx = RecordTable(tx, "item", "", req.UserId, com.OpCreate, id)
	}
	if tx.Error != nil {
		tx.Rollback()
		log.Error(tx.Error)
		return resp, errors.Wrap(tx.Error, "db error")
	} else {
		tx.Commit()
	}

	c.recordItem(com.OpCreate, req.UserId, id)
	resp.Id = id
	return resp, nil
}

func (c *ConfigModel) recordItem(opType com.OpType, userId string, itemId ...string) {
	if len(itemId) == 0 {
		log.Warn("no itemId to recordItem")
		return
	}

	var items []ConfigItem
	db := database.Conn()
	db = db.Table("item").Select("*").Where("id IN (?)", itemId).Find(&items)
	if db.Error != nil {
		log.Error(db.Error)
		return
	}
	itemMap := map[string][]ConfigItem{}
	for _, item := range items {
		itemMap[item.NamespaceId] = append(itemMap[item.NamespaceId], item)
	}

	now := time.Now().Unix()
	tx := database.Conn().Begin()
	for nsId, nsItems := range itemMap {
		data, _ := json.Marshal(CommitItem{
			opType: nsItems,
		})
		tx = database.Insert(tx, "commit", map[string]interface{}{
			"id":           uuid.NewV1().String(),
			"namespace_id": nsId,
			"change_sets":  data,
			"create_by":    userId,
			"create_time":  now,
			"update_by":    userId,
			"update_time":  now,
		})
	}
	if tx.Error != nil {
		tx.Rollback()
		log.Error(tx.Error)
	} else {
		tx.Commit()
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
	return validation.ValidateStruct(c,
		validation.Field(&c.Id, validation.Required, validation.Length(36, 36)),
		validation.Field(&c.Key, validation.Required, validation.Length(1, 128)),
		validation.Field(&c.Comment, validation.Length(1, 255)),
		validation.Field(&c.UserId, validation.Required, validation.Length(36, 36)),
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
		Comment     string
		NamespaceId string
	}
	db := database.Conn()
	db = db.Table("item").Select("id,`key`,value,comment,namespace_id").Where("id=? AND is_delete=0", req.Id).Scan(&oldItem)
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
	if oldItem.Value == req.Value && req.Comment == oldItem.Comment {
		return nil
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
	tx = RecordTable(tx, "item", "", req.UserId, com.OpUpdate, req.Id)
	if tx.Error != nil {
		tx.Rollback()
		log.Error(tx.Error)
		return errors.Wrap(tx.Error, "db error")
	} else {
		tx.Commit()
	}

	c.recordItem(com.OpUpdate, req.UserId, req.Id)
	return nil
}

type DeleteConfigReq struct {
	Id     string `json:"id"`
	UserId string `json:"user_id"`
}

func (c *DeleteConfigReq) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Id, validation.Required, validation.Length(36, 36)),
		validation.Field(&c.UserId, validation.Required, validation.Length(36, 36)),
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
	tx = RecordTable(tx, "item", "", req.UserId, com.OpDelete, req.Id)
	if tx.Error != nil {
		tx.Rollback()
		log.Error(tx.Error)
		return errors.Wrap(tx.Error, "db error")
	} else {
		tx.Commit()
	}

	c.recordItem(com.OpDelete, req.UserId, req.Id)
	return nil
}

type ConfigHistoryReq struct {
	NamespaceId string `json:"namespace_id"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

func (c *ConfigHistoryReq) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.NamespaceId, validation.Required, validation.Length(36, 36)),
		validation.Field(&c.Limit, validation.Max(100)),
		validation.Field(&c.Offset, validation.Min(0)),
	)
}

type CommitItem map[com.OpType][]ConfigItem

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
		ChangeSets []byte
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
		if err := json.Unmarshal(cm.ChangeSets, &sets); err != nil {
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
	return validation.ValidateStruct(c,
		validation.Field(&c.NamespaceId, validation.Required, validation.Length(36, 36)),
		validation.Field(&c.Name, validation.Required, validation.Length(1, 64)),
		validation.Field(&c.Comment, validation.Length(1, 255)),
		validation.Field(&c.UserId, validation.Required, validation.Length(36, 36)),
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

	//get current release config
	lastRelease, err := c.getLastRelease(req.NamespaceId)
	if err != nil {
		return err
	}

	var preReleaseId string
	var lastHistory struct {
		ReleaseId string
		OpType    ReleaseOpType
	}
	db = database.Conn()
	db = db.Table("release_history").Select("release_id,op_type").Where("namespace_id=? AND is_delete=0", req.NamespaceId).
		Order("update_time DESC").Limit(1).Scan(&lastHistory)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
	}
	if lastHistory.OpType == ReleaseOpNormal {
		preReleaseId = lastHistory.ReleaseId
	} else { //ReleaseOpRollback
		var lastReleaseHistory struct {
			PreReleaseId string
		}
		db = database.Conn()
		db = db.Table("release_history").Select("pre_release_id").
			Where("release_id=? AND op_type=?", lastHistory.ReleaseId, ReleaseOpNormal).
			Order("update_time DESC").Limit(1).Scan(&lastReleaseHistory)
		if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
			log.Error(db.Error)
			return errors.Wrap(db.Error, "db error")
		}
		preReleaseId = lastReleaseHistory.PreReleaseId
	}

	//get not release items
	var unReleaseItem struct {
		Id string
	}
	db = database.Conn()
	db = db.Table("item").Select("id").
		Where("namespace_id=? AND update_time > ?", req.NamespaceId, lastRelease.UpdateTime).
		Limit(1).Scan(&unReleaseItem)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
	}
	if unReleaseItem.Id == "" {
		return errors.New("no new configs to release")
	}

	var items []struct {
		Key   string
		Value string
	}
	db = database.Conn()
	db = db.Table("item").Select("`key`,value").Where("namespace_id=? AND is_delete=0", req.NamespaceId).Find(&items)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
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
		"pre_release_id": preReleaseId,
		"op_type":        ReleaseOpNormal,
		"create_by":      req.UserId,
		"create_time":    now,
		"update_by":      req.UserId,
		"update_time":    now,
	}

	tx := database.Conn().Begin()
	tx = database.Insert(tx, "`release`", release)
	tx = database.Insert(tx, "release_history", releaseHistory)
	tx = RecordTable(tx, "release", "", req.UserId, com.OpCreate, id)
	tx = RecordTable(tx, "release_history", "", req.UserId, com.OpCreate, historyId)
	if tx.Error != nil {
		tx.Rollback()
		log.Error(tx.Error)
		return errors.Wrap(tx.Error, "db error")
	} else {
		tx.Commit()
	}

	go c.notifyChange()

	return nil
}

type ConfigReleaseHistoryReq struct {
	NamespaceId string `json:"namespace_id"`
	Limit       int    `json:"limit"`
	Offset      int    `json:"offset"`
}

func (c *ConfigReleaseHistoryReq) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.NamespaceId, validation.Required, validation.Length(36, 36)),
		validation.Field(&c.Limit, validation.Max(100)),
		validation.Field(&c.Offset, validation.Min(0)),
	)
}

type ChangeItem struct {
	Key      string
	Type     com.OpType
	NewValue string
	OldValue string
}

type ReleaseItem struct {
	Id           string                `json:"id"`
	ReleaseId    string                `json:"release_id"`
	PreReleaseId string                `json:"pre_release_id"`
	OpType       ReleaseOpType         `json:"op_type"`
	CreateBy     string                `json:"create_by"`
	CreateTime   int                   `json:"create_time"`
	UpdateBy     string                `json:"update_by"`
	UpdateTime   int                   `json:"update_time"`
	Name         string                `json:"name"`
	Comment      string                `json:"comment"`
	Config       map[string]string     `json:"config"`
	Change       map[string]ChangeItem `json:"change"`
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
		Joins("JOIN `release` t2 ON t2.id=t1.release_id AND t2.is_delete=0").
		Joins("LEFT JOIN `release` t3 ON t3.id=t1.pre_release_id AND t3.is_delete=0").
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
							Type:     com.OpUpdate,
							NewValue: val,
							OldValue: pre,
						}
					}
				} else {
					change[key] = ChangeItem{
						Key:      key,
						Type:     com.OpCreate,
						NewValue: val,
						OldValue: "",
					}
				}
			}
			for key, pre := range preConfig {
				if _, ok := config[key]; !ok {
					change[key] = ChangeItem{
						Key:      key,
						Type:     com.OpDelete,
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
		Joins("JOIN `release` t2 ON t2.id=t1.release_id AND t2.is_delete=0").
		Where("t1.namespace_id=? AND t1.is_delete=0", req.NamespaceId).Count(&resp.Total)
	if db.Error != nil {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}

	return resp, nil
}

type RollbackConfigReq struct {
	NamespaceId string `json:"namespace_id"`
	UserId      string `json:"user_id"`
}

func (c *RollbackConfigReq) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.NamespaceId, validation.Required, validation.Length(36, 36)),
		validation.Field(&c.UserId, validation.Required, validation.Length(36, 36)),
	)
}

//rollback is not a type of release, after rollback need to release manually.
//you can rollback multiple times at a time.
func (c *ConfigModel) Rollback(req *RollbackConfigReq) error {
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

	//get last valid history which can rollback
	var lastHistory struct {
		ReleaseId    string
		PreReleaseId string
	}
	db = database.Conn()
	db = db.Table("release_history t1").Select("t2.release_id,t2.pre_release_id").
		Joins("JOIN release_history t2 ON t1.release_id=t2.release_id AND t2.op_type=? AND t2.is_delete=0", ReleaseOpNormal).
		Where("t1.namespace_id=? AND t1.is_delete=0", req.NamespaceId).
		Order("t1.update_time DESC").Limit(1).Scan(&lastHistory)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
	}
	if lastHistory.PreReleaseId == "" {
		return errors.New("the config is the first version, can not rollback anymore")
	}

	//get history config to rollback
	var backRelease struct {
		Config []byte
	}
	db = database.Conn()
	db = db.Table("release").Select("config").Where("id=?", lastHistory.PreReleaseId).Scan(&backRelease)
	if db.Error != nil {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
	}
	var config map[string]string
	if err := json.Unmarshal(backRelease.Config, &config); err != nil {
		log.Error(err)
		return err
	}

	var items []struct {
		Id    string
		Key   string
		Value string
	}
	db = database.Conn()
	db = db.Table("item").Select("id,`key`,value").Where("namespace_id=?", req.NamespaceId).Find(&items)
	if db.Error != nil {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
	}

	//mark changed config items
	now := time.Now().Unix()
	var updateItems []map[string]interface{}
	var updateItemIds []string
	for _, item := range items {
		if val, ok := config[item.Key]; ok {
			if val != item.Value {
				updateItems = append(updateItems, map[string]interface{}{
					"id":          item.Id,
					"value":       val,
					"is_delete":   0,
					"update_time": now,
					"update_by":   req.UserId,
				})
				updateItemIds = append(updateItemIds, item.Id)
			}
		} else {
			updateItems = append(updateItems, map[string]interface{}{
				"id":          item.Id,
				"value":       item.Value,
				"is_delete":   1,
				"update_time": now,
				"update_by":   req.UserId,
			})
			updateItemIds = append(updateItemIds, item.Id)
		}
	}

	release := map[string]interface{}{
		"is_disabled": 1,
		"update_by":   req.UserId,
		"update_time": now,
	}
	historyId := uuid.NewV1().String()
	releaseHistory := map[string]interface{}{
		"id":             historyId,
		"app_id":         namespace.AppId,
		"cluster_id":     namespace.ClusterId,
		"namespace_id":   req.NamespaceId,
		"release_id":     lastHistory.PreReleaseId,
		"pre_release_id": lastHistory.ReleaseId,
		"op_type":        ReleaseOpRollback,
		"create_by":      req.UserId,
		"create_time":    now,
		"update_by":      req.UserId,
		"update_time":    now,
	}

	tx := database.Conn().Begin()
	for _, item := range updateItems {
		tx = database.Update(tx, "item", item, "id=?", item["id"])
	}
	tx = database.Update(tx, "`release`", release, "id=?", lastHistory.PreReleaseId)
	tx = database.Insert(tx, "release_history", releaseHistory)
	tx = RecordTable(tx, "release", "", req.UserId, com.OpUpdate, lastHistory.PreReleaseId)
	tx = RecordTable(tx, "release_history", "", req.UserId, com.OpCreate, historyId)
	if tx.Error != nil {
		tx.Rollback()
		log.Error(tx.Error)
		return errors.Wrap(tx.Error, "db error")
	} else {
		tx.Commit()
	}

	if len(updateItemIds) > 0 {
		c.recordItem(com.OpUpdate, req.UserId, updateItemIds...)
	}

	return nil
}

type SyncConfigReq struct {
	FromNamespaceId string   `json:"namespace_id"`
	ToClusterIds    []string `json:"to_cluster_ids"`
	Keys            []string `json:"keys"`
	UserId          string   `json:"user_id"`
}

func (c *SyncConfigReq) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.FromNamespaceId, validation.Required, validation.Length(36, 36)),
		validation.Field(&c.ToClusterIds, validation.Required),
		validation.Field(&c.Keys, validation.Required),
		validation.Field(&c.UserId, validation.Required, validation.Length(36, 36)),
	)
}

//sync is not a type of release, after sync need to release manually
func (c *ConfigModel) Sync(req *SyncConfigReq) error {
	if err := req.Validate(); err != nil {
		log.Warn(err)
		return err
	}

	var namespace struct {
		Id        string
		Name      string
		ClusterId string
	}
	db := database.Conn()
	db = db.Table("namespace").Select("id,name,cluster_id").Where("id=? AND is_delete=0", req.FromNamespaceId).Scan(&namespace)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
	}
	if namespace.Id == "" {
		return errors.New("the namespace not exists")
	}
	for _, cid := range req.ToClusterIds {
		if cid == namespace.ClusterId {
			return errors.New("the clusters to be sync contains source cluster")
		}
	}

	var clusters []struct {
		Id string
	}
	db = database.Conn()
	db = db.Table("cluster").Select("id,name").Where("id IN (?) AND is_delete=0", req.ToClusterIds).Scan(&clusters)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
	}
	if len(req.ToClusterIds) != len(clusters) {
		return errors.New("the clusters to be sync not exist")
	}

	//get items to sync
	var items []struct {
		Id      string
		Key     string
		Value   string
		Comment string
	}
	db = database.Conn()
	db = db.Table("item").Select("id,`key`,value,comment").Where("namespace_id=? AND key IN (?) AND is_delete=0", req.FromNamespaceId, req.Keys).Scan(&items)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
	}
	if len(req.Keys) != len(items) {
		return errors.New("the keys not exist")
	}

	//get items to be sync
	type toItem struct {
		Id          string
		Key         string
		Value       string
		NamespaceId string
	}
	var toItems []toItem
	db = database.Conn()
	db = db.Table("item t1").Select("t1.id,t1.key,t1.value,t2.id namespace_id").
		Joins("JOIN namespace t2 ON t1.namespace_id=t2.id AND t2.name=? AND cluster_id IN (?) AND t2.is_delete=0",
			namespace.Name, req.ToClusterIds).
		Find(&toItems)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
	}
	if namespace.Id == "" {
		return errors.New("the namespace not exists")
	}

	itemMap := map[string]map[string]toItem{}
	for _, item := range toItems {
		if itemMap[item.NamespaceId] == nil {
			itemMap[item.NamespaceId] = map[string]toItem{}
		}
		itemMap[item.NamespaceId][item.Key] = item
	}

	//get max order_num per namespace
	var itemOrderNums []struct {
		NamespaceId string
		MaxOrderNum int
	}
	db = database.Conn()
	db = db.Raw("SELECT namespace_id,MAX(order_num) max_order_num FROM item WHERE namespace_id IN (?) GROUP BY namespace_id").Find(&itemOrderNums)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return errors.Wrap(db.Error, "db error")
	}
	itemOrderNumMap := map[string]int{}
	for _, num := range itemOrderNums {
		itemOrderNumMap[num.NamespaceId] = num.MaxOrderNum
	}

	var updateItems []map[string]interface{}
	var insertItems []map[string]interface{}
	var updateItemIds []string
	var insertItemIds []string
	now := time.Now().Unix()
	for _, item := range items {
		for nsId, nsItem := range itemMap {
			if v, ok := nsItem[item.Key]; ok {
				if v.Value != item.Value {
					updateItems = append(updateItems, map[string]interface{}{
						"id":          v.Id,
						"value":       item.Value,
						"is_delete":   0,
						"update_time": now,
						"update_by":   req.UserId,
					})
					updateItemIds = append(updateItemIds, v.Id)
				}
			} else {
				itemOrderNumMap[nsId]++
				id := uuid.NewV1().String()
				insertItems = append(insertItems, map[string]interface{}{
					"id":           id,
					"namespace_id": nsId,
					"key":          item.Key,
					"value":        item.Value,
					"comment":      item.Comment,
					"order_num":    itemOrderNumMap[nsId],
					"is_delete":    0,
					"create_by":    req.UserId,
					"create_time":  now,
					"update_by":    req.UserId,
					"update_time":  now,
				})
				insertItemIds = append(insertItemIds, id)
			}
		}
	}

	//分别对被修改方进行 增删改
	tx := database.Conn().Begin()
	for _, item := range updateItems {
		tx = database.Update(tx, "item", item, "id=?", item["id"])
	}
	tx = database.InsertMany(tx, "item", insertItems)
	tx = RecordTable(tx, "item", "", req.UserId, com.OpUpdate, updateItemIds...)
	tx = RecordTable(tx, "item", "", req.UserId, com.OpCreate, insertItemIds...)
	if tx.Error != nil {
		tx.Rollback()
		log.Error(tx.Error)
		return errors.Wrap(tx.Error, "db error")
	} else {
		tx.Commit()
	}

	c.recordItem(com.OpUpdate, req.UserId, updateItemIds...)
	c.recordItem(com.OpCreate, req.UserId, insertItemIds...)

	return nil
}

type WatchConfigReq struct {
	Host    string      `json:"host"`
	Port    int         `json:"port"`
	Env     com.EnvType `json:"env"`
	Cluster string      `json:"cluster"`
	App     string      `json:"app"`
}

func (c *WatchConfigReq) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Host, validation.Required),
		validation.Field(&c.Port, validation.Required),
		validation.Field(&c.Cluster, validation.Required),
		validation.Field(&c.App, validation.Required),
		validation.Field(&c.Env, validation.Required),
	)
}

type WatchConfigResp struct {
	InstanceId string                   `json:"instance_id"`
	EventType  com.ConfigWatchEventType `json:"event_type"`
	//Configs    map[com.OpType][]core.ChangeConfig `json:"configs"`
}

func (c *ConfigModel) Watch(req *WatchConfigReq) (*WatchConfigResp, error) {
	resp := &WatchConfigResp{}

	if err := req.Validate(); err != nil {
		log.Error(err)
		return resp, err
	}

	server := core.GetServer()
	if server.Env != req.Env {
		err := errors.Errorf("server env[%s] is not match instance env[%s]", server.Env, req.Env)
		log.Warn(err)
		return resp, err
	}

	now := time.Now().Unix()

	var cluster struct {
		AppId     string
		ClusterId string
	}
	db := database.Conn()
	db = db.Table("cluster t1").Select("t1.id cluster_id, t2.id app_id").
		Joins("JOIN app t2 ON t1.app_id=t2.id AND t2.name=? AND t2.is_delete=0", req.App).
		Where("t1.name=? AND t1.is_delete=0", req.Cluster).Scan(&cluster)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}
	if cluster.ClusterId == "" {
		return resp, errors.New("cluster or app not exists")
	}

	var instance struct {
		Id string
	}
	db = database.Conn()
	db = db.Table("instance").Select("id").
		Where("app_id=? AND cluster_id=? AND host=? AND port=? AND is_delete=0", cluster.AppId, cluster.ClusterId, req.Host, req.Port).
		Scan(&instance)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}
	if instance.Id == "" {
		instance.Id = uuid.NewV1().String()
		db := database.Conn()
		db = database.Insert(db, "instance", map[string]interface{}{
			"id":          instance.Id,
			"app_id":      cluster.AppId,
			"cluster_id":  cluster.ClusterId,
			"host":        req.Host,
			"port":        req.Port,
			"create_time": now,
			"update_time": now,
		})
		if db.Error != nil {
			log.Error(db.Error)
			return resp, errors.Wrap(db.Error, "db error")
		}
	}
	resp.InstanceId = instance.Id

	instances := server.Instances
	ins, ok := instances.Load(instance.Id)
	if !ok {
		ins = &core.Instance{
			Id:       instance.Id,
			AppId:    cluster.AppId,
			Cluster:  cluster.ClusterId,
			Host:     req.Host,
			Port:     req.Port,
			Status:   com.OnlineStatus,
			Life:     60,
			ChChange: make(chan bool, 1),
		}
		instances.Store(instance.Id, ins)
		resp.EventType = com.CwRefreshAll
		return resp, nil
	}

	var isConfigChange bool
	if ins.Status != com.OnlineStatus {
		isConfigChange = true
		resp.EventType = com.CwRefreshAll
	}
	if len(ins.ChChange) > 0 {
		isConfigChange = true
		resp.EventType = com.CwRefreshAll
		<-ins.ChChange
	}
	ins.Status = com.OnlineStatus
	ins.Life = 60
	instances.Store(instance.Id, ins)

	if isConfigChange {
		return resp, nil
	}

	select {
	case <-ins.ChChange:
		resp.EventType = com.CwRefreshAll
	case <-time.After(time.Second * 45):
		resp.EventType = com.CwNothing
	}

	return resp, nil
}

func (c *ConfigModel) notifyChange() {
	server := core.GetServer()
	server.Instances.Range(func(instanceId string, val *core.Instance) bool {
		if len(val.ChChange) == 0 {
			val.ChChange <- true
		}
		return true
	})
}
