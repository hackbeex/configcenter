package model

import (
	"github.com/hackbeex/configcenter/server/database"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"time"
)

type OpType string

const (
	OpCreate OpType = "create"
	OpUpdate OpType = "update"
	OpDelete OpType = "delete"
)

func RecordTable(db *gorm.DB, table, id, comment, userId string, op OpType) *gorm.DB {
	now := time.Now().Unix()
	return database.Insert(db, "record", map[string]interface{}{
		"id":          uuid.NewV1().String(),
		"table_name":  table,
		"table_id":    id,
		"op_type":     op,
		"comment":     comment,
		"create_by":   userId,
		"create_time": now,
		"update_by":   userId,
		"update_time": now,
	})
}
