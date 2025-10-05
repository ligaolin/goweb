package db

// 更新参数
type UpdateParam struct {
	ID    int32  `json:"id" validate:"required:主键值必须"`
	Field string `json:"field" validate:"required:字段名必须"`
	Value any    `json:"value"`
}

type DeleteParam struct {
	IDS []int32 `json:"id" validate:"required:主键值必须"`
}

type FirstParam struct {
	ID int32 `form:"id" validate:"required:主键值必须"`
}

type ListParamBase struct {
	Page     int32  `form:"page"`
	PageSize int32  `form:"page_size"`
	Order    string `form:"order"`
}

type ModelID struct {
	ID int32 `gorm:"column:id;primaryKey;type:autoIncrement;comment:ID" json:"id"`
}

type ModelCreatedAt struct {
	CreatedAt Time `gorm:"column:created_at;type:datetime(3);autoCreateTime:milli;comment:创建时间" json:"created_at"`
}

type ModelUpdatedAt struct {
	UpdatedAt Time `gorm:"column:updated_at;type:datetime(3);autoUpdateTime:milli;comment:更新时间" json:"updated_at"`
}

type ModelDeleteAt struct {
	DeleteAt Time `gorm:"column:delete_at;type:datetime(3);comment:删除时间" json:"delete_at"`
}

type ModelSort struct {
	Sort int32 `gorm:"column:sort;type:bigint(11);default:100;comment:排序" json:"sort"`
}

type ModelState struct {
	State string `gorm:"column:state;type:enum('开启','关闭');default:'开启';comment:状态" json:"state"`
}

type ModelHasChildren struct {
	HasChildren bool `gorm:"-:all;default:false" json:"hasChildren"`
}

type ModelChildren[T any] struct {
	PID      int32 `gorm:"column:pid;type:bigint(20);default:0;comment:父级id" json:"pid"`
	Level    int32 `gorm:"column:level;type:tinyint(1);default:1;comment:层级" json:"level"`
	Children []T   `gorm:"-:migration;foreignKey:pid;references:id" json:"children"`
}
