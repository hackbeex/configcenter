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

func (a *AppModel) Create(req *CreateAppReq) (string, error) {
	if err := req.Validate(); err != nil {
		log.Warn(err)
		return "", err
	}

	conn := database.Conn()
	var existApp struct {
		Id string
	}
	db := conn.New()
	db = db.Table("app").Select("id").Where("appid=?", req.Appid).Scan(&existApp)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		log.Error(db.Error)
		return "", errors.Wrap(db.Error, "db error")
	}
	if existApp.Id != "" {
		return "", errors.New("the appid exists")
	}

	now := time.Now().Unix()
	id := uuid.NewV1().String()
	db = conn.New()
	db = database.Insert(db, "app", map[string]interface{}{
		"id":          id,
		"appid":       req.Appid,
		"name":        req.Name,
		"comment":     req.Comment,
		"create_by":   req.UserId,
		"create_time": now,
		"update_by":   req.UserId,
		"update_time": now,
	})
	if db.Error != nil {
		log.Error(db.Error)
		return "", errors.Wrap(db.Error, "db error")
	}

	return id, nil
}
