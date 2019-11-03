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

type NamespaceModel struct {
}

type CreateNamespaceReq struct {
	ClusterId string `json:"cluster_id"`
	Name      string `json:"name"`
	Comment   string `json:"comment"`
	UserId    string `json:"user_id"`
}

func (c *CreateNamespaceReq) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.ClusterId, validation.Required, validation.Length(32, 32)),
		validation.Field(&c.Name, validation.Required, validation.Length(1, 64)),
		validation.Field(&c.Comment, validation.Length(1, 255)),
		validation.Field(&c.UserId, validation.Required, validation.Length(32, 32)),
	)
}

type CreateNamespaceResp struct {
	Id string `json:"id"`
}

func (a *NamespaceModel) Create(req *CreateNamespaceReq) (*CreateNamespaceResp, error) {
	resp := &CreateNamespaceResp{}

	if err := req.Validate(); err != nil {
		log.Warn(err)
		return resp, err
	}

	var cluster struct {
		Id    string
		AppId string
	}
	db := database.Conn()
	db = db.Table("cluster").Select("id,app_id").Where("id=? AND is_delete=0", req.ClusterId).Scan(&cluster)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}
	if cluster.Id == "" {
		return resp, errors.New("the cluster not exists")
	}

	var existNamespace struct {
		Id string
	}
	db = database.Conn()
	db = db.Table("namespace").Select("id").Where("name=? AND is_delete=0", req.Name).Scan(&existNamespace)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}
	if existNamespace.Id != "" {
		return resp, errors.New("the namespace name exists")
	}

	now := time.Now().Unix()
	id := uuid.NewV1().String()
	db = database.Conn()
	db = database.Insert(db, "cluster", map[string]interface{}{
		"id":          id,
		"app_id":      cluster.AppId,
		"cluster_id":  req.ClusterId,
		"is_public":   0,
		"name":        req.Name,
		"comment":     req.Comment,
		"create_by":   req.UserId,
		"create_time": now,
		"update_by":   req.UserId,
		"update_time": now,
	})
	if db.Error != nil {
		log.Error(db.Error)
		return resp, errors.Wrap(db.Error, "db error")
	}

	resp.Id = id
	return resp, nil
}
