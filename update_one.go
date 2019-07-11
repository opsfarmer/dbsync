package dbsync

import "github.com/jinzhu/gorm"

// 插入单条数据
func DoUpdateOne(db gorm.SQLCommon, tableName string, data []interface{}, options UpdateOptions) (err error) {
	return update(db, tableName, [][]interface{}{data}, options)
}
