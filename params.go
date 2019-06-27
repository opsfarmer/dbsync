package dbsync

// 获取增量数据和更新增量数据之间传递的参数
type Params struct {
	Columns     []string        `json:"columns"`      // 列名称
	ColumnTypes []string        `json:"column_types"` // 列的特殊类型，目前只记忆time.Time
	Data        [][]interface{} `json:"data"`         // 待同步的数据，每一行是一条数据，与列名称一一对应
}
