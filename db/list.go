package db

import (
	"math"
)

type ListData struct {
	Data      any   `json:"data"`
	Total     int64 `json:"total"`      // 总数量
	TotalPage int64 `json:"total_page"` // 总页数
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
}

// 查询列表
func (m *Model) List(data *ListData) *Model {
	if m.Error != nil {
		return m
	}

	// 查询总数
	if err := m.Db.Model(m.Model).Count(&data.Total).Error; err != nil {
		m.Error = err
		return m
	}

	// 添加分页
	if data.Page > 0 {
		if data.PageSize == 0 {
			data.PageSize = 10
		}
		offset := (data.Page - 1) * data.PageSize
		m.Db = m.Db.Limit(int(data.PageSize)).Offset(int(offset))
	}

	// 执行查询
	if err := m.Db.Find(&data.Data).Error; err != nil {
		m.Error = err
		return m
	}

	// 计算总页数
	if data.Page > 0 {
		data.TotalPage = int64(math.Ceil(float64(data.Total) / float64(data.PageSize)))
	} else {
		data.TotalPage = 1
	}

	return m
}
