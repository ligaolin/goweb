package db

type ListData struct {
	Data      any   `json:"data"`
	Total     int32 `json:"total"`      // 总数量
	TotalPage int32 `json:"total_page"` // 总页数
	Page      int32 `json:"page"`
	PageSize  int32 `json:"page_size"`
}

// 查询列表
func (m *Model[T]) List(data *ListData) *Model[T] {
	if m.Error != nil {
		return m
	}

	// 查询总数
	var total int64
	if err := m.Db.Model(m.Model).Count(&total).Error; err != nil {
		m.Error = err
		return m
	}
	data.Total = int32(total)

	// 添加分页
	if data.Page > 0 {
		if data.PageSize <= 0 {
			data.PageSize = 10
		}
		m.Db = m.Db.Offset(int((data.Page - 1) * data.PageSize))
	}
	if data.PageSize > 100 {
		data.PageSize = 100
	}
	if data.PageSize > 0 {
		m.Db = m.Db.Limit(int(data.PageSize))
	}

	// 执行查询
	if err := m.Db.Find(&data.Data).Error; err != nil {
		m.Error = err
		return m
	}

	// 计算总页数
	if data.Page > 0 {
		data.TotalPage = (data.Total + data.PageSize - 1) / data.PageSize
	} else {
		data.TotalPage = 1
	}

	return m
}
