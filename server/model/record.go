package model

import (
	"github.com/hackbeex/configcenter/server/database"
	"github.com/hackbeex/configcenter/util/com"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"time"
)

func RecordTable(db *gorm.DB, table, comment, userId string, op com.OpType, ids ...string) *gorm.DB {
	if len(ids) == 0 {
		return db
	}
	now := time.Now().Unix()
	var data []map[string]interface{}
	for _, id := range ids {
		data = append(data, map[string]interface{}{
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
	return database.InsertMany(db, "record", data)
}
