package database

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strings"
)

func InsertMany(db *gorm.DB, table string, data []map[string]interface{}) *gorm.DB {
	if len(data) == 0 || len(data[0]) == 0 {
		return db
	}
	rowLen := len(data)
	colLen := len(data[0])
	//
	ques := make([]string, colLen)
	fields := make([]string, colLen)
	field2s := make([]string, colLen)
	i := 0
	for k := range data[0] {
		ques[i] = "?"
		fields[i] = k
		field2s[i] = "`" + k + "`"
		i++
	}
	mark := "(" + strings.Join(ques, ",") + ")"
	fieldStr := strings.Join(field2s, ",")
	//
	perNum := 50 //每次插入的最大数量
	curNum := 0
	useNum := 0
	for curNum < rowLen {
		if rowLen-curNum > perNum {
			useNum = perNum
		} else {
			useNum = rowLen - curNum
		}
		valueStrings := make([]string, useNum)
		values := make([]interface{}, 0, useNum*colLen)
		for i := range data[curNum : curNum+useNum] {
			valueStrings[i] = mark
			fieldData := data[curNum+i]
			for _, field := range fields {
				values = append(values, fieldData[field])
			}
		}
		sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, fieldStr, strings.Join(valueStrings, ","))
		//fmt.Println("++info:", sql, values)
		db = db.Exec(sql, values...)
		if db.Error != nil {
			return db
		}
		curNum += useNum
	}
	return db
}

func Insert(db *gorm.DB, table string, data map[string]interface{}) *gorm.DB {
	return InsertMany(db, table, []map[string]interface{}{data})
}

func Update(db *gorm.DB, table string, data map[string]interface{}, where string, whereParams ...interface{}) *gorm.DB {
	if len(data) == 0 {
		return db
	}
	ques := make([]string, 0, len(data))
	params := make([]interface{}, 0, len(data)+len(whereParams))
	for k, v := range data {
		ques = append(ques, "`"+k+"`=?")
		params = append(params, v)
	}
	quesStr := strings.Join(ques, ",")

	var sql string
	if where == "" {
		sql = fmt.Sprintf("UPDATE %s SET %s", table, quesStr)
	} else {
		sql = fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, quesStr, where)
		params = append(params, whereParams...)
	}
	return db.Exec(sql, params...)
}
