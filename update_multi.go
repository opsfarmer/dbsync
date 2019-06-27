package dbsync

import "database/sql"

// 插入多条数据时的配置信息
type UpdateMultiOptions struct {
	UpdateOptions
	BatchCount int // 一次批量插入的条数，用于加快执行速度
}

// 插入多条增量更新的数据
func DoUpdate(db *sql.DB, tableName string, data [][]interface{}, options UpdateMultiOptions) (err error) {
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
	dataLen := len(data)
	for i := 0; i < (dataLen-1)/options.BatchCount+1; i++ {
		offset := options.BatchCount * i
		var piece [][]interface{}
		if offset+options.BatchCount > dataLen {
			piece = data[offset:]
		} else {
			piece = data[offset : offset+options.BatchCount]
		}
		err = update(tx, tableName, piece, options.UpdateOptions)
		if err != nil {
			panic(err)
		}
	}
	_ = tx.Commit()
	return
}
