package goweb

import "math"

// List 数据做分页
func List[T any](page int32, pageSize int32, list []T) (total, totalPage int64, res []T) {
	total = int64(len(list))
	// 计算总页数
	if pageSize > 0 {
		totalPage = int64(math.Ceil(float64(total) / float64(pageSize)))
	} else {
		totalPage = 1
	}

	if page > 0 {
		if pageSize == 0 {
			pageSize = 10
		}
		if pageSize > 100 {
			pageSize = 100
		}
	} else {
		page = 1
		pageSize = 100
	}

	start := (page - 1) * pageSize // 开始位置
	end := start + pageSize        // 结束位置
	if end > int32(total)-1 {
		end = int32(total)
	}
	if start > int32(total)-1 {
		start = max(int32(total)-1, 0)
	}
	return total, totalPage, list[start:end]
}
