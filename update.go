package dbsync

import "database/sql"

// 插入增量更新的数据
func DoUpdate(db *sql.DB, tableName string, data [][]interface{}, options UpdateOptions) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
			_ = tx.Rollback()
		}
	}()
	// TODO 批量优化
	for _, item := range data {
		err = DoUpdateOne(tx, tableName, item, options)
		if err != nil {
			panic(err)
		}
	}
	_ = tx.Commit()
	return
}
