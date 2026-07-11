package file

import "math"

// List 数据做分页
func List[T any](page int32, pageSize int32, list []T) (total, totalPage int64, res []T) {
	total = int64(len(list))

	totalPage = int64(1)
	if pageSize > 0 {
		totalPage = int64(math.Ceil(float64(total) / float64(pageSize)))
	}

	if page <= 0 {
		page = 1
		pageSize = 100
	} else if pageSize <= 0 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100
	}

	start := int((page - 1) * pageSize)
	end := start + int(pageSize)
	if end > len(list) {
		end = len(list)
	}
	if start > end {
		start = end
	}
	return total, totalPage, list[start:end]
}
