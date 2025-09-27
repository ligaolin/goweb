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
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Order    string `form:"order"`
}

// 模型基础字段
type IDCreatedAtUpdatedAt struct {
	ID        int32 `json:"id" gorm:"primarykey;comment:ID"`
	CreatedAt Time  `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt Time  `json:"updated_at" gorm:"comment:更新时间"`
}

// 模型基础字段
type IDCreatedAtUpdatedAtDeletedAt struct {
	IDCreatedAtUpdatedAt
	DeletedAt Time `json:"deleted_at" gorm:"index;comment:删除时间"`
}

// 模型排序
type SortStruct struct {
	Sort int32 `json:"sort" gorm:"type:int(11);default:100;comment:排序"`
}

// 模型状态
type StateStruct struct {
	State string `json:"state" gorm:"type:enum('开启','关闭');default:开启;comment:状态"`
}

// 模型排序和状态
type SortState struct {
	SortStruct
	StateStruct
}

// 模型基础字段
type IDCreatedAtUpdatedAtDeletedAtSortState struct {
	IDCreatedAtUpdatedAtDeletedAt
	SortState
}

// 模型基础字段
type IDCreatedAtUpdatedAtSortState struct {
	IDCreatedAtUpdatedAt
	SortState
}

type HasChildrenStruct struct {
	HasChildren bool `json:"hasChildren" gorm:"-:all;default:false"`
}

type ChildrenStruct[T any] struct {
	Children []T `json:"children" gorm:"-:all;"`
}

type PIDLevel struct {
	PID   int32 `json:"pid" gorm:"column:pid;type:int(11);default:0;comment:父级id"`
	Level int32 `json:"level" gorm:"type:int(2);default:1;comment:层级"`
}
