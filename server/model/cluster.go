package model

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/hackbeex/configcenter/server/database"
	"github.com/hackbeex/configcenter/util/com"
	"github.com/hackbeex/configcenter/util/log"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"time"
)

type ClusterModel struct {
}

type CreateClusterReq struct {
	AppId   string `json:"app_id"`
	Name    string `json:"name"`
	Comment string `json:"comment"`
	UserId  string `json:"user_id"`
}

func (c *CreateClusterReq) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.AppId, validation.Required, validation.Length(32, 32)),
		validation.Field(&c.Name, validation.Required, validation.Length(1, 64)),
		validation.Field(&c.Comment, validation.Length(1, 255)),
		validation.Field(&c.UserId, validation.Required, validation.Length(32, 32)),
	)
}

type CreateClusterResp struct {
	Id string `json:"id"`
}

func (a *ClusterModel) Create(req *CreateClusterReq) (*CreateClusterResp, error) {
	resp := &CreateClusterResp{}

	if err := req.Validate(); err != nil {
		log.Warn(err)
		return resp, err
	}

	var app struct {
		Id string
	}
	db := database.Conn()
	db = db.Table("app").Select("id").Where("id=? AND is_delete=0", req.AppId).Scan(&app)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}
	if app.Id == "" {
		return resp, errors.New("the app not exists")
	}

	var existCluster struct {
		Id string
	}
	db = database.Conn()
	db = db.Table("cluster").Select("id").Where("name=? AND is_delete=0", req.Name).Scan(&existCluster)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}
	if existCluster.Id != "" {
		return resp, errors.New("the cluster name exists")
	}

	now := time.Now().Unix()
	id := uuid.NewV1().String()
	tx := database.Conn().Begin()
	tx = database.Insert(tx, "cluster", map[string]interface{}{
		"id":          id,
		"app_id":      req.AppId,
		"name":        req.Name,
		"comment":     req.Comment,
		"create_by":   req.UserId,
		"create_time": now,
		"update_by":   req.UserId,
		"update_time": now,
	})
	tx = RecordTable(tx, "cluster", "", req.UserId, com.OpCreate, id)
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
